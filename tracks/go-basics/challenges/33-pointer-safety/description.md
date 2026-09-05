# Pointer Safety

Go has pointers but no pointer arithmetic, and it collects garbage, so most of
the ways a pointer can hurt you in C are gone. Three things are left, and this
challenge is all three.

## Returning the address of a local

```go
func NewAccount(owner string, balance int) *Account {
	return &Account{Owner: owner, Balance: balance}
}
```

In C that is a dangling pointer. In Go it is the standard way to write a
constructor. The compiler runs an escape analysis, sees the address outlive the
function, and allocates the value on the heap instead of the stack. You never
say which, and you cannot get it wrong.

## A nil pointer is a usable receiver

Calling a method on a nil pointer does **not** panic by itself. The method runs
with a nil receiver, and only panics if it goes on to read a field:

```go
func (a *Account) Deposit(n int) bool {
	if a == nil {
		return false
	}
	a.Balance += n
	return true
}
```

So a pointer-receiver method can define what "no value" means. That is how
`(*T).String` methods stay printable and how a nil `*Node` can be a valid empty
tree. The rule is only that the method must check before it dereferences.

Note the difference from a **value** receiver: there is no such thing as a nil
`Account`, so a value-receiver method can never be in this position.

## Aliasing: a pointer into a slice

`&accounts[i]` is the address of the element itself, not of a copy. Returning
one hands the caller a live view of the slice:

```go
p := Richest(accounts)
p.Deposit(100)      // accounts[i] changed too
```

That is deliberate here, and it is also the sharpest edge in the exercise. A
pointer into a slice keeps the whole backing array alive, and `append` may
reallocate that array, after which the pointer refers to the old one and the
two silently stop agreeing. Take pointers into a slice you are not going to
grow.

`Clone` is the opposite move: `copied := *a` dereferences into a new value, and
`&copied` is a pointer nobody else holds. Note that this is a **shallow** copy -
it copies each field, so a pointer field would still be shared.

## Task

Fill in `main.go`. `NewAccount` and `Clone` make new pointers; `Deposit`,
`Describe` and `BalanceOr` handle nil; `TopUp`, `Richest` and `Addresses` work
through the slice rather than through copies; `Total` skips nils.

## Hints

- `Richest` must return `&accounts[best]`. Building a local `best := accounts[0]`
  and returning `&best` returns the address of a copy, and the aliasing test
  will say so.
- It also cannot start from a zero `Account`: balances can be negative, so the
  running best has to start at `accounts[0]` after a length check. Compare with
  `>` so a tie keeps the earlier account.
- `Addresses` needs `for i := range accounts` and `&accounts[i]`. Under Go 1.22
  and later `for _, a := range` gives a fresh `a` each iteration, so `&a` yields
  distinct pointers - but they point at copies, not at the slice, which the test
  checks.
- `Describe` on the zero account is `": 0"`, not `"<no account>"`. Only a nil
  pointer is missing; an account with an empty name and no money is a real one.
- `BalanceOr` follows the same line: `&Account{}` returns 0, not the fallback.
