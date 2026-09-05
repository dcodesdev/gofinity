package main

import "fmt"

// Account is a named balance. The tests hold both Account values and
// *Account pointers, so watch which one each function is given.
type Account struct {
	Owner   string
	Balance int
}

// NewAccount returns a pointer to a new Account. Returning the address of a
// local is legal in Go and always safe: the compiler sees the pointer escape
// and puts the value on the heap.
func NewAccount(owner string, balance int) *Account {
	// TODO
	return nil
}

// Deposit adds n to the balance and reports whether it did. A nil *Account is
// a no-op returning false rather than a panic: the method still runs, the
// receiver is just nil, so it can check.
func (a *Account) Deposit(n int) bool {
	// TODO
	return false
}

// Describe renders "Owner: N" for a real account and "<no account>" for a nil
// one.
func (a *Account) Describe() string {
	// TODO
	return ""
}

// TopUp adds n to every account in the slice, in place.
func TopUp(accounts []Account, n int) {
	// TODO
}

// Richest returns a pointer to the account with the highest balance, or nil
// when there are none. The pointer must alias the slice element itself, so
// depositing through it changes the slice.
func Richest(accounts []Account) *Account {
	// TODO
	return nil
}

// Total sums the balances of a slice of pointers, skipping nil entries.
func Total(accounts []*Account) int {
	// TODO
	return 0
}

// Addresses returns one pointer per account in the slice, each aliasing the
// element it came from.
func Addresses(accounts []Account) []*Account {
	// TODO
	return nil
}

// Clone returns a pointer to an independent copy, so writing through the copy
// leaves the original alone. A nil account clones to nil.
func Clone(a *Account) *Account {
	// TODO
	return nil
}

// BalanceOr returns the balance, or fallback when the pointer is nil. This is
// the read-side counterpart of the nil check inside Deposit.
func BalanceOr(a *Account, fallback int) int {
	// TODO
	return 0
}

func main() {
	accounts := []Account{{Owner: "ada", Balance: 10}, {Owner: "linus", Balance: 4}}
	TopUp(accounts, 5)
	Richest(accounts).Deposit(100)
	for _, a := range accounts {
		fmt.Println(a.Describe())
	}
	var missing *Account
	fmt.Println(missing.Describe(), BalanceOr(missing, -1))
}
