package main

import (
	"fmt"
	"os"
	"os/signal"

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

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	<-signalCh
	fmt.Println("Shutting down...")
}
