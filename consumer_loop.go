package fizbus

import (
	"fmt"
	"github.com/streadway/amqp"
)

// A consumer loop keeps pulling messages off of a given queue (as represented by the deliveries go channel )
// and executes the specified handler for each message
type queueConsumerImpl struct{}

func (queueConsumerImpl) Consume(runtime *runtimeContext, deliveries <-chan amqp.Delivery, handler handler) {
	// Once the loop exits, we need to signal the bus
	defer func() {
		runtime.Shutdown.Done()
		fmt.Println("bus: queue consumer worker loop exited")
	}()

	for {
		select {
		case d := <-deliveries:
			// todo: need to check if the channel is closed
			handler.Handle(runtime.Context, d)

		case <-runtime.Context.Done():
			// The bus is shutting down, so let's bail out.
			return
		}
	}
}
