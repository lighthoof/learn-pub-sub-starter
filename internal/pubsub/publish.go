package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int
type AckType int

const (
	Durable SimpleQueueType = iota
	Transient
)

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func PublishJSON[T any](
	ch *amqp.Channel,
	exchange,
	key string,
	val T,
) error {
	valBytes, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        valBytes,
		},
	)
}
