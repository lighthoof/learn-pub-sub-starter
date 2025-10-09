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
	pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.Transient,
	)

	gameState := gamelogic.NewGameState(username)
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
			gameState.CommandMove(input)
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
