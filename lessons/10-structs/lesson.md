# Structs

Go has one way to build a new aggregate type, and this is it. A struct is a
fixed set of named fields laid out next to each other in memory. There are no
classes, no inheritance and no constructors, and once you stop looking for them
the whole design gets simpler rather than poorer.

## Declaring and making one

```go
type Book struct {
	Title  string
	Author string
	Pages  int
}
```

That declares a type. Fields beginning with a capital letter are exported and
visible from other packages; lowercase ones are not, and that is the only
visibility control a struct has.

Three ways to make a value, and only one is a habit worth keeping:

```go
var b Book                                     // the zero value
b := Book{Title: "Learning Go", Pages: 376}    // named fields
b := Book{"Learning Go", "Bodner", 376}        // positional
```

Prefer the named form. The positional form must list every field in declaration
order, so adding a field later breaks every positional literal in the program -
or, if the types happen to line up, keeps compiling with the values shifted by
one. Named fields may be given in any order and any subset; whatever you leave
out is zero.

The zero value matters more in Go than in most languages. `var b Book` is
immediately usable: `""`, `""`, `0`. A well-designed struct is one whose zero
value is already meaningful, which is why `var b strings.Builder` and
`var mu sync.Mutex` need no setup. A `NewBook` function is an ordinary function
that happens to return a `Book`; the language gives it no special standing, and
you only write one when there is real work to do.

Fields are read and written with a dot:

```go
b.Pages = 400
```

Nested structs nest their literals, and an anonymous struct type is legal
anywhere a type is - which is what makes the table-driven test pattern work:

```go
tests := []struct {
	in   string
	want int
}{
	{"a", 1},
	{"bb", 2},
}
```

That slice's element type has no name and needs none. It exists for the length
of the test.

## Structs are values

This is the point that trips everyone arriving from Java, Python or JavaScript.
A struct is a value, not a reference. Assigning one copies it. Passing one to a
function copies it. Putting one in a slice copies it in, and taking it back out
copies it out.

```go
func addPages(b Book, n int) {
	b.Pages += n     // changes the local copy, and nothing else
}
```

Two ways to actually change something: return the modified copy, or take a
`*Book`. Both are idiomatic, and which one to reach for is the subject of the
next lesson. For now, notice that a copy is cheap for a small struct and is what
makes `Book` safe to hand around: nobody else can change yours behind your back.

The copy is **shallow**, which is the one caveat. A struct holding a slice or a
map copies the header, not the contents, so the copy and the original share the
same backing array:

```go
c := original            // a copy...
c.Hosts[0] = "changed"   // ...that still writes through to the same array
```

Everything the slices lesson said about aliasing applies here, one level down.

Because a struct is a value, a slice of structs holds them inline rather than as
pointers, and ranging over one gives you copies:

```go
for _, b := range shelf {
	b.Pages = 0          // pointless: b is a copy of the element
}
for i := range shelf {
	shelf[i].Pages = 0   // this one writes to the slice
}
```

## Embedding: composition, not inheritance

A field declared with a type and no name is **embedded**:

```go
type Base struct {
	ID   int
	Name string
}

type Article struct {
	Base           // embedded
	Title string
}
```

`Article` now has a field whose name is its type, `Base`, plus a shortcut:
`Base`'s fields are **promoted**, so `a.ID` compiles and means `a.Base.ID`.
Nothing is copied into `Article`, and no subtype relationship is created. An
`Article` is not a `Base` and cannot be passed where one is expected. It has
one, and it knows a shorter way to say so.

In a composite literal the embedded field's key is its type name, and there is
no way to set a promoted field from the outer literal:

```go
a := Article{
	Base:  Base{ID: 7, Name: "base"},
	Title: "Embedding",
}
// Article{ID: 7} does not compile.
```

Assignment through a promoted selector does work, because a promoted selector is
an ordinary selector with the middle left out: `a.ID = 8`.

Promotion is a name lookup with one rule: **shallowest wins**. A field the outer
struct declares itself sits at depth 0 and shadows anything embedded:

```go
type Feature struct {
	Article        // Article and Base are both inside here
	Name string    // depth 0, so this is what f.Name means
}
```

The shadowed field is not gone. `f.Article.Base.Name` still reaches it.
Shadowing changes what the short spelling means and nothing else.

When two candidates sit at the **same** depth, neither wins:

