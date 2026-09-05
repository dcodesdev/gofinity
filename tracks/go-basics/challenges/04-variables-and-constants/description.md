# Variables and Constants

Go has three ways to introduce a name, and they are not interchangeable.

- `const` fixes a value at compile time. It can never be reassigned, and it
  costs nothing at run time.
- `var` declares a variable, with or without an initial value. Without one it
  gets that type's zero value.
- `:=` declares *and* assigns in one step, inferring the type. It only works
  inside a function.

Package-level names must use `const` or `var`; `:=` is function-only.

## Task

Open `main.go` and finish three things.

1. `AppName` is a constant string. Give it the value `"Gofinity"`.
2. `MaxRetries` is a constant integer. Give it the value `3`.
3. Implement `RetryStatus(used int) string` and `Remaining(used int) int`.

`RetryStatus` returns one sentence, built from the two constants:

```
Gofinity: 2 of 3 retries used, 1 remaining
```

`Remaining` returns just the number of retries left.

The rules:

- A negative `used` counts as `0`. `RetryStatus(-4)` reports `0 of 3 ... 3
  remaining`.
- A `used` above `MaxRetries` counts as `MaxRetries`. The sentence must never
  claim more retries than exist, and `Remaining` must never go below `0`.
- Both functions read `MaxRetries` rather than a hard-coded `3`. Change the
  constant and every answer should move with it.

## Hints

- `const AppName = "Gofinity"` is an *untyped* constant. It has no type until
  it is used, so it slots into anything a string fits, and `MaxRetries` slots
  into anything an integer fits. That is why you rarely need to write
  `const MaxRetries int = 3`, though you may.
- Clamping reads well as a small unexported helper both functions call:

  ```go
  func clamp(used int) int {
  	if used < 0 {
  		return 0
  	}
  	...
  }
  ```

  Lowercase means package-private, which is right for something no caller
  outside this file needs.
- Inside a function, `spent := clamp(used)` is the idiomatic declaration. Save
  `var` for when you want the zero value or need the type spelled out.
- `fmt.Sprintf("%s: %d of %d ...", AppName, spent, MaxRetries, left)` builds the
  sentence. Count your verbs and your arguments: a mismatch shows up as
  `%!d(MISSING)` in the output rather than as a compile error.
