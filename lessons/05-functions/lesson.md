# Functions

A Go function declaration is `func`, a name, a parameter list, a result list
and a body:

```go
func Add(a int, b int) int {
	return a + b
}
```

When consecutive parameters share a type you can write it once, which is
convention rather than a trick: `func Add(a, b int) int`. The result type comes
after the parameters, and a function that returns nothing simply omits it.

Everything about Go functions follows from three decisions the language made:
they can return more than one value, they are values themselves, and there is
no overloading. That last one is worth saying out loud early: there is exactly
one `Add` in a package. No same-name-different-signature, and no default
argument values either. If you want two behaviours, you write two names, and the
call site tells the reader which one it got.

## Multiple return values

A function can return any number of values, and the second one is very often a
`bool` or an `error` saying whether the first one means anything:

```go
func Divide(a, b int) (int, bool) {
	if b == 0 {
		return 0, false
	}
	return a / b, true
}

q, ok := Divide(7, 2)
```

This is the **comma-ok** shape, and once you have seen it you see it everywhere:
map lookups, type assertions, channel receives, `strings.Cut`. It replaces
several things other languages need separate machinery for. There are no out
parameters, because a second result is easier to read than a pointer you write
through. There is no sentinel value like `-1` for "not found", because the
sentinel is always someone's real data eventually. And there are no exceptions,
because a failure is a value the caller has to acknowledge in the assignment.

When `ok` is `false`, the other results are the zero value. That is a
convention, not a rule, but it is a strong one: return `0, false`, never
`half, false`. A caller who forgets to check the flag then gets something
harmless rather than something plausible.

The caller must do something with every value. `Divide(7, 2)` on its own does
not compile in an assignment context, and an unused variable is a compile error
too. To take one result and discard another, use the **blank identifier**:

```go
_, ok := Divide(7, 0)
```

`_` is not a variable. It cannot be read, and it is how you say "I know there is
a value here, I do not want it" in a way the compiler can see.

## Named results

The result list can name its values, which declares them as ordinary variables,
zeroed on entry:

```go
func MinMax(nums []int) (lo, hi int, ok bool) {
	if len(nums) == 0 {
		return   // 0, 0, false
	}
	lo, hi, ok = nums[0], nums[0], true
	...
	return lo, hi, ok
}
```

Two things come with the names. They document the results in a way the type
alone cannot: `(lo, hi int, ok bool)` says which one is which, and that shows up
in the generated documentation. And a bare `return` returns whatever the named
results currently hold.

Bare returns are worth a rule of thumb. Use one for an early exit like the empty
case above, where the reader can see the whole story on one line. Do not use one
at the end of a forty-line function, where the reader has to reconstruct what
the results hold by tracing every assignment. The Go standard library uses named
results heavily and bare returns sparingly, and that is the balance to copy.

Named results have one more use, which the next lesson leans on: a deferred
function can modify them after the `return` statement has run. That is how
`recover` turns a panic into an error value.

## Variadic parameters

The last parameter can accept any number of arguments:

```go
func Sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

Sum()            // 0
Sum(1, 2, 3)     // 6
```

Inside the function, `nums` is a plain `[]int`. With no arguments it is `nil`,
and `range` over `nil` runs zero times, so the empty case usually needs no code.

To pass a slice you already have, spread it with `...`:

```go
values := []int{1, 2, 3}
Sum(values...)
```

The two call forms are not the same underneath. `Sum(1, 2, 3)` allocates a fresh
slice for the call. `Sum(values...)` passes `values` itself, sharing its backing
array, so a variadic function that writes to its parameter writes to the
caller's data. Treat a variadic parameter as read-only unless you have a reason
not to, and if you need to sort or filter it, copy it first.

Spreading is also how you forward one variadic to another, which is what every
logging wrapper in every Go codebase does:

```go
func Describe(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
```

Leave the dots off and `args` becomes a single argument of type `[]any`, so the
verbs in the format string line up against the wrong things. It compiles. It is
`go vet`'s job to catch it, not the compiler's.

When a function needs *at least one* argument, put a required parameter in front
of the variadic: `func Largest(first int, rest ...int) int`. Now the signature
enforces it and there is no runtime check to write, or forget.

## Functions are values

A function has a type - `func(int) int` - and can be stored, passed and
returned like anything else:

```go
double := func(n int) int { return n * 2 }

func Apply(nums []int, f func(int) int) []int
Apply(nums, double)
```

A function written without a name is a **function literal**, and it can be
called immediately, assigned, or handed straight to another function. This is
what lets `sort.Slice` take a comparison, `http.HandleFunc` take a handler and
`Apply` above take any transformation at all, without any of them knowing what
they are calling.

## Closures

A function literal declared inside another function can refer to that function's
variables, and those variables live as long as the literal does:

```go
func Counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

next := Counter()
next()   // 1
next()   // 2
```

`n` is not copied into the returned function. The function holds a reference to
it, and the compiler quietly moves `n` off the stack so it can outlive the call
to `Counter`. You do not manage that, and you do not need to think about it
beyond knowing that it happens.

Two consequences, and both are the point rather than a caveat:

- **Each call creates new state.** `Counter()` twice gives two functions with
  two separate `n`s, counting independently.
- **Closures over the same variable share it.** Return two functions from one
  call and they see each other's writes. That is how you build a getter and a
  setter over one private value, and it is also how you write a data race if
  those functions end up on different goroutines.

Closures are the ordinary way to carry a little state in Go without declaring a
type: a counter, a running total, a cache, an iterator that remembers where it
got to. When the state grows past two or three fields, that is the signal to
reach for a struct with methods instead, which is a concept five lessons from
here.

### The loop variable, then and now

The most famous Go closure bug used to look like this:

```go
for _, n := range ns {
	fns = append(fns, func() int { return n })
}
```

Before Go 1.22 there was one `n` for the whole loop, so every closure captured
the same variable and they all returned the last element. Go 1.22 changed the
spec: the loop variable is a new variable on each iteration, and this code now
does what it looks like it does.

It is still worth knowing, for two reasons. You will read code written before
the change, including the `n := n` shadowing line people used to add to work
around it. And a module whose `go.mod` declares a language version below 1.22
still gets the old behaviour, because Go changed the semantics per module rather
than breaking existing builds.

## Further reading

- [Function declarations](https://go.dev/ref/spec#Function_declarations) - the
  spec on parameters, named results and variadic signatures.
- [Function literals](https://go.dev/ref/spec#Function_literals) - closures, and
  the sentence saying they capture variables rather than values.
- [Effective Go: functions](https://go.dev/doc/effective_go#functions) - multiple
  return values, named results and `defer`, in the order they were designed.
- [For statements](https://go.dev/ref/spec#For_statements) - the per-iteration
  loop variable Go 1.22 introduced, defined in the spec itself.

## Practise

Three challenges. The first returns two and three values at a time, with named
results, comma-ok and the blank identifier. The second is variadic parameters
from both ends: collecting arguments, spreading a slice, and forwarding one
variadic into another. The third makes functions into values - counters,
accumulators, a `map` over a slice, composition, and one closure per element of
a slice.
