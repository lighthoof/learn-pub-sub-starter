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
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	//Create a new game state for the user
	gameState := gamelogic.NewGameState(userName)

	//REPL loop
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		//parse input and run relevant code
		switch words[0] {
		case "spawn":
			if err = gameState.CommandSpawn(words); err != nil {
				log.Printf("Error during spawning a unit: %v", err)
				continue
			}
		case "move":
			_, err := gameState.CommandMove(words)
			if err != nil {
				log.Printf("Error during moving a unit: %v", err)
				continue
			}
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming is not allowed yet")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Incorrect command, please use the command from provided list")
		}
	}
}
