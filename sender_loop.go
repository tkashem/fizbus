package fizbus

import (
	"errors"
	"fmt"

	"github.com/tkashem/fizbus/upstream"
)

type senderLoopImpl struct{}

func (sl *senderLoopImpl) Send(runtime *runtimeContext, channel upstream.Channel, sender sender) {
	// when we exit the loop, signal the bus
	defer func() {
		runtime.Shutdown.Done()
		fmt.Println("bus: sender loop exited")
	}()

	for {
		select {
		case <-runtime.Context.Done():
			// The bus is shutting down, time to quit.
			return

		case s, ok := <-sender.Outgoing():
			// if the channel is closed, then exit the loop, but before that notify the bus.
			if !ok {
				err := errors.New("send loop channel has been closed")
				runtime.Error.Send(err)
				return
			}

			sender := s.sender
			sender.Publish(channel, s.request)
		}
	}
}
