# Capstone

Every lesson so far took one idea and held it still. Maps, then errors, then
goroutines. That is how you learn a language, and it is not how you use one: a
two hundred line program touches a dozen of those ideas in the first
paragraph, and the difficulty moves from "what does `append` do" to "what
shape should this be".

This lesson is about the shape. No new syntax, nothing you have not already
seen, just the handful of decisions that turn a pile of working functions into
a program somebody else can read.

## Programs are pipelines of values

Start by writing down the stages and the type between each pair of them:

```
text  ->  Tokenize  ->  []string
      ->  Tally     ->  map[string]int
      ->  Rank      ->  []Count
      ->  Report     -> string
```

Four functions, four types, and nothing in the middle knows where the text came
from or where the report is going. That layout is worth the small amount of
extra typing for three reasons:

- Every stage is testable with a literal. `Rank` takes a map you can write out
  by hand, so testing the ranking needs no text and no file.
- Every stage is nameable. `Tally` says what it does; the eighty line function
  it came out of said "process".
- A bug lands in one stage. When the report is wrong you look at the value
  between the stages and the wrong one tells you where to go.

Then a top-level function - `Analyze`, `Summarize`, `Run` - is the four of them
in a row, and it is the only function that knows the whole story. In a larger
program that function is also the only one that touches the outside world:
reads the file, writes to stdout, opens the connection. Keep IO at the edge and
the middle stays pure, which is the same reason it stays testable.

## Decide what your data is before you write the code

The first field of a struct is a design decision and the rest follow from it.
`Entry{Level, Service, Millis, Message}` is a choice: the duration is an `int`
of milliseconds and not a `time.Duration`, the level is a `string` and not a
custom `type Level int`, the message is not parsed further. Every one of those
is arguable, and writing the struct down is what makes them arguable rather
than accidental.

Two habits that pay for themselves:

- **Name the units.** `Millis`, not `Duration`; `Bytes`, not `Size`. The name is
  where the unit lives when the type is `int`.
- **Give the aggregate its own type.** `Summary`, `ServiceStat`, `Count`. A
  function returning `(map[string]int, []string, int, error)` has a struct
  hiding in it, and the struct is the version you can add a field to.

## Errors carry what the caller will act on

By now you know `errors.Is`, `errors.As` and `%w`. A capstone is where they
stop being separate techniques and become one policy, which is worth stating
plainly:

- The **sentinel** says what kind of failure it is: `ErrNoWords`,
  `ErrBadLevel`. Callers switch on it with `errors.Is`.
- The **struct** carries the details: which line, which file, what it was
  reading. Callers dig them out with `errors.As`.
- The **wrapping** adds context on the way up: `fmt.Errorf("analyze: %w", err)`.
  It is one clause per layer, and the clause names the operation, not the
  failure - the failure is already in there.

```go
type ParseError struct {
	Line int
	Raw  string
	Err  error // one of the sentinels
}

func (e *ParseError) Error() string { return fmt.Sprintf("line %d: %v: %q", e.Line, e.Err, e.Raw) }
func (e *ParseError) Unwrap() error { return e.Err }
```

That struct is both at once: `errors.Is(err, ErrBadLevel)` finds the sentinel
through `Unwrap`, and `errors.As(err, &pe)` gets the line number. Neither
caller ever reads the message text, which is the whole point - a message is for
a human, and code that parses one breaks the moment you improve the wording.

The other half of the policy is **who stops**. A bad line in a log file is
data: parsing it fails, the file keeps going, and the failures come back as a
slice. A missing input file is not: nothing downstream can proceed, so the
error goes up and the program exits. Deciding which of the two you are looking
at is a design question, and it is usually the difference between a tool people
run on real data and one that dies on line 12 of a rotated log.

## An interface appears where the caller stops caring

Do not start with interfaces. Start with the concrete type, and when a second
implementation shows up - even if the second one is a test double - look at
what the caller actually used:

```go
type Sink interface {
	Add(e Entry)
}
```

The parser hands finished records to something. It does not aggregate, does not
know what a `Summary` is, and cannot be broken by a change to the counting. The
test passes it a recorder, the program passes it the real statistics, and both
of them are one method.

That is the Go rule you have met a dozen times: the interface is declared by
the consumer, it holds the methods the consumer calls, and it is small enough
to be satisfied by accident.

