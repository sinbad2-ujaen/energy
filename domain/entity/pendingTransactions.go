package entity

import (
	"sort"
	"sync"
)

type PendingTransactions struct {
	Mu           sync.RWMutex
	Transactions map[string]Transaction
}

func (ws *PendingTransactions) AddTransaction(transaction Transaction) {
	ws.Mu.Lock()
	defer ws.Mu.Unlock()
	ws.Transactions[transaction.ID] = transaction
}

func (ws *PendingTransactions) GetTransaction(id string) (Transaction, bool) {
	ws.Mu.RLock()
	defer ws.Mu.RUnlock()
	transaction, ok := ws.Transactions[id]
	return transaction, ok
}

func (ws *PendingTransactions) DeleteTransaction(id string) {
	ws.Mu.Lock()
	defer ws.Mu.Unlock()

	delete(ws.Transactions, id)
}

func (pt *PendingTransactions) GetOldestTransactions() []Transaction {
	pt.Mu.RLock()
	defer pt.Mu.RUnlock()

	// Create a slice to store all transactions
	allTransactions := make([]Transaction, 0, len(pt.Transactions))

	// Populate the slice with all transactions
	for _, transaction := range pt.Transactions {
		allTransactions = append(allTransactions, transaction)
	}

	// Sort transactions by timestamp in ascending order
	sort.Slice(allTransactions, func(i, j int) bool {
		return allTransactions[i].TimestampCreated.Before(allTransactions[j].TimestampCreated)
	})

	// Return the oldest two transactions
	if len(allTransactions) >= 2 {
		return allTransactions[:2]
	}

	// If there are fewer than two transactions, return all transactions
	return allTransactions
}

func (ws *PendingTransactions) InitPendingTransactions() {
	ws.AddTransaction(NewTransaction(
		"9a27ad25d050b690b05d38e1cf20c71e8c5314cff5a936efc024a2b5e9b07f04",
		"8501df062b55e6f938cf5c2c36849e8c11663f8f79e28bd1a99431d825792a44",
		0.01,
		"test",
		TransactionTypeStandard,
		0.0,
		304,
	))

	ws.AddTransaction(NewTransaction(
		"9a27ad25d050b690b05d38e1cf20c71e8c5314cff5a936efc024a2b5e9b07f04",
		"dc22da8161544f7af3f949148fee1d6486b5ef3fcd339b0226384b11afb79f2e",
		0.01,
		"test",
		TransactionTypeStandard,
		0.0,
		304,
	))

	ws.AddTransaction(NewTransaction(
		"9a27ad25d050b690b05d38e1cf20c71e8c5314cff5a936efc024a2b5e9b07f04",
		"34030f5e884b81808993def07ebe442ec22c610bdf98d566fb56674ea57f6953",
		0.01,
		"test",
		TransactionTypeFast,
		1.0,
		304,
	))

	ws.AddTransaction(NewTransaction(
		"9a27ad25d050b690b05d38e1cf20c71e8c5314cff5a936efc024a2b5e9b07f04",
		"f929aa0e806109b35f9cc2ed8483f8ce0db9b08c7e18ca7d9fbb3d9238dc9b89",
		0.01,
		"test",
		TransactionTypeFast,
		1.0,
		304,
	))
}
