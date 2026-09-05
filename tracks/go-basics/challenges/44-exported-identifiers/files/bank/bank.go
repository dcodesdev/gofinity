// Package bank models an account whose balance can only move through the
// methods this package exports.
package bank

import "errors"

// ErrInsufficientFunds is returned by Withdraw when the account does not hold
// enough. It is a package-level var so callers can compare against it with
// errors.Is; the Err prefix is the convention every standard library package
// follows.
var ErrInsufficientFunds = errors.New("bank: insufficient funds")

// ErrInvalidAmount is returned when an amount is zero or negative.
var ErrInvalidAmount = errors.New("bank: amount must be positive")

// Account holds an owner's balance in whole cents.
//
// Both fields are lower case, so no code outside this package can read or write
// them. That is the point: an exported field is a promise that anybody may set
// it to anything, and a balance is not that kind of value.
//
// TODO: add the `balance int` field. Do NOT capitalise either field.
type Account struct {
	owner string
}

// NewAccount returns an account for owner with an opening balance. A negative
// opening balance is refused with ErrInvalidAmount; zero is fine.
//
// The owner's name is normalised on the way in, so every account stores it the
// same way no matter how the caller typed it.
//
// TODO: validate, normalise the name, and return the account.
func NewAccount(owner string, opening int) (*Account, error) {
	return nil, nil
}

// Owner returns the account owner's name. A getter is how a package exposes a
// value it will not let you assign.
//
// TODO
func (a *Account) Owner() string {
	return ""
}

// Balance returns the current balance in cents.
//
// TODO
func (a *Account) Balance() int {
	return 0
}

// Deposit adds amount to the balance. A non-positive amount is ErrInvalidAmount
// and leaves the balance untouched.
//
// TODO
func (a *Account) Deposit(amount int) error {
	return nil
}

// Withdraw removes amount from the balance. A non-positive amount is
// ErrInvalidAmount; more than the balance is ErrInsufficientFunds. Neither
// changes the balance: a method that fails should leave nothing behind.
//
// TODO
func (a *Account) Withdraw(amount int) error {
	return nil
}

// Transfer moves amount from a to dst, and does nothing at all when it cannot.
//
// TODO: withdraw first, and only deposit when that succeeded.
func (a *Account) Transfer(dst *Account, amount int) error {
	return nil
}

// normalizeOwner trims the surrounding whitespace from name and collapses every
// internal run of whitespace to a single space.
//
// It is unexported because it is an implementation detail: callers get the
// behaviour through NewAccount, and the day this rule changes no other package
// has to be recompiled against a promise we never made.
//
// TODO: strings.Fields plus strings.Join is the whole function.
func normalizeOwner(name string) string {
	return name
}
