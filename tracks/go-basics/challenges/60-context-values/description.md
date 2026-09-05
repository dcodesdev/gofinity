# Context Values

A context carries three things. Two of them - cancellation and deadlines - are
what it is for. The third,
[`context.WithValue`](https://pkg.go.dev/context#WithValue), is the one that
gets abused, so this challenge builds it the way it is meant to be built and
shows what goes wrong otherwise.

## What it is for

Request-scoped data that crosses API boundaries and that the code in between
does not care about: a request id for correlating logs, the authenticated user,
a trace span, a locale.

The test is "does the middle of the stack need to know this exists?". A request
id passing through six layers of code that never read it: a value. A database
handle, a logger, a configuration, a retry count: **not** a value. Those are
parameters or struct fields, and passing them through a context turns a
compile-time contract into a runtime lookup that returns `nil` when somebody
forgets.

## The key must be a type nobody else can name

```go
type contextKey int

const (
	requestIDKey contextKey = iota
	userKey
)
```

`WithValue` stores under an `any` and compares keys with `==`. Two packages
that both pick the string `"user"` silently overwrite each other, and the one
that loses has no way to find out. An **unexported type** makes that
impossible: no other package can construct a `contextKey`, so no other package
can collide with your keys or read your values by guessing.

[`go vet`](https://pkg.go.dev/cmd/vet)'s cousins flag a built-in type used as a
context key for exactly this reason. Do not use a string, an int, or a struct
anyone else can build.

## Wrap the lookup

Never make callers write `ctx.Value(myKey).(string)`. Export a pair of
functions and keep the key private:

```go
func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func UserFrom(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey).(User)
	return u, ok
}
```

The **two-value type assertion** matters. A missing key gives a nil `any`, and
the one-value form would panic on it. The `bool` also lets a caller tell "no
user" from "the zero user", and returning the zero value alongside `false` is
what keeps a caller who ignores the second result safe: `User{}` is not an
admin.

## Contexts are immutable

`WithValue` never modifies anything. It returns a **new** context that wraps
the old one, and a lookup walks up the chain until it finds the key. So:

- Deriving a second value does not disturb the first.
- Setting the same key again shadows it for the child while the parent still
  sees the old value.
- A `WithCancel` or `WithTimeout` child still sees every value its ancestors
  carry: cancellation and values live on the same chain.

That chain is also why values are not free. Each lookup is a walk, and a
context with twenty values is a linked list of twenty. It is a handful of
nanoseconds, not a reason to worry, but it is a reason not to store your whole
program in there.

## Middleware

Context values exist because of this shape. Each layer wraps the next and may
add to the context before calling it:

```go
type Handler func(ctx context.Context) error
type Middleware func(next Handler) Handler
```

Building the chain from the last middleware backwards leaves the first one on
the outside, so the order the call site reads is the order things run. A
middleware that rejects a request returns its error and **does not call
`next`** - that is what makes it a guard rather than a suggestion.

## Task

Implement the keys, `WithRequestID`, `RequestID`, `WithUser`, `UserFrom`,
`Describe`, `Chain` and `RequireAdmin`.

## Hints

- One unexported key type and two constants of it. The tests check that a plain
  string key can neither read your values nor overwrite them.
- `RequestID` stores what it is given, including `""`. Present-and-empty is not
  the same as absent.
- `Describe` reads through your own accessors rather than `ctx.Value`.
- `Chain` loops over `mw` **backwards**, wrapping as it goes.
- `RequireAdmin` returns a `Handler` that closes over `next`, and returns
  `ErrForbidden` before calling it when the user is missing or not an admin.
