# Defer Order

`defer` schedules a call to run when the surrounding function returns, no matter
which `return` it takes or whether it panics on the way out:

```go
func read() {
	f := open("data.txt")
	defer close(f)
	// ... close(f) runs here, whatever happens above
}
```

Two rules explain everything `defer` does.

**Deferred calls run last-in, first-out.** They go on a stack, so the last one
you defer is the first one to run. That is what you want: you release things in
the reverse of the order you acquired them.

**The arguments are evaluated at the `defer` statement, not when the call
runs.** `defer fmt.Println(n)` snapshots `n` immediately. A deferred *closure*
that reads `n` in its body is the opposite: it sees whatever `n` holds at
return time.

Because deferred functions run after the return value has been set, a closure
over a **named result** can still change it, and the caller sees the change.

## Task

Fill in the five functions in `main.go`.

1. `LIFO(labels []string) (order []string)` defers one appending call per label
   and returns the order they ran in, so `a, b, c` comes back as `c, b, a`.
2. `Steps() (log []string)` logs `enter`, defers `cleanup-1` then `cleanup-2`,
   logs `work`, and returns `[enter work cleanup-2 cleanup-1]`.
3. `CapturedValue() (recorded int)` passes `n` to the deferred function as an
   argument before changing `n`, and returns what was recorded.
4. `CapturedVariable() (recorded int)` is the same, except the deferred function
   closes over `n` instead.
5. `DoubleResult(n int) (result int)` returns `n` doubled, with the doubling
   done by a deferred function.

## Hints

- These functions all have **named results**. A bare `return` sets them to what
  they currently hold and then runs the defers, which can still change them.
- `defer f()` calls `f` now and defers whatever it returned. To defer a block of
  work, defer a function literal: `defer func() { ... }()`, trailing `()`
  included.
- `LIFO` needs no reversing logic and no counter. Defer inside the loop and let
  the stack do it.
- The difference between 3 and 4 is one line of syntax: `func(seen int) { ... }(n)`
  freezes `n`, `func() { ... }()` does not.
- `DoubleResult` cannot double before the `return`; the test would still pass,
  but the point is that the deferred call edits `result` after it is set.