```go
type Article struct {
	Base           // has Name
	Meta           // also has Name
	Title string
}

a.Name    // compile error: ambiguous selector a.Name
a.ID      // fine, only Base has it
```

The struct itself is legal; only the ambiguous selector is rejected, and you fix
it by naming the level you meant. There is no method resolution order here and
no rule about which parent wins, because the compiler declines to have an
opinion. That is the whole difference between composition and inheritance in one
error message.

Embedding earns its keep when methods enter the picture, in the next lesson: an
embedded type's methods are promoted too, which is how a struct satisfies an
interface by embedding something that already does. Field promotion is the same
mechanism, seen early.

Two smaller notes. Embedding a **pointer** (`*Base`) promotes just the same but
gives you a nil to dereference if you forget to set it. And an embedded field's
name for tags, for `%+v` and for `encoding/json` is the type name, which is why
an embedded struct shows up as a nested object in JSON unless you say otherwise.

## Equality

Two structs of the same type compare with `==` when **every field is
comparable**, and the comparison is field by field:

```go
Point{1, 2} == Point{1, 2}   // true
```

No method, no interface, no `Equals`. It is a language rule, and it is what
makes a comparable struct legal as a **map key**.

| Comparable | Not comparable |
| --- | --- |
| numbers, strings, bools | slices |
| pointers, channels | maps |
| interfaces | funcs |
| arrays of comparable types | anything containing the above |

The rule is recursive, so a single slice field anywhere inside makes the whole
struct uncomparable - and `a == b` on two of them is a **compile** error, as is
`map[Config]int`. Caught where you wrote it, which is the good outcome.

Note that `[3]int` is comparable and `[]int` is not. An array carries its length
in its type and its elements in its value; a slice is a header pointing
elsewhere, and Go will not guess whether you meant the header or the contents.

When `==` is unavailable, write the comparison by hand:

```go
func EqualConfigs(a, b Config) bool {
	return a.Name == b.Name && slices.Equal(a.Hosts, b.Hosts)
}
```

It compiles to straight-line code, it is obvious, and you choose the edge cases
- in particular whether a nil slice equals an empty one. `reflect.DeepEqual`
works on anything, but it is slow, invisible to the compiler, and stricter than
people expect: `DeepEqual([]string(nil), []string{})` is **false**. Reach for
`slices.Equal` and `maps.Equal` when you know the shape, and keep `DeepEqual`
for tests over data you do not control.

Two traps are worth carrying away. First, a **pointer field compares
addresses**. A struct holding a `*Node` is comparable, but `==` asks "the same
object?", not "the same contents?", so two separately allocated but identical
nodes are unequal.

Second, comparing **interfaces can panic**. `==` on two `any` values always
compiles, because interfaces are comparable as a type. At run time it compares
the dynamic types and then the values - and if that dynamic type turns out to be
a slice or a map, the program panics with "comparing uncomparable type
[]string". The one place the compiler cannot protect you is exactly the place a
runtime failure waits. `reflect.TypeOf(v).Comparable()` asks first, and returns
a nil type for a nil interface, so check for that before calling it.

Finally, field names, types and order are all part of the struct type. `struct{A,
B int}` and `struct{B, A int}` are different types and cannot be compared at
all.

## Further reading

- [Struct types](https://go.dev/ref/spec#Struct_types) - the spec on fields, tags,
  embedding and the promotion rules that make a selector ambiguous.
- [Effective Go: composite literals](https://go.dev/doc/effective_go#composite_literals):
  named-field literals, and taking the address of one.
- [Effective Go: embedding](https://go.dev/doc/effective_go#embedding) - why
  composition is the whole answer and inheritance is not missing.
- [Comparison operators](https://go.dev/ref/spec#Comparison_operators) - which
  struct types are comparable, and therefore which can be a map key.
- [reflect](https://pkg.go.dev/reflect) - the home of `DeepEqual`, and its warning
  about what it does and does not treat as equal.

## Practise

Three challenges. The first is the struct itself: named-field literals, the zero
value, and a function that has to return its modified copy because the argument
was one. The second is embedding, with two embedded types that both declare a
`Name` so the promoted selector is ambiguous and has to be spelled out, and an
outer field that shadows both. The third is equality: a comparable struct as a
map key, a hand-written comparison for one that is not, the nil-versus-empty
disagreement between `slices.Equal` and `reflect.DeepEqual`, and a pointer field
that compares addresses.
