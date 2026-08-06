package messaging

import (
	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	ch        *amqp091.Channel
	queueName string
}

func NewConsumer(mq *RabbitMQ, queueName string) (*Consumer, error) {
	ch, err := mq.NewChannel()
	if err != nil {
		return nil, err
	}
	return &Consumer{
		ch:        ch,
		queueName: queueName,
	}, nil
}

func (p *Consumer) Consume() (<-chan amqp091.Delivery, error) {
	return p.ch.Consume(p.queueName, "",
		false, false, false, false, nil,
	)
}
