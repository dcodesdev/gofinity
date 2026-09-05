# Defer, Panic and Recover

A function has more than one exit. There is the last line, but also every early
`return`, every `return` inside a loop, and the failure three calls down that
unwinds straight through it. Anything the function acquired has to be released
on all of them, and the version of that code which stays correct as the function
grows is the one written **once, next to the acquire**.

That is `defer`:

```go
f, err := os.Open(path)
if err != nil {
	return err
}
defer f.Close()
```

Two lines that belong together, written together. Every path out of the function
from there on closes the file, including the ones nobody has written yet.

## What defer actually does

`defer` takes a function call and schedules it to run when the surrounding
function returns. Not at the end of the block, not at the end of the loop
iteration: when the **function** returns.

Three rules cover every use of it.

### Deferred calls run last-in, first-out

They go on a stack. The last one deferred is the first one to run:

```go
defer fmt.Println("first")
defer fmt.Println("second")
defer fmt.Println("third")
// prints third, second, first
```

That reversal is not a quirk to work around, it is the behaviour you want. You
acquire a lock, then a file; you release the file, then the lock. Deferring each
release as you acquire gives you the correct order for free, with no ordering
logic to keep in your head and nothing to update when a fourth resource appears
in the middle.

### The arguments are evaluated immediately

This is the one that surprises people:

```go
n := 1
defer fmt.Println(n)   // prints 1
n = 99
```

`defer` evaluates the function value and its arguments **at the `defer`
statement** and stores them with the pending call. Only the call itself is
postponed. `n` was 1 when the line ran, so 1 is what was saved.

A deferred *closure* behaves the opposite way, because it takes no arguments and
reads the variable when its body executes:

```go
n := 1
defer func() { fmt.Println(n) }()   // prints 99
n = 99
```

Both are useful, and the difference is one line of syntax. Snapshot a value by
passing it as an argument; track a variable by closing over it. Knowing which
one you wrote is most of the skill.

### A deferred function can change a named result

Deferred calls run **after** the return values have been set but **before** the
caller gets them. If the results are named, they are ordinary variables, so a
deferred closure can still edit what the caller receives:

```go
func double(n int) (result int) {
	defer func() { result *= 2 }()
	result = n
	return result   // sets result = n, then the defer doubles it
}
```

This is the forward reference from the functions lesson, and on its own it is a
party trick. Paired with `recover` below it becomes the standard way a function
reports that it caught something.

### defer is scoped to the function

```go
for _, name := range names {
	f := open(name)
	defer f.Close()   // does not close until the whole loop is done
}
```

Nothing closes until the function returns, so a loop over ten thousand names
holds ten thousand open files. The fix is to give each iteration its own
function, either a helper you call or a literal you invoke on the spot:

```go
for _, name := range names {
	func() {
		f := open(name)
		defer f.Close()
	}()
}
```

The helper is usually better, because it can also return an error.

## Panic

Most failures in Go are values you return. Some are not. Dividing an integer by
zero, indexing past the end of a slice, dereferencing a nil pointer: those
**panic**. So does calling `panic` yourself.

A panic stops the normal flow of the function, runs its deferred calls, then
does the same to its caller, and its caller, all the way up. If it reaches the
top of the goroutine, the program prints the panic value and a stack trace and
exits.

The deferred calls still run on the way, which is what makes `defer` a real
guarantee rather than a happy path convenience: your file is closed and your
lock released even while the stack is coming apart.

Panic is not Go's error handling, and reaching for it is almost always the wrong
call. Return an `error` for anything a caller could reasonably expect - a
missing file, bad input, a failed request. Panic is for the case where
continuing makes no sense: a programmer error, or a broken invariant that means
the rest of the function's assumptions are already false. The standard library
follows the same split, which is why there are pairs like `regexp.Compile`
returning an error and `regexp.MustCompile` panicking, the `Must` version being
for package-level variables where a failure means the program cannot start.

## Recover

`recover` stops a panic and hands you the value that was passed to it:

```go
func Guard(body func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	body()
	return nil
}
```

Everything about that shape is load-bearing.

`recover` only works **inside a deferred function** of the call that is
panicking. Called directly in the function body it returns `nil` and does
nothing, and called from a function the deferred one invokes it does nothing
either. It has to be the deferred function itself.

Its result is an `any`, because that is what `panic` accepts. A runtime failure
gives you a `runtime.Error`; your own code can hand over anything. An `error` is
the best thing to panic with, because whoever recovers it can pass it straight
back out with no reformatting and no lost identity.

`if r := recover(); r != nil` is the idiom, and the `nil` check matters: a
function that did not panic still runs its defers, and `recover` returns `nil`
there. (Since Go 1.21, `panic(nil)` panics with a `*runtime.PanicNilError`
rather than being silently indistinguishable from no panic at all.)

And the recovery is only useful if the function can report it, which is why
`err` is a **named result**. The deferred closure sets it after the `return`
has already chosen its values.

Reset the other results while you are there. Return `0, err`, not whatever half
a computation left behind. It is the same rule as comma-ok: a caller who ignores
the error should get something harmless rather than something plausible.

Recovering is for **boundaries**, not for sprinkling over risky-looking code. A
server that handles one request per goroutine recovers at the top of the handler
so a bug in one request does not take down every other connection. A library
that uses panic internally to unwind out of a deep recursive parser recovers at
its public function and returns an error, so the panic never escapes its own
package. Outside cases like those, catching a panic hides a bug you would rather
see immediately, in a stack trace, right where it happened.

## Further reading

- [Defer, panic and recover](https://go.dev/blog/defer-panic-and-recover) - the
  blog post that introduced all three, with the LIFO order and the named-result
  trick spelled out.
- [Defer statements](https://go.dev/ref/spec#Defer_statements) - the spec on
  when arguments are evaluated and when the call actually runs.
- [Handling panics](https://go.dev/ref/spec#Handling_panics) - the spec rules for
  `panic` and `recover`, including what makes a `recover` call effective.
- [Effective Go: recover](https://go.dev/doc/effective_go#recover) - recovering at
  a package boundary and turning the panic back into an error.

## Practise

Three challenges. The first is `defer` on its own: the LIFO order, arguments
evaluated at the defer statement versus a closure that reads the variable later,
and a deferred call editing a named result. The second is `recover` at a
boundary, turning a divide-by-zero and an out-of-range index into errors and
converting arbitrary panic values into an `error`. The third is the reason the
whole thing exists: releasing what you acquired on every exit path, including
the loop that needs a function per iteration and the panic that has to become an
error without leaking the resource.
