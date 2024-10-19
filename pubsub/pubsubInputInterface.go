package pubsub

import (
	"energy/domain/entity"
)

type PubsubInputInterface interface {
	WalletCreateMessage(message entity.PubsubMessage)
	NewTransactionMessage(message entity.PubsubMessage)
}
