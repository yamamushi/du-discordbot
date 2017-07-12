package main

import (
	"errors"
)

// Wallet struct
type Wallet struct {
	Account string `storm:"id"`
	Pin     string `storm:"index"`
	Balance int    `storm:"index"`
}

// GetBalance function
func (w *Wallet) GetBalance() int {
	return w.Balance
}

// AddBalance function
func (w *Wallet) AddBalance(amount int) {
	w.Balance = w.Balance + amount
}

// RemoveBalance function
func (w *Wallet) RemoveBalance(amount int) {
	w.Balance = w.Balance - amount
}

// SendBalance function
func (w *Wallet) SendBalance(receiver *Wallet, amount int) error {

	if w.GetBalance()-amount < 0 {
		return errors.New("Insufficient Sender Balance")
	}

	if amount < 0 {
		return errors.New("You cannot send a negative amount")
	}

	w.Balance = w.Balance - amount
	receiver.Balance = receiver.Balance + amount
	return nil
}
