package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rmqConnString := "amqp://guest:guest@localhost:5672/"

	fmt.Println("Starting Peril client...")
	connection, err := amqp.Dial(rmqConnString)
	if err != nil {
		fmt.Println("Cannot obtain a server connection")
		return
	}
	defer connection.Close()
	fmt.Println("Connection successful")

	username, _ := gamelogic.ClientWelcome()

	queueName := routing.PauseKey + "." + username
	queueType := pubsub.SimpleQueueType(1)
	pubsub.DeclareAndBind(connection, routing.ExchangePerilDirect, queueName, routing.PauseKey, queueType)

	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
}
