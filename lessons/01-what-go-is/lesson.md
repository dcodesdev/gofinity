# What Go is

Go is a small, compiled, statically typed language built at Google in 2009 for
a specific problem: large programs, written by large teams, that have to keep
compiling quickly and running predictably for years.

Almost every design decision follows from that. Go has no inheritance, no
exceptions, no generics until it needed them, and exactly one way to format
source code. The language is small enough to hold in your head, which is the
point: you spend your attention on the program, not on the language.

What you do get, and what nothing else gives you as cheaply:

- **A real compiler, fast.** Go compiles a large program in seconds to a single
  static binary with no runtime to install.
- **Concurrency in the language.** Goroutines and channels are syntax, not a
  library you pick.
- **A standard library that ships production code.** HTTP servers, JSON, crypto
  and testing are all in the box.
- **One formatting.** `gofmt` settles every style argument before it starts.

## The shape of a program

Here is a complete Go program.

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, Gofinity!")
}
```

Four things are happening, and all four matter.

### `package main`

Every Go file begins by declaring the package it belongs to. All files in one
directory belong to the same package, and the directory is the unit of
compilation.

`main` is special: it is the package that produces an executable. Any other
name produces a library, meant to be imported.

### `import "fmt"`

Imports name the packages this file uses. `fmt` ("format") is the standard
library's printing and formatting package.

Go is strict here in a way that surprises newcomers: **an import you do not use
is a compile error**, not a warning. So is a declared local variable you never
read. The reasoning is that unused code is almost always a mistake or a
leftover, and catching it at compile time keeps it from accumulating.

```go
import (
	"fmt"
	"strings"
)
```

That is the parenthesised form, used whenever there is more than one.

### `func main()`

`func` declares a function. `main` in package `main` is the entry point: it
takes no arguments, returns nothing, and the program exits when it returns.

Other functions look the same, with parameters and a result type:

```go
func Greet(name string) string {
	return "Hello, " + name + "!"
}
```

The type comes *after* the name, both for parameters and for the result. That
reads backwards if you are coming from C or Java, and forwards once you get
used to it: "name, which is a string".

### `fmt.Println`

A call, qualified by its package. `Println` prints its arguments and a newline.

Its more useful sibling is `Sprintf`, which *returns* the formatted string
instead of printing it:

```go
fmt.Sprintf("Hello, %s!", name)   // "Hello, Ada!"
fmt.Sprintf("%d. %s", 3, "Alan")  // "3. Alan"
```

`%s` is a string, `%d` an integer, and `%v` is "whatever this value is,
sensibly". Building a string with `Sprintf` and returning it is far more common
in real code than printing, because a function that returns a string can be
tested. A function that prints cannot, easily.

## Exported and unexported

Capitalisation is not a style choice in Go, it is access control. An identifier
starting with a capital letter is **exported**: visible to code in other
packages. A lowercase one is private to its own package.

```go
func Greet(name string) string { ... }  // callable from anywhere
func greet(name string) string { ... }  // this package only
```

That is the whole rule. There is no `public` or `private` keyword.

## Building strings

Go has `+` for strings, and it works, but it allocates a new string every time.
For a fixed handful of pieces that is fine. For a list, build a slice of pieces
and join it once:

```go
lines := []string{"first", "second", "third"}
strings.Join(lines, "\n")   // "first\nsecond\nthird"
```

`Join` puts the separator *between* elements, never at the end, so joining an
empty slice gives `""` and joining one element gives that element unchanged.
Both edge cases fall out for free, which is why the idiom is worth learning
before you ever reach for `+=` in a loop.

## Tests are how you know

Every challenge here is graded by a `_test.go` file you can read. Go's testing
package is in the standard library and needs no configuration:

```go
func TestGreet(t *testing.T) {
	if got := Greet("Ada"); got != "Hello, Ada!" {
		t.Errorf("Greet(\"Ada\") = %q, want %q", got, "Hello, Ada!")
	}
}
```

A test fails by calling `t.Errorf`. That is the entire mechanism: no assertion
library, no matchers. `%q` prints a quoted string, which is how you see the
difference between `"Ada"` and `"Ada "` in a failure message.

You will meet this properly in the testing concept much later. For now, read
the test file before you write any code. It is the specification.

## Further reading

- [Tutorial: get started with Go](https://go.dev/doc/tutorial/getting-started) -
  installing the toolchain and running a first program, from the source.
- [Effective Go](https://go.dev/doc/effective_go) - the house style for the
  language, and the closest thing Go has to a book about itself.
- [The strings package](https://pkg.go.dev/strings) - `Join` and everything else
  worth reaching for before you write a loop over bytes.
- [The testing package](https://pkg.go.dev/testing) - what `*testing.T` offers
  beyond `t.Errorf`, which is what every challenge here is graded by.
