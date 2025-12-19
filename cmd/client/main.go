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

	const rmqConnString = "amqp://guest:guest@localhost:5672/"

	//Establish a connection with RMQ broker
	conn, err := amqp.Dial(rmqConnString)
	if err != nil {
		log.Fatal("Cannot establish connection with broker")
	}
	defer conn.Close()
	fmt.Println("Established connection with broker")

	//Create a publish channel
	pubChannel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Cannot establish move channel: %v", err)
	}

	//Prompt for username
	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Unable to obtain username: %e", err)
	}

	/*//Bind the pause queue to exchange
	_, queue, err := pubsub.DeclareAndBind(
		conn,                                     //connection
		routing.ExchangePerilDirect,              //direct exchange
		routing.PauseKey+"."+userName,            //pause.username
		routing.PauseKey,                         //pause
		pubsub.Transient, //transient queue
	)
	if err != nil {
		log.Fatalf("Unable to subscribe to pause ququeL %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)*/

	//Create a new game state for the user
	gameState := gamelogic.NewGameState(userName)

	//Subscribe to pause queue
	if err = pubsub.SubscribeJSON(
		conn,                          //connection
		routing.ExchangePerilDirect,   //exchange type
		routing.PauseKey+"."+userName, //queue name
		routing.PauseKey,              //routing key
		pubsub.Transient,              //queue type
		handlerPause(gameState),       //handler function
	); err != nil {
		log.Fatalf("Unable to subscribe to pause queue: %v", err)
	}

	//Subscribe to move queue
	if err = pubsub.SubscribeJSON(
		conn,                                 //connection
		routing.ExchangePerilTopic,           //exchange type
		routing.ArmyMovesPrefix+"."+userName, //queue name
		routing.ArmyMovesPrefix+".*",         //routing key
		pubsub.Transient,                     //queue type
		handlerMove(gameState, pubChannel),   //handler function
	); err != nil {
		log.Fatalf("Unable to subscribe to move queue: %v", err)
	}

	//Subscribe to war queue
	if err = pubsub.SubscribeJSON(
		conn,                               //connection
		routing.ExchangePerilTopic,         //exchange type
		routing.WarRecognitionsPrefix,      //queue name
		routing.WarRecognitionsPrefix+".*", //routing key
		pubsub.Durable,                     //queue type
		handlerWar(gameState),              //handler function
	); err != nil {
		log.Fatalf("Unable to subscribe to WAR queue: %v", err)
	}

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
			move, err := gameState.CommandMove(words)
			if err != nil {
				log.Printf("Error during moving a unit: %v", err)
				continue
			}
			//Publish move message
			err = pubsub.PublishJSON(
				pubChannel,                 //channel
				routing.ExchangePerilTopic, //exchange type
				routing.ArmyMovesPrefix+"."+gameState.GetUsername(), //routing key
				move, //payload
			)
			if err != nil {
				log.Printf("Unable to publish move message: %v", err)
			}
			//fmt.Printf("Moved %v units to %s\n", len(move.Units), move.ToLocation)
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
