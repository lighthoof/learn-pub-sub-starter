package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	//Create a channel
	queueChan, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("unable to create channel: %v", err)
	}

	//Set queue parameters
	isDurable := true
	isAutoDelete := false
	isExclusive := false

	if queueType == SimpleQueueType(Transient) {
		isAutoDelete = true
		isExclusive = true
	}

	//Declare the queue
	queue, err := queueChan.QueueDeclare(
		queueName,
		isDurable,
		isAutoDelete,
		isExclusive,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("unable to declare queue: %v", err)
	}

	//Bind the queue
	if err = queueChan.QueueBind(
		queueName,
		key,
		exchange,
		false,
		nil,
	); err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("unable to bind queue: %v", err)
	}

	return queueChan, queue, nil
}
