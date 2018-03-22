package upstream

import (
	"github.com/streadway/amqp"
)

// Request to be sent to a queue to be processed by a consumer bound to the given queue
type Request struct {
	RoutingKey    string
	CorrelationID string
	ReplyTo       string
	Body          string
	Headers       map[string]string
}

// Dialer dials to the AMQP broker specified by the url and returns an interface
// which provides abstraction to an nderlying AMQP connection
type Dialer interface {
	Dial() (Connection, error)
}

// Connection abstracts a connection to AMQP broker
type Connection interface {
	Channel() (Channel, error)
	Notify() <-chan *amqp.Error
	Close()
}

// Channel abstracts operations associated with an AMQP channel
type Channel interface {
	Consume(queue string, isExclusive bool) (<-chan amqp.Delivery, error)
	Publish(request *Request) error
	Prefetch(count int) error
}

// Table converts AMQP headers back and forth
type Table interface {
	// To converts standard map of <string, string> to AMQP message header
	To(h map[string]string) amqp.Table

	// From converts AMQP message header to standard map of <string, string>
	From(t amqp.Table) map[string]string
}
