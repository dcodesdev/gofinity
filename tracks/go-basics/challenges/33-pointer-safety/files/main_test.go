package main

import "testing"

// mustAccount fails fast rather than letting an unfinished NewAccount turn
// every later test into a nil dereference.
func mustAccount(t *testing.T, owner string, balance int) *Account {
	t.Helper()
	a := NewAccount(owner, balance)
	if a == nil {
		t.Fatalf("NewAccount(%q, %d) returned nil", owner, balance)
	}
	return a
}

func TestNewAccount(t *testing.T) {
	a := NewAccount("ada", 10)
	if a == nil {
		t.Fatal("NewAccount returned nil")
	}
	if *a != (Account{Owner: "ada", Balance: 10}) {
		t.Errorf("NewAccount = %+v, want {ada 10}", *a)
	}

	// Two calls must be two accounts, not two names for one.
	b := NewAccount("ada", 10)
	if a == b {
		t.Error("NewAccount returned the same pointer twice")
	}
	b.Deposit(1)
	if a.Balance != 10 {
		t.Errorf("depositing into the second account changed the first to %d", a.Balance)
	}
}

func TestDeposit(t *testing.T) {
	a := mustAccount(t, "ada", 10)
	if !a.Deposit(5) {
		t.Error("Deposit into a real account reported false, want true")
	}
	if a.Balance != 15 {
		t.Errorf("Balance = %d, want 15", a.Balance)
	}

	// A method with a pointer receiver still runs when the pointer is nil.
	// It only panics if it dereferences without looking.
	var missing *Account
	if missing.Deposit(5) {
		t.Error("Deposit into a nil account reported true, want false")
	}
}

func TestDescribe(t *testing.T) {
	if got := mustAccount(t, "ada", 10).Describe(); got != "ada: 10" {
		t.Errorf("Describe = %q, want %q", got, "ada: 10")
	}
	var missing *Account
	if got := missing.Describe(); got != "<no account>" {
		t.Errorf("Describe of a nil account = %q, want %q", got, "<no account>")
	}
	if got := (&Account{}).Describe(); got != ": 0" {
		t.Errorf("Describe of the zero account = %q, want %q", got, ": 0")
	}
}

func TestTopUp(t *testing.T) {
	accounts := []Account{{Owner: "ada", Balance: 10}, {Owner: "linus", Balance: 4}}
	TopUp(accounts, 5)
	if accounts[0].Balance != 15 || accounts[1].Balance != 9 {
		t.Errorf("TopUp left %+v, want balances 15 and 9 - a range copy is not the element", accounts)
	}
	TopUp(nil, 5) // must not panic
}

func TestRichestAliasesTheSlice(t *testing.T) {
	accounts := []Account{{Owner: "ada", Balance: 10}, {Owner: "linus", Balance: 40}, {Owner: "rob", Balance: 7}}
	got := Richest(accounts)
	if got == nil {
		t.Fatal("Richest returned nil for a non-empty slice")
	}
	if got.Owner != "linus" {
		t.Fatalf("Richest = %q, want %q", got.Owner, "linus")
	}

	// The whole point: the returned pointer is the slice element, so writing
	// through it is visible in the slice.
	got.Deposit(60)
	if accounts[1].Balance != 100 {
		t.Errorf("accounts[1].Balance = %d, want 100 - Richest returned a pointer to a copy", accounts[1].Balance)
	}
	if got != &accounts[1] {
		t.Error("Richest did not return the address of the slice element")
	}

	// A tie keeps the earlier account.
	tie := []Account{{Owner: "first", Balance: 5}, {Owner: "second", Balance: 5}}
	if got := Richest(tie); got.Owner != "first" {
		t.Errorf("Richest on a tie = %q, want the earlier %q", got.Owner, "first")
	}

	if got := Richest(nil); got != nil {
		t.Errorf("Richest(nil) = %+v, want nil", got)
	}
	if got := Richest([]Account{}); got != nil {
		t.Errorf("Richest(empty) = %+v, want nil", got)
	}
}

func TestRichestWithNegativeBalances(t *testing.T) {
	// Starting the search from a zero Account rather than accounts[0] would
	// find nothing here.
	accounts := []Account{{Owner: "ada", Balance: -9}, {Owner: "linus", Balance: -2}}
	got := Richest(accounts)
	if got == nil || got.Owner != "linus" {
		t.Errorf("Richest = %+v, want linus", got)
	}
}

func TestTotal(t *testing.T) {
	accounts := []*Account{mustAccount(t, "ada", 10), nil, mustAccount(t, "linus", 4)}
	if got := Total(accounts); got != 14 {
		t.Errorf("Total = %d, want 14 - nil entries are skipped, not dereferenced", got)
	}
	if got := Total(nil); got != 0 {
		t.Errorf("Total(nil) = %d, want 0", got)
	}
	if got := Total([]*Account{nil, nil}); got != 0 {
		t.Errorf("Total of only nils = %d, want 0", got)
	}
}

func TestAddresses(t *testing.T) {
	accounts := []Account{{Owner: "ada", Balance: 10}, {Owner: "linus", Balance: 4}}
	ptrs := Addresses(accounts)
	if len(ptrs) != 2 {
		t.Fatalf("Addresses returned %d pointers, want 2", len(ptrs))
	}

	// Every pointer is distinct, and each one aliases its own element.
	if ptrs[0] == ptrs[1] {
		t.Fatal("Addresses returned the same pointer twice")
	}
	for i, p := range ptrs {
		if p != &accounts[i] {
			t.Errorf("ptrs[%d] does not point at accounts[%d]", i, i)
		}
	}

	ptrs[1].Deposit(6)
	if accounts[1].Balance != 10 {
		t.Errorf("accounts[1].Balance = %d, want 10 - the pointers must alias the slice", accounts[1].Balance)
	}

	if got := Addresses(nil); len(got) != 0 {
		t.Errorf("Addresses(nil) = %v, want nothing", got)
	}
}

func TestClone(t *testing.T) {
	original := mustAccount(t, "ada", 10)
	copied := Clone(original)
	if copied == nil {
		t.Fatal("Clone returned nil for a real account")
	}
	if copied == original {
		t.Fatal("Clone returned the same pointer - that is an alias, not a copy")
	}
	if *copied != *original {
		t.Errorf("Clone = %+v, want %+v", *copied, *original)
	}

	copied.Deposit(90)
	if original.Balance != 10 {
		t.Errorf("writing through the clone changed the original to %d", original.Balance)
	}

	if got := Clone(nil); got != nil {
		t.Errorf("Clone(nil) = %+v, want nil", got)
	}
}

func TestBalanceOr(t *testing.T) {
	if got := BalanceOr(mustAccount(t, "ada", 10), -1); got != 10 {
		t.Errorf("BalanceOr = %d, want 10", got)
	}
	var missing *Account
	if got := BalanceOr(missing, -1); got != -1 {
		t.Errorf("BalanceOr(nil) = %d, want the fallback -1", got)
	}
	if got := BalanceOr(&Account{}, 7); got != 0 {
		t.Errorf("BalanceOr of a zero account = %d, want 0 - a real account with balance 0 is not a missing one", got)
	}
}
