package main

import (
	"fmt"
	"log"
	"os"

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
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		switch gs.HandleMove(move) {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: gs.Player,
					Defender: move.Player,
				},
			)
			if err != nil {
				log.Printf("Error publishing war recognition: %s", err.Error())
			}
			return pubsub.NackRequeue
		default:
			return pubsub.NackDiscard
		}
	}
}

func handlerWar(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, _, _ := gs.HandleWar(rw)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			return pubsub.Ack
		default:
			fmt.Printf("Unknown outcome: %d\n", outcome)
			return pubsub.NackDiscard
		}
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Printf("Error connecting to rabbitmq: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Println("Successfully connected to rabbitmq.")
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Error creating channel: %s\n", err.Error())
		os.Exit(1)
	}
	defer ch.Close()

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("Error getting username: %s\n", err.Error())
		os.Exit(1)
	}

	gamestate := gamelogic.NewGameState(username)
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gamestate),
	)
	if err != nil {
		fmt.Printf("Error subscribing to queue: %s\n", err.Error())
		os.Exit(1)
	}
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
		handlerMove(gamestate, ch),
	)
	if err != nil {
		fmt.Printf("Error subscribing to queue: %s\n", err.Error())
		os.Exit(1)
	}
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.Durable,
		handlerWar(gamestate),
	)
	if err != nil {
		fmt.Printf("Error subscribing to queue: %s\n", err.Error())
		os.Exit(1)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		command := words[0]
		if command == "quit" {
			gamelogic.PrintQuit()
			break
		}
		if command == "help" {
			gamelogic.PrintClientHelp()
			continue
		}
		if command == "status" {
			gamestate.CommandStatus()
			continue
		}
		if command == "spam" {
			fmt.Println("Spamming not allowed yet!")
			continue
		}
		if command == "move" {
			move, err := gamestate.CommandMove(words)
			if err != nil {
				fmt.Printf("Error moving: %s\n", err.Error())
				continue
			}
			err = pubsub.PublishJSON(ch, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, move)
			if err != nil {
				fmt.Printf("Error publishing move: %s\n", err.Error())
			}
			log.Println("Move published successfully.")
			continue
		}
		if command == "spawn" {
			err := gamestate.CommandSpawn(words)
			if err != nil {
				fmt.Printf("Error spawning: %s\n", err.Error())
			}
			continue
		}
		fmt.Println("Command not recognized.")
	}

	// signalCh := make(chan os.Signal, 1)
	// signal.Notify(signalCh, os.Interrupt)
	// <-signalCh
	// fmt.Println("Shutting down...")
}
