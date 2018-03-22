package fizbus

import (
	"fmt"
)

// Sets up a sender
type senderSetup struct {
	sender     sender
	chain      handlerFactory
	consumer   queueConsumer
	senderLoop senderLoop
}

func (ss senderSetup) initialize(startup *startupContext, runtime *runtimeContext) {
	defer func() {
		startup.Ready.Done()
		fmt.Println("bus: sender setup completed")
	}()

	channel, err := runtime.Connection.Channel()
	if err != nil {
		startup.Error.Send(err)
		return
	}

	deliveries, err := channel.Consume(ss.sender.ReplyTo(), true)
	if err != nil {
		startup.Error.Send(err)
		return
	}

	// we need a handler that will be invoked for each message we pump off this queue
	handler := ss.chain.NewReplyHandler(ss.sender, channel)

	/* --
	a sender needs two worker loops -
	first one pumps messages off of reply queue and does stuff
	second one pulls messages out of sending channel and publishes to queue
	*/
	runtime.Shutdown.Add(2)
	go ss.consumer.Consume(runtime, deliveries, handler)
	go ss.senderLoop.Send(runtime, channel, ss.sender)
}
