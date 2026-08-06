package messaging

import (
	"github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	ch        *amqp091.Channel
	queueName string
}

func NewPublisher(mq *RabbitMQ, queueName string) (*Publisher, error) {
	ch, err := mq.NewChannel()
	if err != nil {
		return nil, err
	}
	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &Publisher{
		ch:        ch,
		queueName: queueName,
	}, nil
}

func (p *Publisher) Publish(bytes []byte) error {
	if err := p.ch.Publish(
		"",
		p.queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        bytes,
		}); err != nil {
		return err
	}
	return nil
}
