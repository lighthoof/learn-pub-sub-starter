package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerLogs() func(routing.GameLog) pubsub.AckType {
	return func(gameLog routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		if err := gamelogic.WriteLog(gameLog); err != nil {
			log.Printf("Unable to write a log to disk: %v", err)
			return pubsub.NackDiscard
		}
		return pubsub.Ack
	}

}
