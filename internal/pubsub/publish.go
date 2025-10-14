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
	exchange, key string,
	val T,
) error {

	valJSON, err := json.Marshal(val)
	if err != nil {
		return err
	}

	ctx := context.Background()
	msg := amqp.Publishing{ContentType: "application/json", Body: valJSON}

	if err = ch.PublishWithContext(ctx, exchange, key, false, false, msg); err != nil {
		return err
	}

	return nil
}
