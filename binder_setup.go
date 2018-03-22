package fizbus

import (
	"fmt"
)

type binderSetup struct {
	routingKey  string
	factory     handlerFactory
	consumer    queueConsumer
	userHandler Handler
}

// sets up a binder
func (bs *binderSetup) initialize(startup *startupContext, runtime *runtimeContext) {
	defer func() {
		startup.Ready.Done()
		fmt.Println("bus: binder setup completed")
	}()

	channel, err := runtime.Connection.Channel()
	if err != nil {
		// failed to open a channel, propagate the error to the bus and bail out
		startup.Error.Send(err)
		return
	}

	err = channel.Prefetch(startup.configuration.PrefetchCount)
	if err != nil {
		// failed to open a channel, propagate the error to the bus and bail out
		startup.Error.Send(err)
		return
	}

	deliveries, err := channel.Consume(bs.routingKey, false)
	if err != nil {
		startup.Error.Send(err)
		return
	}

	// chain of handler that will be invoked for each message we pump off this queue
	handler := bs.factory.NewRequestHandler(runtime, channel, bs.userHandler)

	// kick off the consumer loop that will consume off of a queue and execute the handler specified
	runtime.Shutdown.Add(startup.configuration.PrefetchCount)
	for i := 1; i <= startup.configuration.PrefetchCount; i++ {
		go bs.consumer.Consume(runtime, deliveries, handler)
	}
}
