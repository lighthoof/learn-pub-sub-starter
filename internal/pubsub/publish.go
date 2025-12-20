package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
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
		context.Background(), //context
		exchange,             //exchange name
		key,                  //routing key
		false,                //mandatory
		false,                //immediate
		amqp.Publishing{ //message
			ContentType: "application/json",
			Body:        valBytes,
		},
	)
}

func PublishGob[T any](
	ch *amqp.Channel,
	exchange,
	key string,
	val T,
) error {
	//create a send buffer
	var buffer bytes.Buffer
	//create an encoder
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(val); err != nil {
		return err
	}

	return ch.PublishWithContext(
		context.Background(), //context
		exchange,             //exchange name
		key,                  //routing key
		false,                //mandatory
		false,                //immediate
		amqp.Publishing{ //message
			ContentType: "application/gob",
			Body:        buffer.Bytes(),
		},
	)
}

/*func EncodeGob(gameLog routing.GameLog) ([]byte, error) {
	var buffer bytes.Buffer        //send buffer
	enc := gob.NewEncoder(&buffer) //Create an encoder
	if err := enc.Encode(gameLog); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func DecodeGob(byteData []byte) (routing.GameLog, error) {
	buffer := bytes.NewBuffer(byteData)
	dec := gob.NewDecoder(buffer) //create a decoder
	var data routing.GameLog      //data to return
	if err := dec.Decode(&data); err != nil {
		return routing.GameLog{}, err
	}
	return data, nil

}*/
