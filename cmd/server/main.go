package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rmqConnString := "amqp://guest:guest@localhost:5672/"

	//fmt.Println("Starting Peril server...")
	connection, err := amqp.Dial(rmqConnString)
	if err != nil {
		fmt.Println("Cannot obtain a server connection")
		return
	}
	defer connection.Close()
	fmt.Println("Connection successful")

	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}