There is one thing an interface cannot say, and this is where it matters:
`Add` will be called from several goroutines at once. No signature expresses
that, so the documentation does, and every implementation owes it a mutex.
When you write an interface that will be used concurrently, say so in the
comment above it, because the implementer will not guess.

## Concurrency is a decision about the seams

The parser could parse each line as it reads it. Instead the lines go into a
channel, some workers take them, and the results go to a sink:

```go
jobs := make(chan job)
var wg sync.WaitGroup
for range workers {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := range jobs {
			...
		}
	}()
}
for i, raw := range lines { jobs <- job{i + 1, raw} }
close(jobs)
wg.Wait()
```

You have seen every line of that. What is new is knowing where to put it. Two
questions decide:

1. **Is the work independent?** Parsing line 40 does not need line 39. If the
   stages need each other's results in order, concurrency buys nothing but
   bugs.
2. **Is the work worth a goroutine?** A goroutine is cheap, not free. Spreading
   ten string splits across four workers is slower than doing them.

And when the answer is yes, the concurrency lives at one seam - here, "parse
the lines" - rather than being sprinkled through the program. The stages either
side of it stay sequential and stay simple.

## The output must not depend on the schedule

This is the rule people learn last and the one that makes concurrent programs
usable. Goroutines finish in whatever order they finish in, so anything the
program *prints* has to be put back in a defined order before it is printed:

- Errors come back sorted by line number, not by whoever failed first.
- Ranked output sorts by the count and then by the name, so the ties do not
  shuffle.
- "The slowest entry" ties on the smaller message, not on the winning
  goroutine.

Each of those is one comparison in a `sort.Slice`, and together they are the
difference between output you can diff between two runs and output you cannot
test at all. Ranging over a map gives the same problem without any goroutines
at all: Go randomises the order deliberately, so every report built from a map
needs a sort before it is printed.

While you are there, hand out copies. A `Snapshot` method that returns the
internal map has published shared state, and the next reader of it races with
the next writer:

```go
func (s *Stats) Snapshot() Summary {
	s.mu.Lock()
	defer s.mu.Unlock()
	// build a fresh map and a fresh slice, copy the values in, return them
}
```

## The checklist

When you sit down in front of a blank `main.go`:

1. **What are the values?** Write the structs. Name the units.
2. **What are the stages?** Write the function signatures, with a type between
   each pair. `main` is glue.
3. **What can go wrong, and who stops?** Sentinels for the kinds, a struct for
   the details, `%w` on the way up. Bad record or bad run.
4. **Where is the seam?** If two things want to vary independently, one small
   interface between them, declared by the caller.
5. **Is anything independent enough to be concurrent?** If so, one seam, a
   `WaitGroup`, and a sort before you print.
6. **Write the test as a table**, with the boring cases first: empty input, one
   item, a tie, a duplicate, the wrong type of line.

None of that is Go specific, and all of it is easier in Go than in most
languages, because the type system is small enough that the shape of a program
is visible in its signatures.

## What comes after this

You have the language. What is left is the standard library and the habits, and
both are learned by writing things:

- `net/http` for servers and clients, `encoding/json` you already met, `io` and
  `bufio` for streaming rather than slurping a whole file.
- `flag` or a CLI library, so the program takes its input from the command line
  rather than a constant.
- `go doc`, `go vet`, `-race` and `-cover` on everything you write.
- The standard library's own source, which is unusually readable and is where
  most of the idioms in this track came from.

Two capstones follow this lesson. They are longer than anything else in the
track, they have no new syntax in them, and they are graded on the parts that
are easy to get almost right: the tie-break, the wrapped sentinel, the copy,
the sort. Take them slowly.

## Further reading

- [Effective Go](https://go.dev/doc/effective_go) - the one document to read
  end to end now that the syntax is no longer in the way.
- [The Go Programming Language Specification](https://go.dev/ref/spec) - short,
  precise, and the answer to every "does it really do that" question.
- [Standard library](https://pkg.go.dev/std) - the index to browse when you are
  about to write something the library already has.
- [How to Write Go Code](https://go.dev/doc/code) - modules, packages, tests
  and commands, for the first program you start from nothing.
- [Go FAQ](https://go.dev/doc/faq) - why the language is shaped the way it is,
  which is most of what is left to learn.
