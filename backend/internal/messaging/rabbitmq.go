package messaging

import "github.com/rabbitmq/amqp091-go"

type RabbitMQ struct {
	conn *amqp091.Connection
	// ch   *amqp091.Channel
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}
	// ch, err := conn.Channel()
	// if err != nil {
	// 	conn.Close()
	// 	return nil, err
	// }

	return &RabbitMQ{
		conn: conn,
		// ch:   ch,
	}, nil
}

// func (r *RabbitMQ) Publish(queue string, body []byte) error {
// 	_, err := r.ch.QueueDeclare(
// 		queue,
// 		true,
// 		false,
// 		false,
// 		false,
// 		nil,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	return r.ch.Publish(
// 		"",
// 		queue,
// 		false,
// 		false,
// 		amqp091.Publishing{
// 			ContentType: "application/json",
// 			Body:        body,
// 		},
// 	)
// }

// func (r *RabbitMQ) Close() error {
// 	_ = r.ch.Close()
// 	return r.conn.Close()
// }

func (r *RabbitMQ) NewChannel() (*amqp091.Channel, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, err
	}
	return ch, nil
}

// func (r *RabbitMQ) NewChannel(queue string) (<-chan amqp091.Delivery, error) {
// 	return r.ch.Consume(queue, "wallet-worker-1", false, false, false, false, nil)
// }
