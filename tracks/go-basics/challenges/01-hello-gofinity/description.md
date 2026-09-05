# Hello, Gofinity

Every Go program starts the same way: a package declaration, some imports, and
functions. Executable programs live in `package main` and start at `func main`.

Your job is to fill in `Greet`, which takes a name and returns a greeting.

## Task

Implement `Greet(name string) string` in `main.go` so that it returns:

```
Hello, <name>!
```

For example, `Greet("Gofinity")` must return `Hello, Gofinity!`.

If `name` is empty, greet the world instead - `Greet("")` must return
`Hello, World!`.

## Hints

- `fmt.Sprintf` formats a string and returns it, rather than printing it:
  `fmt.Sprintf("Hello, %s!", name)`.
- A `string` compares to the empty string with `==`, and `len(name) == 0` works
  just as well.
- Run the tests with the Run button. `main_test.go` is read-only - the same
  tests run when you submit.
