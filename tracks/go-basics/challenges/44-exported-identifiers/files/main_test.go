package main

import (
	"errors"
	"reflect"
	"testing"

	"gofinity/exported/bank"
)

func newAccount(t *testing.T, owner string, opening int) *bank.Account {
	t.Helper()
	a, err := bank.NewAccount(owner, opening)
	if err != nil {
		t.Fatalf("bank.NewAccount(%q, %d) returned %v, want nil", owner, opening, err)
	}
	if a == nil {
		t.Fatalf("bank.NewAccount(%q, %d) returned a nil account", owner, opening)
	}
	return a
}

// The visibility rule, checked mechanically: from out here an exported field
// would be assignable by anybody, so bank.Account must not have one.
func TestAccountFieldsAreUnexported(t *testing.T) {
	typ := reflect.TypeOf(bank.Account{})
	names := make([]string, 0, typ.NumField())
	for i := range typ.NumField() {
		field := typ.Field(i)
		names = append(names, field.Name)
		if field.IsExported() {
			t.Errorf("Account.%s is exported - a balance is not something callers may assign", field.Name)
		}
	}
	want := map[string]bool{"owner": true, "balance": true}
	for name := range want {
		found := false
		for _, got := range names {
			if got == name {
				found = true
			}
		}
		if !found {
			t.Errorf("Account has no %q field, got %v", name, names)
		}
	}
}

func TestFormatCents(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "$0.00"},
		{5, "$0.05"},
		{99, "$0.99"},
		{100, "$1.00"},
		{1234, "$12.34"},
		{100000, "$1000.00"},
	}
	for _, c := range cases {
		if got := FormatCents(c.in); got != c.want {
			t.Errorf("FormatCents(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestReport(t *testing.T) {
	a := newAccount(t, "  ada   lovelace ", 1234)
	if got, want := Report(a), "ada lovelace: $12.34"; got != want {
		t.Errorf("Report = %q, want %q", got, want)
	}
}

func TestDescribe(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"nil", nil, "ok"},
		{"insufficient", bank.ErrInsufficientFunds, "not enough money"},
		{"invalid", bank.ErrInvalidAmount, "that is not a valid amount"},
		{"other", errors.New("boom"), "boom"},
	}
	for _, c := range cases {
		if got := Describe(c.err); got != c.want {
			t.Errorf("Describe(%s) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestNewAccountRejectsANegativeOpeningBalance(t *testing.T) {
	if _, err := bank.NewAccount("ada", -1); !errors.Is(err, bank.ErrInvalidAmount) {
		t.Errorf("NewAccount with -1 returned %v, want ErrInvalidAmount", err)
	}
	if _, err := bank.NewAccount("ada", 0); err != nil {
		t.Errorf("NewAccount with 0 returned %v, want nil - an empty account is fine", err)
	}
}

func TestDepositAndWithdraw(t *testing.T) {
	a := newAccount(t, "ada", 100)
	if err := a.Deposit(50); err != nil {
		t.Fatalf("Deposit(50) returned %v, want nil", err)
	}
	if got := a.Balance(); got != 150 {
		t.Errorf("Balance after Deposit(50) = %d, want 150", got)
	}
	if err := a.Withdraw(30); err != nil {
		t.Fatalf("Withdraw(30) returned %v, want nil", err)
	}
	if got := a.Balance(); got != 120 {
		t.Errorf("Balance after Withdraw(30) = %d, want 120", got)
	}
}

func TestAFailedOperationLeavesTheBalanceAlone(t *testing.T) {
	a := newAccount(t, "ada", 100)
	if err := a.Withdraw(101); !errors.Is(err, bank.ErrInsufficientFunds) {
		t.Errorf("Withdraw(101) returned %v, want ErrInsufficientFunds", err)
	}
	if err := a.Deposit(0); !errors.Is(err, bank.ErrInvalidAmount) {
		t.Errorf("Deposit(0) returned %v, want ErrInvalidAmount", err)
	}
	if err := a.Withdraw(-5); !errors.Is(err, bank.ErrInvalidAmount) {
		t.Errorf("Withdraw(-5) returned %v, want ErrInvalidAmount", err)
	}
	if got := a.Balance(); got != 100 {
		t.Errorf("Balance after three failed calls = %d, want 100", got)
	}
	if got := a.Owner(); got != "ada" {
		t.Errorf("Owner = %q, want %q", got, "ada")
	}
}

func TestWithdrawingTheWholeBalanceIsAllowed(t *testing.T) {
	a := newAccount(t, "ada", 100)
	if err := a.Withdraw(100); err != nil {
		t.Fatalf("Withdraw(100) of a 100 balance returned %v, want nil", err)
	}
	if got := a.Balance(); got != 0 {
		t.Errorf("Balance = %d, want 0", got)
	}
}

func TestTransfer(t *testing.T) {
	from := newAccount(t, "ada", 500)
	to := newAccount(t, "grace", 100)

	if err := from.Transfer(to, 200); err != nil {
		t.Fatalf("Transfer(200) returned %v, want nil", err)
	}
	if got := from.Balance(); got != 300 {
		t.Errorf("source balance = %d, want 300", got)
	}
	if got := to.Balance(); got != 300 {
		t.Errorf("destination balance = %d, want 300", got)
	}

	if err := from.Transfer(to, 10_000); !errors.Is(err, bank.ErrInsufficientFunds) {
		t.Errorf("Transfer of more than the balance returned %v, want ErrInsufficientFunds", err)
	}
	if got, want := from.Balance(), 300; got != want {
		t.Errorf("source balance after a failed transfer = %d, want %d", got, want)
	}
	if got, want := to.Balance(), 300; got != want {
		t.Errorf("destination balance after a failed transfer = %d, want %d - nothing may arrive", got, want)
	}
}
