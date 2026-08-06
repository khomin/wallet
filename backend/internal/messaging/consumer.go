package messaging

import (
	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	ch         *amqp091.Channel
	exchange   string
	routingKey string
}

func NewConsumer(mq *RabbitMQ, exchange string, routingKey string) (*Consumer, error) {
	ch, err := mq.NewChannel()
	if err != nil {
		return nil, err
	}
	return &Consumer{
		ch:         ch,
		exchange:   exchange,
		routingKey: routingKey,
	}, nil
}

func (p *Consumer) Consume() (<-chan amqp091.Delivery, error) {
	// if err := p.ch.Consume(p.exchange, p.routingKey, false,
	// 	false,
	// 	amqp091.Publishing{
	// 		ContentType: "application/json",
	// 		Body:        bytes,
	// 	}); err != nil {
	// 	return err
	// }
	return p.ch.Consume(p.exchange, p.routingKey, false, false, false, false, nil)
	// if err != nil {
	// 	return err
	// }
	// return nil
}
