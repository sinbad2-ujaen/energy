package response

import (
	"energy/domain/entity"
)

type SelectionTipsResponse struct {
	SelectionTips []entity.Transaction `json:"transactions"`
}
