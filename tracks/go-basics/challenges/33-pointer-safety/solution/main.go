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
	return &Account{Owner: owner, Balance: balance}
}

// Deposit adds n to the balance and reports whether it did. A nil *Account is
// a no-op returning false rather than a panic: the method still runs, the
// receiver is just nil, so it can check.
func (a *Account) Deposit(n int) bool {
	if a == nil {
		return false
	}
	a.Balance += n
	return true
}

// Describe renders "Owner: N" for a real account and "<no account>" for a nil
// one.
func (a *Account) Describe() string {
	if a == nil {
		return "<no account>"
	}
	return fmt.Sprintf("%s: %d", a.Owner, a.Balance)
}

// TopUp adds n to every account in the slice, in place.
func TopUp(accounts []Account, n int) {
	for i := range accounts {
		accounts[i].Balance += n
	}
}

// Richest returns a pointer to the account with the highest balance, or nil
// when there are none. The pointer must alias the slice element itself, so
// depositing through it changes the slice.
func Richest(accounts []Account) *Account {
	if len(accounts) == 0 {
		return nil
	}
	best := 0
	for i := range accounts[1:] {
		// i counts from 0 over the tail, so the element is i+1.
		if accounts[i+1].Balance > accounts[best].Balance {
			best = i + 1
		}
	}
	// &accounts[best], not &copyOfTheElement: the caller expects an alias.
	return &accounts[best]
}

// Total sums the balances of a slice of pointers, skipping nil entries.
func Total(accounts []*Account) int {
	total := 0
	for _, a := range accounts {
		if a == nil {
			continue
		}
		total += a.Balance
	}
	return total
}

// Addresses returns one pointer per account in the slice, each aliasing the
// element it came from.
func Addresses(accounts []Account) []*Account {
	out := []*Account{}
	for i := range accounts {
		out = append(out, &accounts[i])
	}
	return out
}

// Clone returns a pointer to an independent copy, so writing through the copy
// leaves the original alone. A nil account clones to nil.
func Clone(a *Account) *Account {
	if a == nil {
		return nil
	}
	// *a is the value, and a struct value copies field by field.
	copied := *a
	return &copied
}

// BalanceOr returns the balance, or fallback when the pointer is nil. This is
// the read-side counterpart of the nil check inside Deposit.
func BalanceOr(a *Account, fallback int) int {
	if a == nil {
		return fallback
	}
	return a.Balance
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
