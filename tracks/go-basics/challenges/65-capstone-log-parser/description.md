# Capstone: Log Parser

The second capstone, and the one that looks most like work. A log file arrives
as lines of text; you have to turn it into records, survive the lines that are
not records, aggregate what is left, and do it on more than one core without
the answer changing.

That is structs, interfaces, errors, `sync` and goroutines in one program,
which is exactly the point: those five things were separate lessons, and
nothing real ever uses them separately.

## The format

```
LEVEL|service|duration_ms|message
INFO|auth|12|login ok
ERROR|auth|145|token expired for user 91
WARN|search|310|slow query: SELECT | FROM docs
```

Four fields, pipe separated, every field trimmed. The message is the *rest* of
the line, pipes and all, which is why it is
[`strings.SplitN`](https://pkg.go.dev/strings#SplitN)`(raw, "|", 4)` and not
[`strings.Split`](https://pkg.go.dev/strings#Split). `Split` on the third line
above gives you six fields and a truncated message; `SplitN` with a limit of 4
stops cutting once it has the first three.

## A bad line is data, not a crash

Real logs contain rotated fragments, a truncated last line, a level someone
invented on a Friday. One bad line must not stop the other ten thousand, so
parsing returns an error rather than panicking, and the caller collects the
errors and carries on.

The error has to say *which* line and *what* was wrong, and both of those want
to be machine readable:

```go
type ParseError struct {
	Line int
	Raw  string
	Err  error
}

func (e *ParseError) Unwrap() error { return e.Err }
```

`Err` is one of three sentinels. `Unwrap` is what lets `errors.Is(err,
ErrBadLevel)` see through the wrapper, and the struct is what lets
`errors.As(err, &pe)` get the line number back out. Those two are the pair:
[`errors.Is`](https://pkg.go.dev/errors#Is) for "what kind of failure",
[`errors.As`](https://pkg.go.dev/errors#As) for "give me the details".

Note the pointer receiver. `Error`, `Unwrap` and `errors.As(&pe)` all agree
that the error *value* is `*ParseError`, so return `&ParseError{...}`, never
`ParseError{...}`.

## The sink is an interface so the parser can forget about counting

```go
type Sink interface {
	Add(e Entry)
}
```

`ParseAll` does not know what happens to an entry. It parses, it calls `Add`,
it is finished. That is what lets the tests hand it a recorder instead of the
real aggregator, and it is what would let you add a second aggregator later
without touching the parser.

One line of that interface's documentation is load-bearing: **`Add` is called
from several goroutines at once**. There is no way to say that in a signature,
so it is said in the comment, and every implementation owes it a mutex.

## Concurrency you can check

`ParseAll` takes a worker count and spreads the lines across that many
goroutines. The standard shape:

```go
jobs := make(chan job)
var wg sync.WaitGroup
for range workers {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := range jobs {  // ends when jobs is closed
			...
		}
	}()
}
for i, raw := range lines { jobs <- job{i + 1, raw} }
close(jobs)
wg.Wait()
```

The workers share one channel and take the next line whenever they are free,
so a slow line does not hold up a worker with nothing to do. `close` is what
ends the `range`, and `wg.Wait` is what makes "ParseAll returned" mean "every
line was handled".

## Determinism is the hard part

Concurrency reorders things, and a report that reorders itself is a report
nobody can diff:

- Errors come back **sorted by line number**, not in the order the goroutines
  happened to finish. The number is inside the `*ParseError`, so sorting means
  `errors.As` first.
- Services come out sorted by total time, and ties break on the name.
- The slowest entry for a service ties on the *smaller message*, not on
  whichever goroutine won the race.

Every one of those is a rule that exists only because the work is concurrent.
Write them down and the output is a function of the input again.

## Snapshot hands out copies

`Snapshot` builds a `Summary` while holding the lock, and it copies: a fresh
map, a fresh slice, the `ServiceStat` values dereferenced out of whatever the
`Stats` keeps them in. Return the internal map and the caller is reading shared
state without a lock, which is a data race with no code in sight to blame.

## Task

Implement `ParseError`'s two methods, `ParseLine`, `Stats` with `NewStats`,
`Add` and `Snapshot`, then `ParseAll` and `Summarize`, to the contracts in
their doc comments.

`Stats` has no fields yet. Choose them: the counters you need, and the lock
that makes `Add` safe to call from four goroutines at once.

## Hints

- `strings.SplitN(raw, "|", 4)` and check `len(fields) < 4` before you index
  anything.
- A tiny `fail := func(err error) (Entry, error)` closure at the top of
  `ParseLine` keeps the four failure paths to one line each.
- [`strconv.Atoi`](https://pkg.go.dev/strconv#Atoi) rejects `"fast"` and `""`;
  you reject a negative yourself.
- Keep services as a `map[string]*ServiceStat` and you can update a service in
  place rather than reading, editing and writing back a value.
- `Add` is two lines of bookkeeping between `s.mu.Lock()` and a
  `defer s.mu.Unlock()`. Snapshot takes the same lock: reading an `int` while
  another goroutine writes it is a race even when the answer looks right.
- Collect the errors into a shared slice under a second, separate mutex, then
  sort once at the end, after `wg.Wait`.
- `for range workers` is Go 1.22's range-over-int. `for i := 0; i < workers;
  i++` is the same thing.
