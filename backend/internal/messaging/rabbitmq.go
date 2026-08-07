package messaging

import "github.com/rabbitmq/amqp091-go"

type RabbitMQ struct {
	conn *amqp091.Connection
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}
	return &RabbitMQ{
		conn: conn,
	}, nil
}

func (r *RabbitMQ) Close() error {
	return r.conn.Close()
}

func (r *RabbitMQ) NewChannel() (*amqp091.Channel, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, err
	}
	return ch, nil
}
