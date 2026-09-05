# Greeting Lines

One greeting was a warm-up. Now do it for a whole list, and number the lines.

Go has no string concatenation operator you would want to use in a loop. The
idiom is to build a slice of lines and then join them once, with
[`strings.Join`](https://pkg.go.dev/strings#Join).

## Task

Implement `GreetAll(names []string) string` in `main.go`. It returns one line
per name, joined by `"\n"`:

```
1. Hello, Ada!
2. Hello, Grace!
3. Hello, Alan!
```

The rules:

- Lines are numbered from `1`, not from `0`.
- An empty name is greeted as `World`, exactly like the last challenge:
  `[]string{"Ada", "", "Alan"}` gives `2. Hello, World!` in the middle.
- No names at all gives the empty string, `""`. A `nil` slice counts as no
  names.
- There is no trailing newline. `Join` puts the separator *between* elements,
  which is precisely what you want here.

## Hints

- `for i, name := range names` gives you the index and the value at once. The
  line number is `i + 1`.
- `fmt.Sprintf("%d. Hello, %s!", i+1, name)` formats one line. `%d` is an
  integer, `%s` a string.
- Collect the lines in a `[]string` and `append` to it, then
  `strings.Join(lines, "\n")`.
- `strings` is not imported yet. Add it, and note that Go refuses to compile a
  file that imports a package it never uses:

  ```go
  import (
  	"fmt"
  	"strings"
  )
  ```

- `strings.Join` of an empty slice is `""`, so the "no names" case needs no
  special handling.
