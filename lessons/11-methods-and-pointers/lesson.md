# Methods and Pointers

Structs gave you a way to hold data. Methods are how behaviour attaches to it,
and Go's version has one idea at the centre: a method is a function with an
extra parameter written before the name. Everything else - mutation, method
sets, nil receivers - falls out of that.

## A method is a function with a receiver

```go
type Counter struct{ N int }

func (c Counter) Value() int      { return c.N }
func (c *Counter) Add(delta int)  { c.N += delta }
```

`c` is a parameter. It follows the same rule every other parameter follows: a
value is copied, a pointer is not. So `Value` reads a copy, which is fine
because it only reads, and `Add` has the address, which is why it can write
something the caller will see. A value-receiver `Add` would compile, run,
increment a copy and throw it away.

Methods can be declared on **any named type defined in the same package**, not
only on structs:

```go
type Temperature float64

func (t Temperature) String() string { return fmt.Sprintf("%.1fC", float64(t)) }
```

Note `float64(t)` there. Formatting `t` with `%f` would call `String` again, and
that recursion is a stack overflow, not a compile error. It is the classic way
to hang a `String` method.

The receiver type must be defined in the same package, which is why you cannot
add a method to `int` or to `time.Time`. Define your own named type over it
instead.

## Calling: addressability

Go inserts the `&` and the `*` when it can:

```go
var c Counter
c.Add(3)          // really (&c).Add(3)

p := &Counter{}
p.Value()         // really (*p).Value()
```

The first line works because `c` is a variable, and a variable is
**addressable**. Plenty of things are not: a map element, the result of a
function call, a constant, a field of any of those. `m["a"].Add(1)` does not
compile at all, because there is no address to take. That is one of the reasons
maps of structs are usually maps of pointers.

The addressable-but-wrong case is the one that costs time:

```go
for _, c := range counters { c.Add(1) }     // compiles, does nothing
for i := range counters    { counters[i].Add(1) } // adds to the slice
```

`c` is a fresh variable holding a copy of the element, so `&c` exists and the
call is legal. It just mutates the copy. Go 1.22 gave each iteration its own
`c`, which fixed a different bug and did nothing for this one. Any time a loop
over a slice means to change it, index the slice.

## Method sets

The receiver kind decides what an interface will accept:

- the method set of `*T` holds every method, value receiver or pointer receiver
- the method set of `T` holds only the value-receiver ones

```go
type Adder interface{ Add(int) }

var _ Adder = (*Counter)(nil)   // fine
var _ Adder = Counter{}         // does not compile
```

The asymmetry is exactly the addressability rule again: given a `*Counter` the
compiler can always produce a `Counter` by dereferencing, but given some
arbitrary `Counter` value stored in an interface there is no address to hand to
a pointer method. Interfaces come next; the rule is worth remembering now,
because the error message it produces ("does not implement ... method has
pointer receiver") is the one you will see most.

## Choosing a receiver

The practical guidance, roughly in the order it applies:

1. If the method modifies the receiver, it must be a pointer receiver.
2. If the type contains a `sync.Mutex` or anything else that must not be
   copied, pointer receiver.
3. If the type is large, pointer receiver - though a struct with a slice or map
   field is small, because only the header is copied.
4. Otherwise either works, and **consistency wins**: pick one kind per type and
   use it for all of that type's methods, so a reader never has to check.

Small immutable values - a `time.Time`, a `Temperature`, a `Point` - are the
usual value-receiver types. Almost everything else ends up on pointers.

## Constructors and the zero value

Go has no constructors. `NewThing` is an ordinary function that returns a
`*Thing`, and returning the address of a local is completely safe:

```go
func NewAccount(owner string, balance int) *Account {
	return &Account{Owner: owner, Balance: balance}
}
```

The compiler's escape analysis sees the pointer outlive the call and puts the
value on the heap. There is no dangling pointer to create and no `delete` to
forget.

Better still is a type that needs no constructor. A `[]T` field works from zero,
because `append` on a nil slice allocates:

```go
var s Stack
s.Push(1)          // fine
```

A `map` field does not. Reading a nil map returns the zero value, but writing to
one panics with "assignment to entry in nil map", so a type with a map field
either has a constructor or makes the map inside every method that writes.
`bytes.Buffer`, `strings.Builder` and `sync.Mutex` are all designed so that
`var b bytes.Buffer` is ready to go, and that is the standard to aim for.

## Nil receivers

Calling a method on a nil pointer is not itself a panic. The method runs, the
receiver is nil, and only a dereference fails:

```go
func (a *Account) Describe() string {
	if a == nil {
		return "<no account>"
	}
	return fmt.Sprintf("%s: %d", a.Owner, a.Balance)
}
```

So a pointer-receiver method can decide what "absent" means, which is how a nil
`*Node` can be a perfectly good empty tree and how a nil `*log.Logger` wrapper
can be a no-op. The contract has to be deliberate: either the method handles
nil, or the caller guarantees non-nil. A value receiver is never in this
position, because there is no nil `Account`.

## Aliasing

`&accounts[i]` is the address of the element, so a pointer handed back from a
function is a live view of the slice:

```go
p := Richest(accounts)
p.Deposit(100)      // accounts[i] changed as well
```

Useful, and sharp. Two things to keep in mind. A pointer into a slice keeps the
entire backing array alive for as long as you hold it. And `append` may move the
elements to a new array, after which the old pointer refers to the old array and
quietly stops agreeing with the slice. Take pointers into a slice you are not
about to grow.

The opposite move is a copy: `copied := *a` dereferences into a new value, and
`&copied` is a pointer nobody else holds. It is a **shallow** copy - each field
is copied, so a pointer or slice field is still shared with the original.

## Further reading

- [Method declarations](https://go.dev/ref/spec#Method_declarations) - the
  receiver, and the rule that its type must be defined in the same package.
- [Method sets](https://go.dev/ref/spec#Method_sets) - the one paragraph behind
  every "method has pointer receiver" error you will meet.
- [Address operators](https://go.dev/ref/spec#Address_operators) - what is
  addressable, which is why `m["a"].Add(1)` does not compile.
- [FAQ: values or pointers?](https://go.dev/doc/faq#methods_on_values_or_pointers):
  the Go team's own answer, including the consistency rule.
- [Effective Go: pointers vs. values](https://go.dev/doc/effective_go#pointers_vs_values):
  the same choice seen from the caller's side.

## Practise

Three challenges. The first declares the same behaviour both ways: `Add` on a
pointer receiver, `Plus` on a value receiver returning its modified copy, a
loop that has to index the slice, and a `String` method on a named `float64`.
The second builds two types with state - a `Stack` whose zero value works and a
`Tally` whose map field means it does not - and drains a `[]Stack` through the
elements rather than through copies. The third is the pointer rules on their
own: a constructor returning the address of a local, methods that survive a nil
receiver, a function returning a pointer that aliases a slice element, and a
`Clone` that deliberately does not.
