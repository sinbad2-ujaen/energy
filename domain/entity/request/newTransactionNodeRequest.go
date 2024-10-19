package request

import (
	"energy/domain/entity"
)

type NewTransactionNodeRequest struct {
	Transaction   entity.Transaction   `json:"transaction"`
	SelectionTips []entity.Transaction `json:"selectionTips"`
}
