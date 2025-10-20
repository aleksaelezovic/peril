package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Printf("Error connecting to rabbitmq: %s", err.Error())
		os.Exit(1)
	}
	fmt.Println("Successfully connected to rabbitmq.")
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Error creating channel: %s", err.Error())
		os.Exit(1)
	}
	defer ch.Close()

	err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
		IsPaused: true,
	})
	if err != nil {
		fmt.Printf("Error publishing message: %s", err.Error())
		os.Exit(1)
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	<-signalCh
	fmt.Println("Shutting down...")
}
