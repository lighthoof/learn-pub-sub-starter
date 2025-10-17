package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange, queueName, key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	queueChan, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	consChan, _ := queueChan.Consume(queueName, "", false, false, false, false, nil)

	go func(consumers <-chan amqp.Delivery) {
		for cons := range consumers {
			message, err := UnmarshalType[T](cons)
			if err != nil {
				log.Print("Unable to unmarshal consume delivery")
				return
			}
			ackType := handler(message)
			switch ackType {
			case Ack:
				cons.Ack(false)
				//log.Printf("Acknowledged: %v", message)
			case NackRequeue:
				cons.Nack(false, true)
				//log.Printf("Not acknowledged, requeue: %v", message)
			case NackDiscard:
				cons.Nack(false, false)
				//log.Printf("Not acknowledged, discard: %v", message)
			default:
				log.Printf("Unknown ack type registered: %v", ackType)
			}

		}
	}(consChan)

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

	dlxTable := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}

	queue, err := ch.QueueDeclare(queueName, isDurable, isAutoDelete, isExclusive, false, dlxTable)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err

	}

	return ch, queue, nil
}

func UnmarshalType[T any](d amqp.Delivery) (out T, err error) {

	if err := json.Unmarshal(d.Body, &out); err != nil {
		log.Print("Unable to unmarshal the delivery")
		return out, err
	}

	return out, err
}
