package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	deliveriesCh, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for delivery := range deliveriesCh {
			var msg T
			if err := json.Unmarshal(delivery.Body, &msg); err != nil {
				log.Printf("failed to unmarshal message: %v", err)
			} else {
				handler(msg)
			}
			if err := delivery.Ack(false); err != nil {
				log.Printf("failed to ack message: %v", err)
			}
		}
	}()
	return nil
}
