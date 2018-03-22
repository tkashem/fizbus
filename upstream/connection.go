package upstream

import (
	"github.com/streadway/amqp"
)

// NewDialer returns a new dialer
func NewDialer(url string) Dialer {
	return dialer{
		url: url,
	}
}

type dialer struct {
	url string
}

func (d dialer) Dial() (Connection, error) {
	conn, err := amqp.Dial(d.url)
	if err != nil {
		return nil, err
	}

	return connectionImpl{
		connection: conn,
	}, nil
}

type connectionImpl struct {
	connection *amqp.Connection
}

func (c connectionImpl) Channel() (Channel, error) {
	ch, err := c.connection.Channel()
	if err != nil {
		return nil, err
	}

	return channel{
		channel: ch,
		table:   amqpTableConverterImpl{},
	}, nil
}

func (c connectionImpl) Close() {
	// todo: need to implement, and before the bus shuts down we should close the connection
}

func (c connectionImpl) Notify() <-chan *amqp.Error {
	connectionClosedErr := c.connection.NotifyClose(make(chan *amqp.Error))
	return connectionClosedErr
}
