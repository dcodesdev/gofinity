# Exported Identifiers

Go has no `public` or `private` keyword. Visibility is spelling: a name whose
first letter is upper case is **exported** and visible to every package that
imports yours, and a name that starts lower case is visible only inside the
package that declares it.

```go
func Title(s string) string   // callers can use this
func isWordChar(r rune) bool  // only this package can
```

The rule applies to everything with a name - functions, types, methods, struct
fields, constants, variables - and it is checked at compile time, not by
convention. `bank.normalizeOwner` is not a lint warning, it is a build error.

## Why a field would ever be hidden

An exported field is a standing promise that anybody may set it to anything, at
any time, without going through your code:

```go
account.Balance = 1_000_000   // if Balance were exported, this is legal
```

Once the field is lower case, the only way in is the method set you wrote, so
"a withdrawal larger than the balance is refused" is a rule the type can
actually keep. Exporting a getter costs one line and gives up nothing:

```go
func (a *Account) Balance() int { return a.balance }
```

The reverse is just as real. Anything you export is API: someone will depend on
it, and changing it later breaks them. Export the smallest set that lets a
caller do the job.

## Sentinel errors are exported on purpose

`ErrInsufficientFunds` is a package-level `var` created with
[`errors.New`](https://pkg.go.dev/errors#New), and it is exported precisely so
callers can recognise it:

```go
if errors.Is(err, bank.ErrInsufficientFunds) { ... }
```

Comparing `err.Error()` against a string would work today and break the moment
somebody rewords the message. The value is the contract; the text is not.

## Two test files, two packages

`bank/bank_test.go` says `package bank`, so it is inside and can call
`normalizeOwner` and read `a.balance` directly. `main_test.go` says
`package main`: it goes through the exported surface like any other caller, and
it uses [`reflect`](https://pkg.go.dev/reflect) to assert that `Account` has no
exported field - the visibility rule, checked mechanically rather than trusted.

## Task

Finish the `bank` package - the hidden `balance` field, the constructor, the
getters, `Deposit`, `Withdraw`, `Transfer` and `normalizeOwner` - then write
`FormatCents`, `Report` and `Describe` in `main` against nothing but its
exported names.

## Hints

- `Account` needs a second field, `balance int`, and it stays lower case.
  Capitalising it fails the reflect test in `main_test.go`.
- A failed operation must not move money. Check every condition before you
  touch the balance, and return the error instead.
- `Transfer` is the same rule one level up: call `Withdraw` first and return
  early if it fails, so a rejected transfer never credits the destination.
- Zero is a valid opening balance but not a valid deposit or withdrawal. Read
  each doc comment for which comparison it wants.
- `Describe` orders its cases: `nil` first, then
  [`errors.Is`](https://pkg.go.dev/errors#Is) for each sentinel, then
  `err.Error()` for anything else. A `switch` with no expression reads
  better than a chain of `if`s.
- `FormatCents` is integer division and remainder, and the cents need two
  digits: [`fmt.Sprintf("$%d.%02d", ...)`](https://pkg.go.dev/fmt#Sprintf).
