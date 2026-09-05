# Subtests and Helpers

A table-driven test is one loop over rows. That is already good, and it has
three gaps: a failure in row four cannot use `Fatalf` without abandoning rows
five and six, you cannot run one row on its own, and the shared setup either
runs once for every row or is copied into each of them.

[`t.Run`](https://pkg.go.dev/testing#T.Run) closes all three.

```go
for _, c := range cases {
	t.Run(c.name, func(t *testing.T) {
		if got := Double(c.in); got != c.want {
			t.Errorf("got %d, want %d", got, c.want)
		}
	})
}
```

Each row is now a test of its own, with its own `*testing.T`:

- Its name is `TestDouble/below_the_range`, and it appears in the output.
- `go test -run 'TestDouble/below'` runs that row and nothing else.
- A `Fatalf` inside it ends **that row**, not the loop.
- `t.Run` returns a `bool`, so you can skip work that only makes sense if the
  row passed.

Note the shadowing: the inner `t` is not the outer `t`. Using the outer one
inside the closure by accident attributes every failure to the parent, and is
the most common bug in this shape.

Spaces in a subtest name become underscores, so `-run` arguments stay
shell-friendly. Duplicate names get `#01` appended rather than being merged.

## Helpers and t.Helper

The second thing a test file grows is helpers: `assertEqual`, `mustOpen`,
`newFixture`. Without one line they are actively unhelpful:

```go
func assertEqual(t *testing.T, got, want int) {
	if got != want {
		t.Errorf("got %d, want %d", got, want)  // reported here, always here
	}
}
```

Every failure in the file now points at that one line.
[`t.Helper()`](https://pkg.go.dev/testing#T.Helper) at the top of the function
fixes it: the helper is skipped when the failure's line is chosen, so the
reported line is the call site. Put it at the top, next to the signature, not
inside the `if`.

## Errorf or Fatalf

[`Errorf`](https://pkg.go.dev/testing#T.Errorf) marks the test failed and
returns; [`Fatalf`](https://pkg.go.dev/testing#T.Fatalf) marks it failed and
ends the test goroutine with
[`runtime.Goexit`](https://pkg.go.dev/runtime#Goexit). It does not return, so
nothing after it runs.

The rule is about what comes next. If the following lines can still say
something useful, use `Errorf` and let them. If they would only panic on a nil
value you already know is nil, use `Fatalf`.

That has a consequence for helpers: a helper that calls `Fatalf` must be called
from the test goroutine itself. `Fatalf` from a goroutine you started stops
*that* goroutine, and the test carries on believing it passed.

## Cleanup, not defer

`defer` in a test works. `defer` in a *helper* does not: it runs when the
helper returns, which is before the test has done anything.

[`t.Cleanup(fn)`](https://pkg.go.dev/testing#T.Cleanup) registers work for when
the **test** finishes, and it runs on success, on failure, and after `Fatalf`
alike. That is what lets a helper own both halves of a resource:

```go
func newFixture(t *testing.T) *Fixture {
	t.Helper()
	f := open()
	t.Cleanup(f.Close)
	return f
}
```

Cleanups run last registered first, like `defer`, so nested fixtures unwind in
the right order. A subtest's cleanups run when that subtest finishes, not at
the end of the parent.

## Why the interfaces

Helpers here take `TB`, an interface with the five methods they use, rather
than `*testing.T`. The standard library does the same with
[`testing.TB`](https://pkg.go.dev/testing#TB), the interface `*testing.T` and
`*testing.B` share, so one helper serves tests and benchmarks.

`testing.TB` has an unexported method, deliberately: you cannot implement it,
so nobody can pass a homemade `T` to a helper. When you want a fake - to test
the helper itself, which is what this challenge's own tests do - you declare the
methods you use and take that.

`Suite` adds `Run`. `*testing.T` cannot satisfy it, because its `Run` takes a
`func(*testing.T)` and not a `func(Suite)`. That is what `GoSuite` is for: an
adapter, one method long, that wraps the subtest's `*testing.T` back up as a
`Suite` on the way in.

## Task

Implement `Equal`, `Must`, `WithResource`, `RunTable` and `GoSuite.Run`.

## Hints

- Every helper starts with `t.Helper()`. `RunTable` included.
- `Equal` compares with `!=`, which is why `V` is `comparable`. Print with
  `%v`: it works for whatever the caller passed.
- `Must` returns `v` after the `if`. There is no `else`, because `Fatalf` does
  not come back.
- `WithResource` calls `open` first, then registers a closure that captures the
  result. Registering `closeFn` itself will not compile - `Cleanup` takes a
  `func()`.
- `RunTable`'s subtest body must use the `Suite` it is given, not the one from
  the enclosing scope. That is the shadowing trap in interface form.
- `GoSuite.Run` calls `g.T.Run` and wraps the `*testing.T` it is handed:
  `f(GoSuite{t})`. Returning `g.T.Run(...)`'s result is the whole method.
