# Context

Every concurrent program in the last three lessons had the same hole in it. You
could start work, and you could wait for work, but you had no way to say
**stop**.

Real programs need that constantly. A user closes the tab, so the search
running on their behalf is pointless. A request has been going for thirty
seconds and is not going to become useful at thirty-one. The server is being
redeployed and wants ten seconds to finish what is in flight. One of five
parallel calls failed, so the other four are wasted work.

`context.Context` is Go's answer, and it is a small one: a single value, passed
down through every call in a job, carrying one broadcast - *this work is no
longer wanted* - and the deadline and request-scoped data that travel with it.

## The tree

You never construct a context directly. You start from a root and **derive**
children:

```go
ctx := context.Background()

ctx, cancel := context.WithCancel(ctx)
defer cancel()

ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
defer cancel()

ctx = context.WithValue(ctx, userKey, u)
```

`context.Background()` is the empty root: no deadline, no values, never
cancelled. It belongs in `main`, in tests, and at the top of a request. Every
other context in your program is a descendant of one.

Derivation builds a tree, and cancellation flows **down** it. Cancelling a
parent cancels every descendant; cancelling a child leaves the parent and its
siblings alone. That is what makes a context safe to pass into a helper: the
helper can tighten its own copy and can never loosen yours.

Contexts are immutable. Nothing you do to a child is visible to its parent.

## The three questions

A `Context` has four methods, and they answer three questions.

**Should I stop?** `ctx.Done()` returns a channel that is **closed** when the
work should stop. Closed, not sent to: a closed channel is receivable for ever,
by every goroutine at once, which is exactly what a broadcast needs. One
`Done()` channel serves a thousand watchers with no bookkeeping.

**Why did it stop?** `ctx.Err()` is `nil` while the context is live and, once
`Done` is closed, exactly one of two values: `context.Canceled` or
`context.DeadlineExceeded`.

**How long do I have?** `ctx.Deadline()` gives the moment, and whether there is
one at all.

The fourth, `ctx.Value(key)`, is a different feature living in the same value,
and it gets its own section below.

## Listening

Cancellation is a channel, so everything you know about `select` applies:

```go
select {
case v := <-work:
	return v, nil
case <-ctx.Done():
	return 0, ctx.Err()
}
```

Return `ctx.Err()` rather than an error of your own. Several frames up, someone
will ask `errors.Is(err, context.Canceled)` to tell "the user went away" from
"the database is broken", and `errors.New("cancelled")` throws that away.

**Sends need the same treatment as receives.** `out <- v` blocks until someone
receives; if the consumer has given up, that goroutine is parked for the
lifetime of the process. Any blocking operation in a cancellable job belongs in
a select against `ctx.Done()`.

For a loop that does not block, check without waiting:

```go
for _, item := range items {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	process(item)
}
```

A `select` with a `default` never blocks. Put that at the top of each turn of a
long loop and a cancelled job stops at the next item rather than at the end.

## Deadlines

A cancellation you have to remember to trigger is one you will forget. Most
work should give up on its own:

```go
ctx, cancel := context.WithTimeout(parent, 2*time.Second)
defer cancel()
```

`WithTimeout` and `WithDeadline` are the same tool: a duration from now, or a
moment. Use whichever the caller already has.

The property that makes this safe is that **a child is never laxer than its
parent**:

```go
parent, _ := context.WithTimeout(ctx, time.Second)
child, _ := context.WithTimeout(parent, time.Hour)   // still one second
```

A helper that "gives itself an hour" cannot outlive the request it was called
for. There is no API to extend a deadline. Work that genuinely must outlive its
request starts from `context.Background()` again, deliberately and rarely.

`defer cancel()` is not optional even when a timeout will fire anyway. It
releases the timer and unhooks the child from its parent; skipping it leaks
both until the parent finishes, which for a long-lived parent means for ever.
`go vet` reports it as "the cancel function is not used", and it is right every
time.

## Racing work you cannot interrupt

Plenty of functions know nothing about contexts: an old library, a CPU-bound
loop, a blocking syscall. You cannot stop them, but you can stop *waiting*:

```go
done := make(chan result, 1)   // buffered, and this is why
go func() { done <- do() }()

select {
case r := <-done:
	return r.value, r.err
case <-ctx.Done():
	return 0, ctx.Err()
}
```

The buffer is the whole trick. When the timeout wins, nobody will ever receive
from `done`; on an unbuffered channel that goroutine would be parked on its
send permanently. One slot lets it deliver a result nobody wants and exit.

