package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(glMove gamelogic.ArmyMove) pubsub.AckType {
		var ackType pubsub.AckType
		defer fmt.Print("> ")

		mvOut := gs.HandleMove(glMove)
		switch mvOut {
		case gamelogic.MoveOutComeSafe:
			ackType = pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			if err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{Attacker: glMove.Player, Defender: gs.GetPlayerSnap()},
			); err != nil {
				ackType = pubsub.NackRequeue
			} else {
				ackType = pubsub.Ack
			}
		case gamelogic.MoveOutcomeSamePlayer:
			ackType = pubsub.NackDiscard
		default:
			ackType = pubsub.NackDiscard
		}

		return ackType
	}
}

func handlerWar(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		var ackType pubsub.AckType
		defer fmt.Print("> ")

		warOut, winner, loser := gs.HandleWar(rw)
		switch warOut {
		case gamelogic.WarOutcomeNotInvolved:
			ackType = pubsub.NackDiscard
		case gamelogic.WarOutcomeNoUnits:
			ackType = pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			ackType = pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			ackType = pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			ackType = pubsub.Ack
		default:
			log.Print("Incorrect war outcome")
			ackType = pubsub.NackDiscard
		}
		log.Printf("Winner: %v, Loser: %v", winner, loser)

		return ackType
	}
}
