package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
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

	ch.PublishWithContext(ctx, exchange, key, false, false, msg)

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange, queueName, key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {

	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	var isDurable bool
	var isAutoDelete bool
	var isExclusive bool

	switch queueType {
	case Durable:
		isDurable = true
		isAutoDelete = false
		isExclusive = false
	case Transient:
		isDurable = false
		isAutoDelete = true
		isExclusive = true
	}

	queue, err := ch.QueueDeclare(queueName, isDurable, isAutoDelete, isExclusive, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err

	}

	return ch, queue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange, queueName, key string,
	queueType SimpleQueueType,
	handler func(T),
) error {
	queueChan, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	consChan, _ := queueChan.Consume(queueName, "", false, false, false, false, nil)

	go func(consumers <-chan amqp.Delivery) error {
		for cons := range consumers {
			message, err := UnmarshalType[T](cons)
			if err != nil {
				return err
			}
			handler(message)

			cons.Ack(false)
		}
		return nil
	}(consChan)

	return nil
}

func UnmarshalType[T any](d amqp.Delivery) (out T, err error) {

	if err := json.Unmarshal(d.Body, &out); err != nil {
		log.Print("Unable to unmarshal the delivery")
		return out, err
	}

	return out, err
}
