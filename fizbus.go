package fizbus

import (
	"golang.org/x/net/context"

	"github.com/tkashem/fizbus/upstream"
)

// New creates a new instance of Bus to use, it expects a URL to underlying AMQP broker
// Note that this function does not initiate a connection to underlying AMQP broker yet
func New(url string) Bus {
	dialer := upstream.NewDialer(url)

	starter := &starterImpl{
		dialer:  dialer,
		routing: routegen,
	}

	return &bus{
		starter: starter,
		runner:  &busRunnerImpl{},
	}
}

// Default returns bus configuration with default values.
func Default() *Configuration {
	return &Configuration{PrefetchCount: 1}
}

// Configuration allows the user to specify configuration of the bus
type Configuration struct {
	// Prefetch count dictates the server will try to keep on the network for
	// consumers before receiving delivery acks
	PrefetchCount int

	//
	AckThreshold int
}

// Bus is an interface that represents a logical message bus. Buses support two modes of operation.
// In the first mode, an application binds itself to a queue, consumes message(s) off of it and send reply back
// In the second, a return address is configured to support a logically-single-threaded request-response pattern.
// No connections or other resources should be acquired by a Bus until it is started.
type Bus interface {
	// Bind allows a caller to register a new binder with the bus. The handler specified will be executed
	// for each message pulled off of the request queue.
	Bind(string, Handler)

	// NewSender returns a new instance of Sender. Each Sender object is identified with a unique reply queue name
	NewSender(string) Sender

	// Start initiates a connection to underlying AMQP broker and then sets up all the senders and binders
	// On any error while the setup is in progress, the bus will abort and return the appropriate error to the caller
	// If the setup is successful, it returns a function which internalizes a channel which
	// will be closed when the bus completely shuts down
	Start(context.Context, *Configuration) (Done, error)
}

// Handler handles a request message and returns a message to be returned to the caller
type Handler interface {
	Handle(context.Context, Message) Message
}

// Sender is uniquely identified a return address and is capable of sending request(s) to designated queues.
type Sender interface {
	// Send sends a request to a designated queue and returns a Receiver object which
	// can be used to wait and receive the associated reply
	// The returned Reply may contain a Message or an explanatory error.
	Send(context.Context, string, Message) Receiver
}

// Receiver receives a reply message upon sending a request
type Receiver interface {
	Receive() (*Message, error)
}

// Done returns a channel which will be closed when the bus has completely shutdown
type Done func() <-chan struct{}
