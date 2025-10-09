package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rmqConnString := "amqp://guest:guest@localhost:5672/"

	fmt.Println("Starting Peril server...")
	connection, err := amqp.Dial(rmqConnString)
	if err != nil {
		fmt.Println("Cannot obtain a server connection")
		return
	}
	defer connection.Close()
	fmt.Println("Connection successful")

	gamelogic.PrintServerHelp()

	/*
		}*/
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		command := input[0]
		switch command {
		case "pause":
			gamePause(connection, true)
		case "resume":
			gamePause(connection, true)
		case "quit":
			return
		default:
			fmt.Println("Command unclear. Please use a command from the list")
		}
	}

	/*sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan*/
}

func gamePause(connection *amqp.Connection, toPause bool) {
	pauseChan, err := connection.Channel()
	if err != nil {
		fmt.Println("Cannot create pause channel")
	}

	pauseMsg := routing.PlayingState{IsPaused: toPause}
	err = pubsub.PublishJSON(pauseChan, routing.ExchangePerilDirect, routing.PauseKey, pauseMsg)
	if err != nil {
		fmt.Println("Cannot publish pause message")
	}
}
