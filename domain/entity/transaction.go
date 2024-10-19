package entity

import (
	"github.com/google/uuid"
	"time"
)

type Transaction struct {
	ID                 string
	TimestampCreated   time.Time
	TimestampAdded     time.Time
	TimestampConfirmed time.Time
	Token              float64
	Data               string
	From               string
	To                 string
	Nonce              uint64
	Fee                float64
	Status             string
	Type               string
}

func NewTransaction(from string, to string, token float64, data string, transactionType string, fee float64, nonce uint64) Transaction {
	id := generateUniqueID()

	return Transaction{
		ID:                 id,
		From:               from,
		To:                 to,
		Token:              token,
		Data:               data,
		Nonce:              nonce,
		Fee:                fee,
		TimestampCreated:   time.Now(),
		TimestampAdded:     time.Time{},
		TimestampConfirmed: time.Time{},
		Type:               transactionType,
		Status:             TransactionStatusPending,
	}
}

func (t *Transaction) UpdateStatus(newStatus string) {
	t.Status = newStatus
}

func (t *Transaction) UpdateTimestampAdded() {
	t.TimestampAdded = time.Now()
}

func (t *Transaction) UpdateTimestampConfirmed() {
	t.TimestampConfirmed = time.Now()
}

func generateUniqueID() string {
	u := uuid.New()
	return u.String()
}

func IsValidID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func IsValidAddress(address string) bool {
	return len(address) == 64
}

func IsValidFee(fee float64, data string) bool {
	return fee == CalculateFee(data)
}

func IsValidTransactionType(transactionType string) bool {
	return transactionType == TransactionTypeStandard || transactionType == TransactionTypeFast
}

func CalculateFee(transactionData string) float64 {
	fee := float64(len(transactionData)) / 10.0
	if fee < 1.0 {
		fee = 1.0
	}
	return fee
}

const (
	TransactionTypeOrigin   string = "transaction-origin"
	TransactionTypeStandard string = "transaction-standard"
	TransactionTypeFast     string = "transaction-fast"
)

const (
	TransactionStatusPending   string = "pending"
	TransactionStatusConfirmed string = "confirmed"
)
