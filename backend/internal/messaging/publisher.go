package messaging

import (
	"github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	mq        *RabbitMQ
	ch        *amqp091.Channel
	queueName string
}

func NewPublisher(mq *RabbitMQ, queueName string) (*Publisher, error) {
	p := Publisher{
		queueName: queueName,
		mq:        mq,
	}
	if err := p.ensureChannel(); err != nil {
		return nil, err
	}
	return &p, nil
}

func (p *Publisher) Publish(bytes []byte) error {
	if err := p.ensureChannel(); err != nil {
		return err
	}
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

func (p *Publisher) ensureChannel() error {
	if p.ch != nil && !p.ch.IsClosed() {
		return nil
	}
	ch, err := p.mq.NewChannel()
	if err != nil {
		return err
	}
	_, err = ch.QueueDeclare(
		p.queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		return err
	}
	p.ch = ch
	return nil
}
