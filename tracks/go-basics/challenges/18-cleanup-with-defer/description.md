# Cleanup with defer

The reason `defer` exists is that a function has many exits and only one of them
is the last line. An early `return`, a `return` from inside a loop, a panic
three calls deeper: all of them have to release what the function acquired, and
writing the release once, next to the acquire, is the only version that stays
correct as the function grows.

```go
f := open(path)
defer f.Close()
```

Two lines together, and the second one covers every path out. That pairing is
the whole idiom, and it is why Go code puts the cleanup at the *top* of a
function rather than the bottom.

Three things follow.

**Cleanup runs during a panic.** Deferred calls execute as the stack unwinds, so
a resource is released even when the code that used it blew up.

**Nested acquires release in reverse.** Deferred calls are a stack, and reverse
order is what you want: the inner thing goes away before the outer thing it
depends on.

**`defer` is scoped to the function, not to the block.** A `defer` inside a
`for` loop does not run at the end of the iteration; it piles up and runs when
the whole function returns. If you need per-iteration cleanup, put the body in
its own function.

## Task

Fill in the four functions in `main.go`. Each takes `acquire` and `release` as
function values, and the tests pass in a pair that records what happened.

1. `WithResource(name, acquire, release, body)` acquires, runs `body`, releases
   on every path, and lets a panic in `body` continue to the caller.
2. `EachResource(names, acquire, release, body)` handles a whole slice, holding
   **one resource at a time**.
3. `Nested(acquire, release)` acquires `outer` then `inner` and releases them in
   reverse.
4. `TryUse(name, acquire, release, body) error` always releases, and turns a
   panic in `body` into an error `"use <name>: <panic value>"`.

## Hints

- `defer release(name)` on the line after `acquire(name)`. Anything between the
  two is a leak waiting for its first early return.
- `EachResource` is the loop trap: deferring inside the loop holds everything
  until the function ends, and the test checks the interleaving. One helper call
  per iteration fixes it, and you already wrote the helper.
- `Nested` needs no ordering logic. Defer both releases as you acquire and the
  stack reverses them for you.
- `TryUse` combines this lesson with the last one: a named result, a deferred
  `recover`, and the release deferred as well so it happens either way.
- The panic value in `"use file: boom"` is formatted with `%v`, so
  [`fmt.Errorf("use %s: %v", name, r)`](https://pkg.go.dev/fmt#Errorf) is the
  whole line.
