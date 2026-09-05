# Testing

Testing in Go is not a library you choose. It is a file naming convention, a
function naming convention, and a command:

```
foo.go        the code
foo_test.go   its tests, in the same directory
go test ./...
```

There is no runner to install, no config, no `describe`, no `expect`, no
matcher DSL, and - a deliberate omission that surprises everyone once - no
assertion function. The standard library gives you `if` and `t.Errorf`, and
takes the view that a test is a program like any other, so the language you
already know is the language you write tests in.

## The four function shapes

A `_test.go` file may contain four kinds of function, and the tool finds them
by name:

```go
func TestJoin(t *testing.T)       { ... }   // go test
func BenchmarkJoin(b *testing.B)  { ... }   // go test -bench=.
func FuzzJoin(f *testing.F)       { ... }   // go test -fuzz=Fuzz
func ExampleJoin()                { ... }   // go test, and it appears in the docs
```

`Test`, `Benchmark`, `Fuzz` and `Example` are prefixes the toolchain looks for;
the character after the prefix must not be lowercase, so `TestJoin` is a test
and `Testing` is just a function. An `Example` with an `// Output:` comment is
compiled, run, and its output compared - documentation that cannot go stale.

Two package options, and you will use both:

- `package foo` - an *internal* test. It sees unexported identifiers.
- `package foo_test` - an *external* test, in the same directory, which sees
  only the exported API. It is the one that tells you whether your package is
  usable from outside, and it can import packages that import `foo` without a
  cycle.

## The failure methods

`*testing.T` has few methods and they divide cleanly:

| Method | Effect |
| --- | --- |
| `t.Errorf` | mark failed, keep going |
| `t.Fatalf` | mark failed, stop this test now |
| `t.Log` | record output, shown for failures or with `-v` |
| `t.Skip` | stop, and do not count it as a failure |
| `t.Helper` | this function is a helper, blame my caller |
| `t.Cleanup` | run this when the test finishes, however it finishes |
| `t.Run` | start a subtest |
| `t.Parallel` | this test may run alongside its siblings |

`Fatalf` ends the test by calling `runtime.Goexit`, which runs deferred
functions and then abandons the goroutine. It follows that `Fatalf` from a
goroutine **you** started stops that goroutine and not the test - the test then
finishes, sees no failure recorded, and passes. Failures from other goroutines
must go through `Errorf`, or through a channel back to the test goroutine.

There is no `t.Assert`, and the message is yours to write. The convention is
worth following exactly, because everyone else follows it:

```go
t.Errorf("JoinFields(%q, %q) = %q, want %q", parts, sep, got, want)
```

Call, inputs, got, want. "assertion failed" tells you nothing at 3am.

## The table

The dominant shape in Go, in the standard library and everywhere else, is the
table-driven test: a slice of cases, and one loop.

```go
func TestClamp(t *testing.T) {
	cases := []struct {
		name            string
		n, lo, hi, want int
	}{
		{"below the range", -4, 0, 10, 0},
		{"above the range", 40, 0, 10, 10},
		{"strictly inside", 7, 0, 10, 7},
		{"exactly at lo", 0, 0, 10, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Clamp(c.n, c.lo, c.hi); got != c.want {
				t.Errorf("Clamp(%d, %d, %d) = %d, want %d", c.n, c.lo, c.hi, got, c.want)
			}
		})
	}
}
```

Adding a case is one line. The comparison and the message exist once, so they
cannot drift between cases the way six copy-pasted test functions do. And the
rows are data, which means you can read the table and see what is covered - or,
more usefully, what is not.

Two rules about the table itself:

- **The name column names the situation**, not the numbers. The numbers are
  already in the failure message; the name is there to tell you which case you
  had not thought about.
- **Work the expectations out by hand.** A `want` filled in with whatever the
  function printed is not a test, it is a snapshot of today's bug. That is how
  a defect becomes a requirement.

## Subtests

`t.Run(name, func(t *testing.T))` makes each row a test of its own, and the
inner `t` is a new one:

- The failure is reported as `TestClamp/below_the_range`, so you know the row
  without reading the message.
- `go test -run 'TestClamp/below'` runs that row alone. Spaces in a name become
  underscores so that argument stays shell-friendly.
- A `Fatalf` inside ends **that row**, not the loop.
- `t.Run` returns a `bool`, so you can skip follow-up work when a stage failed.
- `t.Parallel()` inside the subtest lets the rows run concurrently, and the
  parent waits for all of them before its own cleanups run.

The bug everybody writes once is using the outer `t` inside the closure:
failures then belong to the parent, and `Fatalf` kills the whole loop. Shadow
it deliberately - `func(t *testing.T)` - and it cannot happen.

## Helpers, and t.Helper

Test files grow helpers: `assertEqual`, `mustOpen`, `newFixture`. One line
decides whether they are useful:

```go
func assertEqual[V comparable](t testing.TB, got, want V) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
```

Without `t.Helper()`, every failure in the file is reported at the line inside
`assertEqual` - the one line you already know is fine. With it, the helper is
skipped when the line is chosen and you get the call site.

Note the parameter type. `testing.TB` is the interface `*testing.T` and
`*testing.B` share, so a helper that takes it works in tests and benchmarks
alike. You cannot implement `testing.TB` yourself - it has an unexported method
so that nobody passes a homemade `T` - so when you want a fake, to test the
helper itself, declare the two or three methods you actually use and take that
interface instead. That is a general Go move, not a testing trick: **accept an
interface, and the thing becomes testable**.

