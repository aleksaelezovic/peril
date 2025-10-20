package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
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
				err = delivery.Nack(false, false)
				if err != nil {
					log.Printf("failed to nack message: %v", err)
				} else {
					log.Println("nacked message without requeue - decoding body failed")
				}
			} else {
				var err error
				switch handler(msg) {
				case Ack:
					err = delivery.Ack(false)
					log.Println("acknowledged message")
				case NackRequeue:
					err = delivery.Nack(false, true)
					log.Println("nacked message with requeue")
				case NackDiscard:
					err = delivery.Nack(false, false)
					log.Println("nacked message without requeue")
				}
				if err != nil {
					log.Printf("failed to ack/nack message: %v", err)
				}
			}
		}
	}()
	return nil
}
