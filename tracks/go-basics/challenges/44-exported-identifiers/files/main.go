package main

import (
	"fmt"

	"gofinity/exported/bank"
)

// FormatCents renders a balance in cents as money: 1234 is "$12.34", 5 is
// "$0.05", 0 is "$0.00".
//
// TODO: integer division and remainder. The cents part is always two digits.
func FormatCents(cents int) string {
	return ""
}

// Report describes an account in one line: "Ada Lovelace: $12.34".
//
// It has to go through the getters. bank.Account keeps its fields unexported,
// so a.balance does not compile from here - and that is the package doing its
// job, not an obstacle.
//
// TODO
func Report(a *bank.Account) string {
	return ""
}

// Describe turns an error from the bank package into a sentence for a person.
//
//	nil                        "ok"
//	bank.ErrInsufficientFunds  "not enough money"
//	bank.ErrInvalidAmount      "that is not a valid amount"
//	anything else              the error's own message
//
// TODO: errors.Is against the two exported sentinels, in that order.
func Describe(err error) string {
	return ""
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
