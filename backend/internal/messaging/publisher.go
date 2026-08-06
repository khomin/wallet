package messaging

import (
	"github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	ch         *amqp091.Channel
	exchange   string
	routingKey string
}

func NewPublisher(mq *RabbitMQ, exchange string, routingKey string) (*Publisher, error) {
	ch, err := mq.NewChannel()
	if err != nil {
		return nil, err
	}
	return &Publisher{
		ch:         ch,
		exchange:   exchange,
		routingKey: routingKey,
	}, nil
}

func (p *Publisher) Publish(bytes []byte) error {
	_, err := p.ch.QueueDeclare(
		p.exchange,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	if err := p.ch.Publish(p.exchange, p.routingKey, false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        bytes,
		}); err != nil {
		return err
	}
	return nil
}
