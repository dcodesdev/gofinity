# Struct Basics

A struct is a fixed set of named fields laid out next to each other in memory.
It is Go's only way to build a new aggregate type, and it is the type you will
write more of than any other.

```go
type Book struct {
	Title  string
	Author string
	Pages  int
}
```

That declares a **type**, not a variable. Fields starting with a capital letter
are exported, which matters the moment the struct is read from another package.

## Making one

There are three ways, and only one of them is a good habit:

```go
var b Book                                          // the zero value
b := Book{Title: "Learning Go", Pages: 376}         // named fields
b := Book{"Learning Go", "Bodner", 376}             // positional
```

Use the named form. The positional form has to list every field, in declaration
order, so adding a field to the struct silently breaks every positional literal
in the program - or worse, still compiles with the values shifted by one.

The zero value is always available and always meaningful: `var b Book` gives
`""`, `""` and `0`. There is no "uninitialised struct" in Go and no constructor
is required. A function named `NewBook` is just an ordinary function that
returns a `Book`; the language gives it no special status.

Fields are read and written with a dot, and a struct value in a variable is
addressable, so you can assign straight into it:

```go
b.Pages = 400
```

## Structs are values

This is the part that trips people coming from Java or Python. A struct is a
value, not a reference. Assigning one **copies** it, and passing one to a
function **copies** it:

```go
func addPages(b Book, n int) {
	b.Pages += n     // changes the copy, and nothing else
}
```

Calling that does nothing the caller can see. Two ways out: return the modified
copy, or take a pointer. This challenge takes the first route, and pointers get
a lesson of their own.

Comparison follows the same logic. Two structs of the same type are `==` when
every field is `==`, so `Book{...} == Book{...}` compares the three fields
rather than any notion of identity.

## Task

Fill in the six functions in `main.go`.

1. `NewBook` builds a `Book` with a named-field literal.
2. `Describe` renders `"Title by Author (N pages)"`.
3. `AddPages` returns a copy with more pages, leaving the caller's book alone.
4. `TotalPages` sums a slice of books.
5. `Longest` finds the book with the most pages, earliest wins on a tie.
6. `ByAuthor` filters on an exact author match.

## Hints

- `fmt.Sprintf("%s by %s (%d pages)", ...)` is all `Describe` needs. There is no
  special case for the zero `Book`: its fields already render as `""` and `0`.
- In `AddPages`, `b` is already a copy. Write to it and return it.
- `Longest` cannot start from `Book{}` as the running best, because a shelf of
  books with 0 pages would then never beat it. Start from `books[0]` after
  checking the length.
- For "earliest wins on a tie", compare with `>` rather than `>=`.
- `ByAuthor` should start from `out := []Book{}` so that "no matches" is an
  empty slice rather than `nil`. Both have length 0, but only one of them
  renders as `[]`.
