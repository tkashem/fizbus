package fizbus

import (
	"golang.org/x/net/context"

	"github.com/streadway/amqp"
	"github.com/tkashem/fizbus/upstream"
)

// These are internal interfaces used by bus implementation
// The reason they are in one file is so that we can generate mocks for them from one source file.

// handler associated with an AMQP delivery
type handler interface {
	Handle(c context.Context, d amqp.Delivery)
}

// This is a communication channel between the bus and all worker go routines.
// On any error, the worker go routine(s) notifies the bus of error it encounters using this interface
type errorChannel interface {
	// Error returns the error channel
	Error() <-chan error

	// A worker uses this method to notify the bus of any error that it encountered
	// This is a nonblocking operation
	Send(err error) bool

	// Retrieves the error message from the channel
	// This is a nonblocking operation
	Receive() error
}

// Converts an AMQP delivery off of a queue to a fizbus message
type converter interface {
	Convert(*amqp.Delivery) Message
}

// executes logic while the bus is in running mode
type busRunner interface {
	Run(runtime *runtimeContext)
}

// sender or binder will implement the following inerface
type initializer interface {
	initialize(*startupContext, *runtimeContext)
}

// The bus uses the following to load and start itself.
// Bus initialization (startup) and runtime have been made first class concepts.
type starter interface {
	StartupContext(configuration *Configuration) (*startupContext, error)
	RuntimeContext(parent context.Context) (*runtimeContext, error)
	NewBinder(routingKey string, handler Handler)
	NewSender(replyTo string) Sender
	Start(*startupContext, *runtimeContext) error
}

// Queue consumer loop, continues to pull messages off a queue and handle it as specified
type queueConsumer interface {
	// todo: should it be chan<- *amqp.Delivery
	Consume(runtime *runtimeContext, deliveries <-chan amqp.Delivery, handler handler)
}

// A sender loops through a go channel and keeps publishing request(s) to desired queue
type senderLoop interface {
	Send(runtime *runtimeContext, channel upstream.Channel, sender sender)
}

// Creates handler appropriately for sender or binder
type handlerFactory interface {
	// Returns a handler associated with a request, appropriate for a binder
	NewRequestHandler(runtime *runtimeContext, channel upstream.Channel, appHandler Handler) handler

	// Returns a handler associated with a reply, appropriate for a sender
	NewReplyHandler(sender sender, channel upstream.Channel) handler
}

// Sending request to queue is an asynchronous process which implies that we need to maintain states for request(s) in flight.
// The following interface provides the necessary abstraction for that.
// The key is request correlation id, and value associated is a go channel of reply
type requestMap interface {
	Add(string, chan reply)
	Remove(string) bool
	Get(string) (chan reply, bool)
}

// Defines the behavior of a sender internally
type sender interface {
	OnReply(d amqp.Delivery)
	Publish(channel upstream.Channel, request *upstream.Request)
	Outgoing() <-chan *send
	ReplyTo() string
	ReplyChannel(correlationID string) <-chan reply
	Cleanup(correlationID string) bool
}

// acker sends batch ACK
type acker interface {
	ack(*amqp.Delivery) error
}
