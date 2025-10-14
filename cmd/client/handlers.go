package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(glMove gamelogic.ArmyMove) pubsub.AckType {
		var ackType pubsub.AckType
		defer fmt.Print("> ")
		mvOut := gs.HandleMove(glMove)
		switch mvOut {
		case gamelogic.MoveOutComeSafe:
			ackType = pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			ackType = pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			ackType = pubsub.NackDiscard
		default:
			ackType = pubsub.NackDiscard
		}

		return ackType
	}
}
