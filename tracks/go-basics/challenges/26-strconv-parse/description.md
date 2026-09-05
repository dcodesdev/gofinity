# Parsing with strconv

Everything that arrives from outside a program - a flag, an environment
variable, a form field, a line of CSV - arrives as a string. Turning it into a
number or a boolean is the [`strconv`](https://pkg.go.dev/strconv) package's
job, and every function in it returns **two** values: the parsed result and an
error.

```go
n, err := strconv.Atoi("42")             // 42, nil
f, err := strconv.ParseFloat("1.5", 64)  // 1.5, nil
b, err := strconv.ParseBool("true")      // true, nil
```

[`Atoi`](https://pkg.go.dev/strconv#Atoi) is shorthand for `ParseInt(s, 10, 0)`.
Going the other way is [`strconv.Itoa(n)`](https://pkg.go.dev/strconv#Itoa),
`FormatFloat`, `FormatBool`.

Two things catch people out. First, these functions are **strict**: no
surrounding whitespace, no trailing junk. `strconv.Atoi(" 42 ")` is an error, so
a [`strings.TrimSpace`](https://pkg.go.dev/strings#TrimSpace) usually comes
first. Second, the error is not a bare message. It is a
[`*strconv.NumError`](https://pkg.go.dev/strconv#NumError) holding the function
name, the input, and one of two sentinel errors:

```go
_, err := strconv.Atoi("abc")
errors.Is(err, strconv.ErrSyntax)   // true - it was not a number at all
_, err = strconv.Atoi("999999999999999999999999")
errors.Is(err, strconv.ErrRange)    // true - it was a number, but too big
```

That is why a parse error is worth passing along unchanged: rewrite it with your
own `fmt.Errorf` message and the caller can no longer tell "not a number" from
"too large".

The other habit worth building is the **`Or` wrapper**. Configuration usually
has a default, and

```go
func ParseIntOr(s string, fallback int) int
```

turns a two-value call into a one-value one at the single place that knows what
the default is.

## Task

Fill in the six functions in `main.go`.

1. `ParseIntOr`, `ParseFloatOr` and `ParseBoolOr` parse a trimmed `s` and fall
   back when it does not parse.
2. `SumFields(csv)` adds up comma-separated integers, skipping empty fields, and
   on the first bad field returns `0` and **strconv's own error**.
3. `ParseKeyValues(s)` reads `"a=1,b=2"` into a `map[string]int`.
4. `JoinInts(nums)` renders a slice back into `"1,2,3"`.

## Hints

- `strings.TrimSpace` before every parse, not after.
- Return the error from `strconv` directly. The test uses
  [`errors.Is(err, strconv.ErrSyntax)`](https://pkg.go.dev/errors#Is) and
  `errors.Is(err, strconv.ErrRange)`, so wrapping it in a message of your own
  with `%v` would break the link.
- [`strings.Cut(entry, "=")`](https://pkg.go.dev/strings#Cut) returns the part
  before, the part after, and whether the separator was there at all - a better
  fit than `strings.Split` when there is exactly one separator to find.
- `ParseKeyValues` must return a non-nil map for an input with no entries, and
  `nil` alongside an error.
- `strconv.Itoa` is the fast path for `int`.
  [`fmt.Sprintf("%d", n)`](https://pkg.go.dev/fmt#Sprintf) gives the same string
  and does much more work to get there.
