package fizbus_test

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"golang.org/x/net/context"

	"github.com/tkashem/fizbus"
)

type pingHandler struct{}

func (pingHandler) Handle(c context.Context, m fizbus.Message) fizbus.Message {
	return fizbus.Message{Payload: "pong"}
}

func ExampleBus() {
	// This is so that the app can reppond to INT signal and can gracefully shutdown the bus
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// prepare root context owned by the app
	root, cancelFunc := context.WithCancel(context.Background())

	bus := fizbus.New("amqp://guest:guest@localhost:5672/")
	sender := bus.NewSender("echo-reply")
	bus.Bind("ping", pingHandler{})

	done, err := bus.Start(root, fizbus.Default())
	if err != nil {
		fmt.Printf("app: type: %T; value: %q\n", err.Error(), err.Error())
		return
	}

	defer func() {
		// wait until the bus has completely shutdown
		<-done()
		fmt.Println("app: got done notification from bus, exiting")
	}()

	go sendAndReceive(root, sender)

	select {
	case s := <-c:
		fmt.Println("app: time to quit, cancelling root context", s)
		cancelFunc()
		return

	case <-done():
		return
	}
}

func sendAndReceive(root context.Context, sender fizbus.Sender) {
	// This function sends a request, waits for a reply with a timeout specified by the context
	// Then, it checks to see if it is time to quit, otherwise waits for a second before sending another request
	for {
		ctx, _ := context.WithTimeout(root, 2*time.Second)
		receiver := sender.Send(ctx, "ping", fizbus.Message{Payload: "Ping?"})

		m, err := receiver.Receive()

		if m != nil {
			// A reply came back from "ping"!
			fmt.Println("app: reply=" + m.Payload)
		} else {
			// No reply came back.
			fmt.Println(err.Error())
		}

		select {
		case <-time.After(time.Second * 1):
		case <-root.Done():
			fmt.Println("app: echo loop exiting")
			return
		}
	}
}