Be honest about what this buys: `select` bounds how long you *wait*, not how
long the work *takes*. The goroutine runs to completion in the background. If
it holds a connection or a lock, teach it about the context instead of racing
it.

## Values

`context.WithValue` is the third feature and the one that gets abused.

It is for **request-scoped data that crosses API boundaries and that the code
in between does not care about**: a request id for correlating logs, the
authenticated user, a trace span, a locale.

The test is whether the middle of the stack needs to know it exists. A request
id threading through six layers that never read it is a value. A database
handle, a logger, a config, a retry count are **not**: they are parameters or
struct fields, and hiding them in a context turns a compile-time contract into
a runtime lookup that returns `nil` when someone forgets.

The key must be a type nobody else can name:

```go
type contextKey int

const (
	requestIDKey contextKey = iota
	userKey
)

func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func UserFrom(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey).(User)
	return u, ok
}
```

`WithValue` stores under an `any` and compares keys with `==`. Two packages
that both choose the string `"user"` overwrite each other silently, and the
loser cannot find out. An unexported type makes that impossible.

Wrap the lookup, always. Callers should never write
`ctx.Value(k).(User)` themselves, and the **two-value** assertion is required:
a missing key is a nil `any`, and the one-value form panics on it. The `bool`
also lets a caller tell "no user" from "the zero user", and returning `User{}`
alongside `false` keeps a caller who ignores the second result safe, because
the zero user is not an admin.

Lookup walks up the chain, so a `WithCancel` child still sees everything its
ancestors carry, and setting a key again shadows it for the child only. It also
means each lookup is a walk: nanoseconds, not a worry, but not a place to store
your program either.

## Where it goes in a signature

- `ctx` is the **first** parameter, named `ctx`, typed `context.Context`.
- Do not store a context in a struct. The exception exists - a struct that *is*
  a request - and you will know when you have one.
- Never pass a `nil` context. `context.TODO()` is the honest placeholder for
  "there will be a real one here later"; it behaves exactly like
  `Background()`, and it reads as a note to yourself.
- A function that takes a `ctx` should respect it. If it cannot be cancelled,
  do not take one.

Everything in the standard library that can block takes one: `http.Request` has
`Context()` and `NewRequestWithContext`, `database/sql` has `QueryContext` and
`ExecContext`, `os/exec` has `CommandContext`. When you write your own blocking
function, follow the same shape.

## The checklist

- `context.Background()` at the top, derive everything else from it.
- `ctx.Done()` is closed, not sent to - one channel broadcasts to everyone.
- `ctx.Err()` is `context.Canceled` or `context.DeadlineExceeded`; return it,
  wrap it with `%w`, compare with `errors.Is`.
- `defer cancel()` every time, including with a timeout.
- A child can only be stricter than its parent; there is no extending.
- Selects guard sends as well as receives.
- `select` with a `default` checks without blocking.
- Racing uninterruptible work needs a buffer of one, or the goroutine leaks.
- Values: request-scoped data only, unexported key type, wrapped accessors,
  two-value assertion.
- `ctx` first, named `ctx`, not in a struct, never nil.

## Further reading

- [context](https://pkg.go.dev/context) - `Background`, `WithCancel`,
  `WithTimeout`, `WithValue`, and the rules each constructor documents.
- [Go Concurrency Patterns: Context](https://go.dev/blog/context) - the
  original post, with a worked example that threads `ctx` through a server.
- [Contexts and structs](https://go.dev/blog/context-and-structs) - why `ctx`
  is the first parameter and not a field.
- [net/http](https://pkg.go.dev/net/http) - where a server's context comes
  from, and what cancels it.
- [errors](https://pkg.go.dev/errors) - `Is`, for comparing against
  `context.Canceled` and `context.DeadlineExceeded` through a wrap.

## Practise

Three challenges. The first is cancellation itself: waiting on `Done` next to
real work, the non-blocking check, a receiver that tells a closed channel from
a cancelled context, and a producer that guards its sends and closes its
channel on both paths. The second is deadlines: reading the remaining budget, a
child that cannot outlive its parent, racing a function that knows nothing
about contexts, and a retry loop whose error carries both the last failure and
the context's own. The third is values: an unexported key type a plain string
cannot reach, accessors that tell absent from empty, and a middleware chain
where a guard that rejects a request never calls what it wraps.
