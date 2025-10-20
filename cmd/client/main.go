package main

import (
	"fmt"
	"os"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Printf("Error connecting to rabbitmq: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Println("Successfully connected to rabbitmq.")
	defer conn.Close()

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("Error getting username: %s\n", err.Error())
		os.Exit(1)
	}

	ch, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		fmt.Printf("Error declaring and binding queue: %s\n", err.Error())
		os.Exit(1)
	}
	defer ch.Close()

	gamestate := gamelogic.NewGameState(username)
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
			_, err := gamestate.CommandMove(words)
			if err != nil {
				fmt.Printf("Error moving: %s\n", err.Error())
			}
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
