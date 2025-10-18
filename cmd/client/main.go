package main

import (
	"fmt"
	"log"

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

	pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.Transient,
	)

	gameState := gamelogic.NewGameState(username)
	pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState),
	)

	warChan, _ := connection.Channel()
	pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
		handlerMove(gameState, warChan),
	)

	pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		"war",
		routing.ArmyMovesPrefix+".*",
		pubsub.Durable,
		handlerWar(gameState, warChan),
	)

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		command := input[0]
		switch command {
		case "spawn":
			gameState.CommandSpawn(input)
		case "move":
			moveChan, err := connection.Channel()
			if err != nil {
				fmt.Println("Cannot create move channel")
			}

			moveMessage, err := gameState.CommandMove(input)
			if err != nil {
				fmt.Println("Cannot make a move")
			}

			if err := pubsub.PublishJSON(moveChan, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, moveMessage); err != nil {
				fmt.Println("Cannot publish pause message")
			}

			log.Print("Move published successfully")
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Command unclear. Please use a command from the list")
		}
	}

	/*sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan*/
}
