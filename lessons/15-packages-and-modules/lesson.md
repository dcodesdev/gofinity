# Packages and Modules

Everything so far lived in one file. Real programs do not, and Go's answer to
"where does this code go" is unusually rigid: a **package is a directory**, a
**module is a tree of directories with a name**, and whether a name escapes its
package is decided by its first letter. Three rules, no configuration.

```
myapp/
  go.mod                 module github.com/you/myapp
  main.go                package main
  textkit/
    textkit.go           package textkit
    slug.go              package textkit  (same directory, same package)
  bank/
    bank.go              package bank
```

## A package is a directory

Every `.go` file begins with a package clause, and every file in one directory
must name the same package. Splitting a package across three files is a
formatting decision with no semantic weight: the three files share one
namespace, and a function in `slug.go` can call an unexported helper in
`textkit.go` without importing anything.

The reverse is also true: two directories are always two packages, even if you
wanted otherwise. There are no nested packages. `textkit/internal/x` is not
"inside" `textkit` in any sense the compiler cares about; it is a separate
package that happens to sit below it in the filesystem.

`package main` is the one special name. A `main` package with a `func main()`
builds an executable; every other package builds a library and cannot be run.

## A module is the unit of versioning

`go.mod` marks the root of a module and gives it a name:

```
module github.com/you/myapp

go 1.24
```

That name is the prefix of every import path inside it. The package in
`textkit/` is imported as:

```go
import "github.com/you/myapp/textkit"
```

The path is a **directory path**, not a package name. The last element is the
directory; the identifier you type at the call site is the package clause in
those files. They are conventionally identical, and when they differ - a
directory called `v2`, a package called `yaml` in `gopkg.in/yaml.v3` - the
reader has to look it up, which is a good reason to keep them the same.

The `go` line is a language-version floor, not a toolchain pin: it says which
version of the language the files may use. A `require` block, when there is one,
lists the other modules this one depends on, and `go.sum` records their
checksums. Neither appears in these challenges, because the standard library is
not a dependency.

## Import declarations

```go
import (
	"fmt"
	"strings"

	"github.com/you/myapp/textkit"
)
```

Convention is standard library first, a blank line, then everything else;
`gofmt` sorts within each group and leaves the grouping to you. An **unused
import is a compile error** - Go would rather you delete it than let the build
carry weight nobody reads.

Two forms are worth knowing and both are worth avoiding by default. An alias
renames a package at the point of import - `crand "crypto/rand"` - and exists to
resolve a collision between two packages with the same name. The blank import
`_ "image/png"` imports a package purely for its `init` side effect, which is
how image decoders and database drivers register themselves; it is deliberately
ugly, because it is doing something invisible.

Imports may not form a **cycle**. If `bank` imports `textkit`, then `textkit`
can never import `bank` - not indirectly either. That single restriction shapes
Go codebases more than any style guide: it forces you to decide which of two
packages is the lower-level one, and the usual fix for a cycle is a third
package holding the types they both need.

## Capitalisation is the visibility keyword

There is no `public`, `private` or `protected`. A name starting with an
upper-case letter is **exported**; anything else is visible only within its own
package.

```go
func Title(s string) string   // any importer can call this
func isWordChar(r rune) bool  // this package only
```

It applies to every declared name: functions, types, methods, struct fields,
constants, package-level variables. And it is enforced by the compiler, so
`textkit.isWordChar` from another package is a build error rather than a
convention someone can talk themselves out of.

Struct fields are where this earns its keep:

```go
type Account struct {
	owner   string
	balance int
}

func (a *Account) Balance() int { return a.balance }
```

An exported field is a standing promise that anybody may assign anything to it.
Once `balance` is lower case, the only way it moves is through the methods you
wrote, so "a withdrawal larger than the balance is refused" becomes a rule the
type can actually keep rather than a hope. The getter costs one line and gives
up nothing.

Read the rule in the other direction too: everything you export is API. Somebody
will depend on it, and changing it later breaks them. Export the smallest
surface that lets a caller do the job, and grow it when you have to - shrinking
it later is the painful direction.

`internal/` is the escape hatch for the middle ground. A package under a
directory named `internal` can only be imported from within the subtree rooted
at that directory's parent, so `myapp/internal/store` is importable everywhere
inside `myapp` and nowhere outside it. That is how a library shares code between
its own packages without adding it to the public API.

## Naming, and the stutter

The package name is part of every call site, so it is part of the identifier a
reader sees. Short, lower case, no underscores, no plural, usually a noun:
`bank`, `textkit`, `strings`, `http`.

Because the qualifier is already there, repeating it in the member's name reads
badly:

```go
bank.NewBankAccount()   // bank, bank
bank.NewAccount()       // this one
```

The standard library is the style guide here: `http.Client`, not
`http.HTTPClient`; `strings.Reader`, not `strings.StringReader`. A package named
`util` fails the test in a different way - it says nothing about what its
members do, so it attracts everything and explains nothing.

## Documentation is a comment

A comment directly above a declaration, with no blank line between, is that
declaration's documentation, and `go doc` prints it:

```go
// Package textkit turns free-form text into the shapes a page needs.
package textkit

// Title returns s with the first letter of every word in upper case.
func Title(s string) string
```

The convention is to start with the name being declared, which makes the
generated docs read as sentences. A package comment goes above the package
clause in exactly one file of the package, conventionally `doc.go` when it is
long.

## Test files pick a side

A file ending in `_test.go` may declare either the package under test or that
package's name with a `_test` suffix:

- `package bank` - an **internal** test. It sees unexported names, so it can
  test a helper like `normalizeOwner` directly.
- `package bank_test` - an **external** test. It has to import `bank` and can
  only use the exported surface, which is exactly what a caller experiences.

Both are ordinary, and a package often has both files. The external form is
worth reaching for when you want to be sure the exported API is enough to use
the package, and it is the only form that can import a package that would
otherwise create a cycle.

`go test ./...` walks the module from the current directory and runs every
package it finds.

## Further reading

- [Go modules reference](https://go.dev/ref/mod) - `go.mod`, the `go` line,
  `require`, `go.sum`, and how a module path becomes an import path.
- [Managing dependencies](https://go.dev/doc/modules/managing-dependencies) -
  the everyday commands, once the standard library is not enough.
- [Organizing a Go module](https://go.dev/doc/modules/layout) - where `main`,
  the libraries and `internal/` go in a real tree.
- [Package names](https://go.dev/blog/package-names) - short, lower case, and
  why `bank.NewBankAccount` reads badly.
- [Go Doc Comments](https://go.dev/doc/comment) - the comment conventions
  `go doc` and pkg.go.dev render.

## Practise

Two challenges. The first splits a program across a `main` package and a
`textkit` package in a subdirectory, and imports it through the module path -
including the two flavours of test file. The second draws the line between
exported and unexported deliberately: a `bank.Account` whose balance is
unreachable from outside, getters, methods that refuse to move money when they
cannot, exported sentinel errors matched with `errors.Is`, and a test that uses
reflection to check no field slipped out into upper case.
