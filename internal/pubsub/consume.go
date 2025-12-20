package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	//create an unmarshaller
	unmarshaller := func(data []byte) (T, error) {
		var target T
		err := json.Unmarshal(data, &target)
		return target, err
	}

	//subscribe to JSON format queue
	if err := subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		unmarshaller,
	); err != nil {
		return err
	}
	return nil
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	//create an unmarshaller
	unmarshaller := func(data []byte) (T, error) {
		buffer := bytes.NewBuffer(data)
		var target T
		dec := gob.NewDecoder(buffer) //create a decoder
		err := dec.Decode(&target)
		return target, err
	}

	//subscribe to JSON format queue
	if err := subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		unmarshaller,
	); err != nil {
		return err
	}
	return nil
}

func subscribe[T any](
	conn *amqp.Connection, //connection
	exchange, //exchange name
	queueName, //queue name
	key string, //routing key
	queueType SimpleQueueType, //queue type
	handler func(T) AckType, //handler function
	unmarshaller func([]byte) (T, error), //unmarshaller function
) error {
	//Declare and biand a queue to subscribe to provided exchange by the key
	queueChannel, queue, err := DeclareAndBind(
		conn,      //connection
		exchange,  //exchange name
		queueName, //queue name
		key,       //routing key
		queueType, //queue type
	)
	if err != nil {
		return err
	}

	//Create a consumer channel for the queue
	consumeChan, err := queueChannel.Consume(
		queue.Name, //queue name
		"",         //consumer name
		false,      //auto-ack
		false,      //exclusive
		false,      //no-local
		false,      //no-wait
		nil,        //args
	)
	if err != nil {
		return err
	}

	//loop over the incomming messages on consumer channel
	go func() {
		for message := range consumeChan {
			//unmarshal the body to raw bytes
			rawBody, err := unmarshaller(message.Body)
			if err != nil {
				log.Printf("Unable to unmarshal message: %v\n", err)
			}

			//handle the message
			switch handler(rawBody) {
			//TODO add error handling if handlers are supposed to return errors

			//Ack received message to remove it from queue
			case Ack:
				if err = message.Ack(false); err != nil {
					log.Printf("Error acking the message %v : %v\n", message, err)
				}
				log.Println("Succesfully acked the messaged")
			//Requeue the message
			case NackRequeue:
				if err = message.Nack(false, true); err != nil {
					log.Printf("Error requeueing the message %v : %v\n", message, err)
				}
				log.Println("Succesfully requeued the messaged")
			//Discard the message
			case NackDiscard:
				if err = message.Nack(false, false); err != nil {
					log.Printf("Error discarding the message %v : %v\n", message, err)
				}
				log.Println("Succesfully discarded the messaged")
			}
		}
	}()
	return nil
}

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
		queueName,    //name
		isDurable,    //queque type
		isAutoDelete, //AutoDeletion option
		isExclusive,  //exclusive option
		false,        //waiting option
		amqp.Table{"x-dead-letter-exchange": "peril_dlx"}, //args
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
