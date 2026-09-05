# Table-Driven Tests

Go has no assertion library in the standard library, no `describe`, no
`expect`. It has a function that takes a
[`*testing.T`](https://pkg.go.dev/testing#T), an `if`, and
[`t.Errorf`](https://pkg.go.dev/testing#T.Errorf). What it does have is a shape
that almost every test in the standard library follows, and once you know it
you can read anyone's tests:

```go
func TestClamp(t *testing.T) {
	cases := []struct {
		name            string
		n, lo, hi, want int
	}{
		{"below the range", -4, 0, 10, 0},
		{"above the range", 40, 0, 10, 10},
		{"strictly inside", 7, 0, 10, 7},
	}
	for _, c := range cases {
		got := Clamp(c.n, c.lo, c.hi)
		if got != c.want {
			t.Errorf("%s: Clamp(%d, %d, %d) = %d, want %d", c.name, c.n, c.lo, c.hi, got, c.want)
		}
	}
}
```

A slice of cases, one loop, one comparison. That is a **table-driven test**.

Adding a case is one line, not one function. The loop is written once, so the
comparison and the failure message are written once, and they cannot drift
between cases the way six copy-pasted test functions do.

## The name column

Every row carries a name, and the name is the first thing in the failure
message. Name the *situation*, not the numbers:

```go
{"below the range", -4, 0, 10, 0}    // good
{"minus four zero ten", -4, 0, 10, 0} // says nothing the numbers do not
```

The numbers are already in the message. The name is there to tell you which
case you forgot to think about.

## Errorf, not Fatalf

`t.Errorf` records a failure and carries on.
[`t.Fatalf`](https://pkg.go.dev/testing#T.Fatalf) records it and stops the test
function immediately.

Inside a table loop you want `Errorf`. If four rows are broken, one run should
tell you about four rows, not about the first and then nothing. Save `Fatalf`
for the case where carrying on is meaningless: the fixture failed to build, the
value you were about to inspect is nil.

## The table is code too

A table with a wrong `want` is worse than no test: it is a test that passes on
broken behaviour, or fails on correct behaviour and gets "fixed" by changing
the implementation. Work each expectation out by hand, from the specification,
not by running the function and pasting what it printed. Pasting what it
printed is how a bug becomes a requirement.

The other half is coverage. A table of five rows that are all "a number inside
the range" tests one branch five times. Go looking for the edges: the
boundaries themselves, the empty case, the negative case, the case where two
parameters are equal.

## Helpers take an interface

The runner in this challenge does not take `*testing.T`. It takes:

```go
type TB interface {
	Helper()
	Errorf(format string, args ...any)
}
```

`*testing.T` satisfies that, so real tests are unaffected - but so does a
little struct that records what it was told, which is how the tests here can
check that your runner reports the right rows without failing themselves.

That trick has a name in the standard library:
[`testing.TB`](https://pkg.go.dev/testing#TB) is the interface `*testing.T` and
[`*testing.B`](https://pkg.go.dev/testing#B) share, and helpers that work for
tests *and* benchmarks take it. You cannot implement `testing.TB` yourself - it
has an unexported method precisely so that nobody does - so when you want a
fake, you declare the two or three methods you actually use, exactly as above.

## t.Helper

When a helper calls `t.Errorf`, the failure is reported at the line inside the
helper. Every failure then points at the same line, which is the one line you
already know is fine.

[`t.Helper()`](https://pkg.go.dev/testing#T.Helper) fixes it: it marks the
calling function as a helper, so the reported line is the caller's. It costs
one line at the top of the function and it is the difference between
"main_test.go:74" for every failure and the line of the row that actually
broke.

## Task

Implement `Clamp`, `ClampCases` and `RunCases`.

`ClampCases` must return rows whose `Want` values are all correct, with unique,
non-empty names, covering seven situations between them: `n` below `lo`, `n`
above `hi`, `n` strictly inside, `n` exactly `lo`, `n` exactly `hi`, a range
that is entirely negative, and a range where `lo == hi`.

`RunCases` must run every case, report only the failing ones, name the case and
show both values in the message, and call `t.Helper()`.

## Hints

- `Clamp` is two `if`s and a return. Callers promise `lo <= hi`, so you do not
  have to defend against the opposite.
- One row can cover two situations - `{N: 9, Lo: 4, Hi: 4}` is both "above hi"
  and "lo == hi" - but eight clear rows read better than five clever ones.
- A negative range is `Lo: -8, Hi: -3`. Check the expectation by hand: clamping
  `-20` into it gives `-8`, not `0`.
- `RunCases` reports with `Errorf`, never `Fatalf`: the test that calls it
  wants to hear about all four broken rows.
- Put `t.Helper()` at the top of `RunCases`, not inside the `if`. It is cheap
  and it belongs with the signature.
