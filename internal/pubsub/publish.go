package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
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

func PublishGob[T any](
	ch *amqp.Channel,
	exchange, key string,
	val T,
) error {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(val); err != nil {
		return err
	}

	ctx := context.Background()
	msg := amqp.Publishing{ContentType: "application/gob", Body: buff.Bytes()}
	if err := ch.PublishWithContext(ctx, exchange, key, false, false, msg); err != nil {
		return err
	}

	return nil
}

func PublishLog(
	ch *amqp.Channel,
	username, msg string,
) error {
	exchange := routing.ExchangePerilTopic
	key := routing.GameLogSlug + "." + username
	log := routing.GameLog{
		CurrentTime: time.Now(),
		Message:     msg,
		Username:    username,
	}
	return PublishGob(ch, exchange, key, log)
}
