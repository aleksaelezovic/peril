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

	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Error creating channel: %s\n", err.Error())
		os.Exit(1)
	}
	defer ch.Close()

	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.Durable,
		func(gl routing.GameLog) pubsub.AckType {
			defer fmt.Print("> ")
			err := gamelogic.WriteLog(gl)
			if err != nil {
				fmt.Printf("Error writing log: %s\n", err.Error())
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		},
	)
	if err != nil {
		fmt.Printf("Error declaring and binding: %s\n", err.Error())
		os.Exit(1)
	}

	gamelogic.PrintServerHelp()
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		command := words[0]
		if command == "quit" {
			fmt.Println("Exiting...")
			break
		}
		if command == "pause" {
			fmt.Println("Sending pause message...")
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
			if err != nil {
				fmt.Printf("Error publishing message: %s\n", err.Error())
			}
			continue
		}
		if command == "resume" {
			fmt.Println("Sending resume message...")
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
			if err != nil {
				fmt.Printf("Error publishing message: %s\n", err.Error())
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
