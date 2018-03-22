package fizbus

import (
	"fmt"
	"golang.org/x/net/context"

	"github.com/tkashem/fizbus/upstream"
)

// Generates appropriate routing key based off a given base key
type routing func(string) string

func routegen(s string) string {
	return fmt.Sprintf("%s-%s", s, newUUID())
}

type starterImpl struct {
	dialer       upstream.Dialer
	initializers []initializer
	routing      routing
}

func (s *starterImpl) StartupContext(configuration *Configuration) (*startupContext, error) {
	sc := startupContext{
		Error: newErrorChannel(),

		// todo: this values should be passed down as a configuration parameter
		// when the bus start method is invoked
		configuration: configuration,
	}

	return &sc, nil
}

func (s *starterImpl) RuntimeContext(parent context.Context) (*runtimeContext, error) {
	// runtime requires a connection to AMQP broker, this is the place we connect to the broker.
	connection, err := s.dialer.Dial()

	if err != nil {
		return nil, err
	}

	// The net/context object owned by the bus and passed to its workers
	context, cancelFunc := context.WithCancel(parent)

	runtime := runtimeContext{
		Error:      newErrorChannel(),
		Parent:     parent,
		Context:    context,
		CancelFunc: cancelFunc,
		Connection: connection,
		done:       make(chan struct{}),
	}

	return &runtime, nil
}

func (s *starterImpl) Start(startup *startupContext, runtime *runtimeContext) error {
	if len(s.initializers) == 0 {
		return nil
	}

	startup.Ready.Add(len(s.initializers))

	// We have a number of initializer objects, we can kick them off now as the connection to underlying
	// AMQP broker has already been established
	for _, initializer := range s.initializers {
		go initializer.initialize(startup, runtime)
	}

	startup.Ready.Wait()
	fmt.Println("bus: startup completed")

	// did any of the initialization go routines encounter any error?
	if busError := startup.Error.Receive(); busError != nil {
		return busError
	}

	// The setup has been successful
	return nil
}

func (s *starterImpl) NewBinder(routingKey string, handler Handler) {
	setup := binderSetup{
		routingKey:  routingKey,
		userHandler: handler,
		consumer:    queueConsumerImpl{},
		factory:     handlerFactoryImpl{},
	}

	// We can't start the setup yet, as connection to AMQP broker hasnt been established yet
	// So, we will save them to be invoked later
	s.append(&setup)
}

func (s *starterImpl) NewSender(replyTo string) Sender {
	replyQueue := s.routing(replyTo)
	sender := &senderImpl{
		replyTo:         replyQueue,
		idgen:           newUUID,
		chgen:           replyChGenerator,
		outgoing:        make(chan *send),
		converter:       converterImpl{table: upstream.NewTable()},
		requestChannels: &requestMapImpl{},
	}

	setup := senderSetup{
		sender:     sender,
		chain:      handlerFactoryImpl{},
		consumer:   queueConsumerImpl{},
		senderLoop: &senderLoopImpl{},
	}

	// We can't start the setup yet, as connection to AMQP broker hasnt been established yet
	// So, we will save them to be invoked later
	s.append(setup)

	return sender
}

func (s *starterImpl) append(i initializer) {
	if len(s.initializers) == 0 {
		s.initializers = make([]initializer, 0)
	}

	s.initializers = append(s.initializers, i)
}
