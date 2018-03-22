package upstream

import (
	"github.com/streadway/amqp"
)

type channel struct {
	channel *amqp.Channel
	table   Table
}

// Consume sets up a queue, binds to it, and returns a channel that consumes from it
func (c channel) Consume(queue string, isExclusive bool) (<-chan amqp.Delivery, error) {
	if _, err := c.channel.QueueDeclare(
		queue,       // name
		false,       // durable
		false,       // delete when usused
		isExclusive, // exclusive
		false,       // no-wait
		nil,         // arguments
	); err != nil {
		return nil, err
	}
	if err := c.channel.QueueBind(
		queue,        // queue name
		queue,        // routing key
		"amq.direct", // exchange
		false,
		nil); err != nil {
		return nil, err
	}

	delivery, err := c.channel.Consume(
		queue,       // queue
		"",          // consumer
		false,       // auto ack
		isExclusive, // exclusive
		false,       // no local
		false,       // no wait
		nil,         // args
	)

	return delivery, err
}

// Publish publishes an AMQP message to the given queue
func (c channel) Publish(request *Request) error {
	headers := c.table.To(request.Headers)

	return c.channel.Publish(
		"amq.direct",       // exchange
		request.RoutingKey, // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			CorrelationId: request.CorrelationID,
			ReplyTo:       request.ReplyTo,
			ContentType:   "text/plain",
			Body:          []byte(request.Body),
			Headers:       headers,
		})
}

func (c channel) Prefetch(count int) error {
	return c.channel.Qos(count, 0, false)
}
