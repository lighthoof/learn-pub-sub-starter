package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	log.Print("Starting Peril client...")

	rmqConnString := "amqp://guest:guest@localhost:5672/"

	//Establish a connection with RMQ broker
	conn, err := amqp.Dial(rmqConnString)
	if err != nil {
		log.Fatal("Cannot establish connection with broker")
	}
	defer conn.Close()
	fmt.Println("Established connection with broker")

	//Prompt for username
	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Unable to obtain username: %e", err)
	}

	//Bind the queue to exchange
	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+userName,
		routing.PauseKey,
		pubsub.SimpleQueueType(pubsub.Transient),
	)
	if err != nil {
		log.Fatalf("Unable to subscribe to pause ququeL %v", err)
	}
	fmt.Printf("Queue %v declared and bound!", queue.Name)

	//Waiting for input
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Printf("Closing connection to RMQ")

}
