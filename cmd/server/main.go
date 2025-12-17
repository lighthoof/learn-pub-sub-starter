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
	const rmqConnString = "amqp://guest:guest@localhost:5672/"

	//Create a connection to the RMQ
	conn, err := amqp.Dial(rmqConnString)
	if err != nil {
		log.Fatalf("Cannot establish connection with broker: %v", err)
	}
	defer conn.Close()
	fmt.Print("Peril server established connection with broker")

	//Create a channel
	pauseChan, err := conn.Channel()
	if err != nil {
		log.Fatalf("Cannot establish pause channel: %v", err)
	}

	//Print help on startup
	gamelogic.PrintServerHelp()

	//Waiting for input
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		//parse input and run relevant code
		switch words[0] {
		case "pause":
			fmt.Println("Sending pause message")
			//Publish pause message
			err = pubsub.PublishJSON(
				pauseChan,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
			if err != nil {
				log.Printf("Unable to publish pause message: %v", err)
			}

		case "resume":
			fmt.Println("Sending resume message")
			//Publish pause message
			err = pubsub.PublishJSON(
				pauseChan,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
			if err != nil {
				log.Printf("Unable to publish resume message: %v", err)
			}

		case "quit":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Incorrect command, please use the command from provided list")
		}
	}

	/*c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Print("Exiting...")*/

}
