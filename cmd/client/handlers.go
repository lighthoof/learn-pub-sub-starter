package main

import (
	"fmt"

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
		defer fmt.Print("> ")
		var ackType pubsub.AckType

		mvOut := gs.HandleMove(glMove)
		switch mvOut {
		case gamelogic.MoveOutcomeSamePlayer:
			ackType = pubsub.Ack
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
		default:
			ackType = pubsub.NackDiscard
		}
		return ackType
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		logMessage := ""
		username := gs.GetUsername()

		warOut, winner, loser := gs.HandleWar(rw)
		fmt.Printf("DEBUG winner=%q loser=%q outcome=%v\n", winner, loser, warOut)

		switch warOut {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			logMessage = fmt.Sprintf("%s won a war against %s", winner, loser)
			if err := pubsub.PublishLog(ch, username, logMessage); err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			logMessage = fmt.Sprintf("%s won a war against %s", winner, loser)
			if err := pubsub.PublishLog(ch, username, logMessage); err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			logMessage = fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			if err := pubsub.PublishLog(ch, username, logMessage); err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			logMessage = fmt.Sprintf("Unexpected war outcome: %v", warOut)
			if err := pubsub.PublishLog(ch, username, logMessage); err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.NackDiscard
		}
	}
}
