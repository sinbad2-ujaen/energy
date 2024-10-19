package entity

import (
	"fmt"
	"sync"
)

type Wallet struct {
	Seed               string
	Balance            float64
	TransactionHistory []Transaction
	Mu                 sync.RWMutex // Mutex for thread safety
}

func NewWallet(seed string) *Wallet {
	return &Wallet{
		Seed: seed,
	}
}

func (w *Wallet) IncreaseBalance(token float64) {
	w.Mu.Lock()
	defer w.Mu.Unlock()

	w.Balance += token
}

func (w *Wallet) DecreaseBalance(token float64) error {
	w.Mu.Lock()
	defer w.Mu.Unlock()

	if w.Balance >= token {
		w.Balance -= token
	} else {
		return fmt.Errorf("insufficient balance")
	}

	return nil
}

func (w *Wallet) AddTransactionToHistory(transaction *Transaction) {
	w.Mu.Lock()
	defer w.Mu.Unlock()

	w.TransactionHistory = append(w.TransactionHistory, *transaction)
}
