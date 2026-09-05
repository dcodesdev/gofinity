# Split into a Package

Every Go file starts with a package clause, and every directory is exactly one
package. That is the whole rule: a package is a directory, its name is the
`package` line its files share, and the way you reach it from elsewhere is an
import path built from the module.

This workspace is one module:

```
go.mod            module gofinity/splitpackage
main.go           package main
textkit/          package textkit
```

The `module` line in `go.mod` names the module, and an import path is that name
plus the directory below it:

```go
import "gofinity/splitpackage/textkit"
```

The last element of the path is the *directory*, not necessarily the package
name, though keeping the two the same is the convention and the kindest thing
you can do for a reader. After the import, the package name is what you type at
the call site: `textkit.Title(...)`.

## What crosses the boundary

Only the capitalised names. `Title`, `Slug` and `WordCount` are exported and
callable from `main`; `isWordChar` is not, and `textkit.isWordChar` does not
compile. Concept 15's second challenge is about designing that line - here just
notice that it exists.

## Two kinds of test file

`textkit/textkit_test.go` declares `package textkit`, so it is *inside* the
package and can call `isWordChar` directly. `main_test.go` declares
`package main`, so it sees only what `main` sees: the exported names it imported
and the functions in its own package.

`go test ./...` walks the module and runs both.

## Imports are grouped, not sorted by hand

[`gofmt`](https://pkg.go.dev/cmd/gofmt) sorts within a group; the convention is
standard library first, then a blank line, then everything else:

```go
import (
	"fmt"

	"gofinity/splitpackage/textkit"
)
```

An unused import is a compile error, not a warning. Go would rather you delete
it than let the build carry weight nobody reads.

## Task

Fill in `textkit.Title`, `textkit.Slug` and `textkit.WordCount`, then use them
from `main` to write `Headline` and `Summary`.

## Hints

- [`strings.Fields`](https://pkg.go.dev/strings#Fields) splits on any run of
  whitespace and never returns an empty field, which handles the leading,
  trailing and doubled spaces for free.
- Upper-casing the first *rune* means `[]rune(s)` and
  [`unicode.ToUpper`](https://pkg.go.dev/unicode#ToUpper), not `s[0]`: a byte
  is not a letter once the text stops being ASCII.
- Build the slug with a
  [`strings.Builder`](https://pkg.go.dev/strings#Builder). Remember whether you
  have skipped a character since the last one you kept, and write the `-`
  lazily when the next word character arrives - that way no separator is ever
  left dangling at either end.
- `Slug` lower-cases first, so `isWordChar` only ever sees lower-case letters
  and digits.
- `Summary` has three shapes and no arithmetic: a `switch` on the count reads
  better than nested `if`s.
