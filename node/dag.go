package node

import (
	"energy/domain/entity"
	"fmt"
	"sync" // Import the sync package
)

type DAG struct {
	Transactions map[string]entity.Transaction
	mu           sync.Mutex // Add a mutex field
}

var dag DAG

func initDAG() {
	dag = DAG{Transactions: make(map[string]entity.Transaction)}

	// Origin transaction
	dag.addTransaction(
		entity.NewTransaction(
			"9a27ad25d050b690b05d38e1cf20c71e8c5314cff5a936efc024a2b5e9b07f04",
			"9a27ad25d050b690b05d38e1cf20c71e8c5314cff5a936efc024a2b5e9b07f04",
			100000000,
			"",
			entity.TransactionTypeOrigin,
			0.0,
			0.0,
		),
	)
}

func (dag *DAG) addTransaction(transaction entity.Transaction) {
	dag.mu.Lock()         // Lock the mutex before modifying the map
	defer dag.mu.Unlock() // Ensure the mutex is unlocked after this function exits
	dag.Transactions[transaction.ID] = transaction
}

func (dag *DAG) getTransactionByID(id string) (entity.Transaction, bool) {
	dag.mu.Lock()         // Lock the mutex before reading the map
	defer dag.mu.Unlock() // Ensure the mutex is unlocked after this function exits
	_transaction, ok := dag.Transactions[id]
	return _transaction, ok
}

func (dag *DAG) ConfirmTransaction(transactionID string) error {
	dag.mu.Lock()         // Lock the mutex before modifying the map
	defer dag.mu.Unlock() // Ensure the mutex is unlocked after this function exits

	_transaction, ok := dag.Transactions[transactionID]
	if !ok {
		return fmt.Errorf("transaction with ID %s not found", transactionID)
	}

	_transaction.UpdateStatus(entity.TransactionStatusConfirmed)
	_transaction.UpdateTimestampConfirmed()

	dag.Transactions[transactionID] = _transaction

	return nil
}
