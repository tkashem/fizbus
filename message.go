package fizbus

import (
	"github.com/streadway/amqp"
	"github.com/tkashem/fizbus/upstream"
)

// Message encapsulates content of a message transmitted over a bus.
type Message struct {
	Payload string

	// AMQP message headers, the key-value pairs below can transmit request scoped out of band data
	Headers map[string]string
}

type converterImpl struct {
	table upstream.Table
}

func (ci converterImpl) Convert(delivery *amqp.Delivery) Message {
	headers := ci.table.From(delivery.Headers)
	m := Message{
		Payload: string(delivery.Body),
		Headers: headers,
	}

	return m
}
