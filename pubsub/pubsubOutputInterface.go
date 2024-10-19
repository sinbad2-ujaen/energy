package pubsub

import (
	"context"
	"energy/domain/entity"
)

type PubsubOutputInterface interface {
	BroadcastMessage(otherContext context.Context, message entity.PubsubMessage)
}