## Cleanup, not defer

`defer` inside a test works. `defer` inside a *helper* does not: it runs when
the helper returns, which is before the test has done anything at all.

`t.Cleanup(fn)` registers work for when the **test** finishes - on success, on
failure, and after `Fatalf` alike. It is what lets a helper own both halves of a
resource:

```go
func newFixture(t *testing.T) *Fixture {
	t.Helper()
	f := open()
	t.Cleanup(f.Close)
	return f
}
```

Cleanups unwind last-registered-first, like `defer`, so nested fixtures close in
the right order. A subtest's cleanups run when that subtest ends, not when the
parent does. `t.TempDir()` is the standard library using exactly this: a
directory removed for you when the test is over.

## Benchmarks

A benchmark is a loop the framework sizes:

```go
func BenchmarkJoin(b *testing.B) {
	parts := BuildParts(20000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SinkString = JoinFields(parts, ",")
	}
}
```

`go test -bench=.` runs the function with `b.N = 1`, then with larger values
until the loop has taken about a second, and divides. Everything else follows
from that:

- **The loop is mandatory.** A body that ignores `b.N` does its work once
  however many iterations were requested, so the framework keeps asking for
  more; an `N` of exactly `1000000000` in the output means the benchmark never
  looped.
- **`b.ResetTimer()` after the setup.** It zeroes the clock *and* the
  allocation counters. A function that does one allocation reported as 700
  allocs/op is a missing `ResetTimer` almost every time.
- **`b.StopTimer()` / `b.StartTimer()` around per-iteration setup**, when each
  iteration needs a fresh input - a slice the function consumes, a file it
  drains. They pause both the clock and the counters. They cost enough that you
  would not wrap a nanosecond in them, but for a rebuild bigger than the call
  they are the difference between measuring your function and measuring your
  fixture.
- **Assign the result to a package-level variable.** A result nobody reads is a
  body the compiler may delete, and a 0.3 ns/op benchmark is that, not a
  breakthrough.
- **`b.ReportAllocs()`** turns on the `B/op` and `allocs/op` columns for that
  benchmark. Allocation counts barely move when the machine is busy, so they
  are the number to trust and the number to optimise: most "faster Go" is
  really "fewer allocations".

Compare runs with `benchstat` rather than eyeballing two numbers; a 3% change
between two runs of the *same* code is normal.

## The rest of the toolchain

Everything below is a flag, not a dependency:

```
go test ./...                     the whole module
go test -run 'TestClamp/inside'   one test, or one subtest
go test -v                        show every test and its logs
go test -race                     the race detector, and it finds real bugs
go test -count=1                  bypass the cached result
go test -cover                    statement coverage
go test -bench=. -benchmem        benchmarks, with allocations
go test -fuzz=FuzzParse           generate inputs until something breaks
```

`-race` deserves the habit: it instruments memory access and reports a data
race when one actually happens, which is how the concurrency you wrote in the
last five lessons gets checked. CI should run it. Test results are cached per
package and set of inputs, which is why an unchanged package prints
`(cached)` - `-count=1` is the way to insist.

Coverage is a measure of lines executed, not of behaviour verified. It is
useful for finding the file nobody tested and worthless as a target.

## Recap

- `foo_test.go`, `TestX(t *testing.T)`, `go test ./...`. No library, no
  assertions.
- `package foo` sees the unexported names; `package foo_test` tests the API a
  user gets.
- `Errorf` continues, `Fatalf` stops - and `Fatalf` from another goroutine
  stops nothing that matters.
- Message shape: call, inputs, got, want.
- Table of cases, one loop, `t.Run` per row: named failures, `-run` targeting,
  and one broken row does not hide the others.
- Work every expectation out by hand; a pasted `want` freezes a bug.
- `t.Helper()` at the top of every helper, `testing.TB` as its parameter type.
- `t.Cleanup` over `defer` in helpers; cleanups unwind in reverse.
- Benchmarks: loop `b.N`, `ResetTimer` after setup, `StopTimer` around
  per-iteration rebuilds, assign to a sink, and read `allocs/op` first.
- `-race` in CI. Coverage is a hint, not a goal.

## Further reading

- [testing](https://pkg.go.dev/testing) - `T`, `B`, `Run`, `Helper`, `Cleanup`
  and `TB`, all of the API this lesson uses.
- [Add a test](https://go.dev/doc/tutorial/add-a-test) - the shortest official
  path from no test file to a passing `go test`.
- [Using Subtests and Sub-benchmarks](https://go.dev/blog/subtests) - `t.Run`,
  naming, and what `-run` can select once the rows have names.
- [The cover story](https://go.dev/blog/cover) - what `-cover` measures, and
  why it is a hint rather than a target.
- [Data Race Detector](https://go.dev/doc/articles/race_detector) - `-race`
  under test, which is where it belongs in CI.

## Practise

Three challenges. The first is the table itself: a function, a table of rows
whose expectations have to be right and whose coverage is checked, and a runner
that reports every failing row rather than the first. The second is the layer
above it - `Equal`, `Must` and a fixture helper written against an interface so
they can be tested, plus a subtest runner and the four-line adapter that turns
`*testing.T` into it. The third is measurement: three functions with an
allocation budget, and two benchmarks whose shapes are themselves graded, so a
missing `ResetTimer` or `StopTimer` fails rather than quietly lying to you.
