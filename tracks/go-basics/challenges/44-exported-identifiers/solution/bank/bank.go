// Package bank models an account whose balance can only move through the
// methods this package exports.
package bank

import (
	"errors"
	"strings"
)

// ErrInsufficientFunds is returned by Withdraw when the account does not hold
// enough.
var ErrInsufficientFunds = errors.New("bank: insufficient funds")

// ErrInvalidAmount is returned when an amount is zero or negative.
var ErrInvalidAmount = errors.New("bank: amount must be positive")

// Account holds an owner's balance in whole cents. Both fields are unexported,
// so the balance can only move through the methods below.
type Account struct {
	owner   string
	balance int
}

// NewAccount returns an account for owner with an opening balance.
func NewAccount(owner string, opening int) (*Account, error) {
	if opening < 0 {
		return nil, ErrInvalidAmount
	}
	return &Account{owner: normalizeOwner(owner), balance: opening}, nil
}

// Owner returns the account owner's name.
func (a *Account) Owner() string {
	return a.owner
}

// Balance returns the current balance in cents.
func (a *Account) Balance() int {
	return a.balance
}

// Deposit adds amount to the balance.
func (a *Account) Deposit(amount int) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	a.balance += amount
	return nil
}

// Withdraw removes amount from the balance.
func (a *Account) Withdraw(amount int) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if amount > a.balance {
		return ErrInsufficientFunds
	}
	a.balance -= amount
	return nil
}

// Transfer moves amount from a to dst, and does nothing at all when it cannot.
func (a *Account) Transfer(dst *Account, amount int) error {
	if err := a.Withdraw(amount); err != nil {
		return err
	}
	return dst.Deposit(amount)
}

// normalizeOwner trims the surrounding whitespace from name and collapses every
// internal run of whitespace to a single space.
func normalizeOwner(name string) string {
	return strings.Join(strings.Fields(name), " ")
}
