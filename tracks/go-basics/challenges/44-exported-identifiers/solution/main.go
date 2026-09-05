package main

import (
	"errors"
	"fmt"

	"gofinity/exported/bank"
)

// FormatCents renders a balance in cents as money.
func FormatCents(cents int) string {
	return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
}

// Report describes an account in one line.
func Report(a *bank.Account) string {
	return fmt.Sprintf("%s: %s", a.Owner(), FormatCents(a.Balance()))
}

// Describe turns an error from the bank package into a sentence for a person.
func Describe(err error) string {
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, bank.ErrInsufficientFunds):
		return "not enough money"
	case errors.Is(err, bank.ErrInvalidAmount):
		return "that is not a valid amount"
	default:
		return err.Error()
	}
}

func main() {
	account, err := bank.NewAccount("  ada   lovelace ", 1000)
	if err != nil {
		fmt.Println(Describe(err))
		return
	}
	fmt.Println(Report(account))
	fmt.Println(Describe(account.Withdraw(5000)))
}
