package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

var table = amqp.Table{
	"x-dead-letter-exchange": "peril_dlx",
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	q, err := ch.QueueDeclare(
		queueName,
		queueType == Durable,
		queueType == Transient,
		queueType == Transient,
		false,
		table,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	err = ch.QueueBind(q.Name, key, exchange, false, table)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return ch, q, nil
}
