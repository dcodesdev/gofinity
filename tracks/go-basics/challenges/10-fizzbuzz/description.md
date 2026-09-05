# FizzBuzz

The smallest interesting program in the world, and the reason it survives as an
exercise: it needs three branches, and the order you write them in decides
whether it is right.

Go's `if` takes a `bool` and nothing else. There is no truthiness, so `if n` is
a compile error and `if n != 0` is what you meant. The condition needs no
parentheses, and the braces are never optional:

```go
if n%3 == 0 {
	return "Fizz"
} else if n%5 == 0 {
	return "Buzz"
}
```

`%` is the remainder operator. In Go the result takes the sign of the *left*
operand, so `-7 % 3` is `-1` and `7 % -3` is `1`. That only matters here because
`n % 3 == 0` is true for negative multiples too, which is what the tests expect.

## Task

Finish the three functions in `main.go`.

1. `FizzBuzz(n int) string` classifies one number: `"FizzBuzz"` when it divides
   by both 3 and 5, `"Fizz"` for 3, `"Buzz"` for 5, and otherwise the number's
   digits as a string. `FizzBuzz(7)` is `"7"`.
2. `FizzBuzzUpTo(n int) []string` returns `FizzBuzz(1)` through `FizzBuzz(n)` in
   order. An `n` below 1 gives a slice with no elements.
3. `IsLeapYear(y int) bool` applies the Gregorian rule: divisible by 4, except
   divisible by 100, except divisible by 400. `2000` is a leap year and `1900`
   is not.

## Hints

- Check the combined case first. `15 % 3` is `0`, so a `Fizz` branch above the
  `FizzBuzz` branch answers `"Fizz"` and the right answer never gets a turn.
  `n%15 == 0` and `n%3 == 0 && n%5 == 0` are the same test; pick the one you
  find clearer.
- Turning an `int` into its digits is
  [`strconv.Itoa(n)`](https://pkg.go.dev/strconv#Itoa). `string(n)` compiles
  and is wrong: it produces the character with that code point. `fmt.Sprint(n)`
  works too and costs more.
- `strconv` is not imported yet. Add it; an unused import is a compile error,
  which is why the starter does not carry one you might not need.
- Build the slice with `append` on a starting `out := []string{}`. The tests
  only check the length, so `nil` would pass too, but an empty non-nil slice is
  the friendlier thing to hand back.
- The leap-year rule is three tests in a specific order: 400 first, then 100,
  then 4. Written in any other order it is either wrong or much longer.
