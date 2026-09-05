# Type Switch

An interface value carries two things: a concrete type and a value of that
type. A **type switch** is how you get them back out.

```go
switch v := v.(type) {
case int:
	// v is an int here
case string:
	// v is a string here
case nil:
	// the interface held nothing at all
default:
	// v is still the original interface
}
```

The `v :=` shadows the parameter deliberately, so each case works with the
narrow type and the name stays the same. In `default` there is nothing to
narrow to, so `v` keeps its interface type - which is what `%T` and `%v` want
anyway.

## Order decides

Cases are tried top to bottom, and a case naming an **interface** matches
anything satisfying it. So this order:

```go
case fmt.Stringer:
case string:
```

is a bug if any of your strings have a `String` method: the first case swallows
them. Put concrete types first and interfaces last, exactly like a `switch` on
values puts its narrow conditions before its broad ones.

A named type is its own type. `type Celsius float64` does **not** match
`case float64`; it matches `case Celsius`, and it matches
[`case fmt.Stringer`](https://pkg.go.dev/fmt#Stringer) if it has the method.
There is no implicit widening anywhere in Go, and this is the same rule showing
up again.

## One case, several types

```go
case int, int64:
	// v is still `any` here
```

When a case lists more than one type there is no single type to narrow to, so
`v` stays as the interface. Write the branches separately when each needs its
own conversion.

## Assertion or switch

For one type, an assertion is shorter and says more:

```go
n, ok := v.(int)
```

Reach for a type switch when there are several answers to distinguish. Reach
for neither when an ordinary interface method would do - a type switch over
your own types tends to mean the behaviour belongs on the types as a method.

## Task

Implement `String` on `Celsius`, then `Describe`, `SumNumbers`, `AsInt` and
`JoinStrings`.

## Hints

- `Describe`'s `case nil` matches an interface holding nothing. It cannot be
  written as `if v == nil` inside another case.
- `%q` quotes a string, `%t` prints a bool, `%.2f` gives two digits, and `%T`
  prints the type name - `main.Point`, with the package qualifier.
- Put `case fmt.Stringer` after the concrete cases, or `Celsius` never reaches
  it... and check the reverse too: a plain `float64` must not reach it.
- `AsInt` is one line: `n, ok := v.(int)`. An `int64` holding 7 is not an `int`.
- `JoinStrings` can collect into a `[]string` and finish with
  [`strings.Join`](https://pkg.go.dev/strings#Join), which handles the separator
  for you.
