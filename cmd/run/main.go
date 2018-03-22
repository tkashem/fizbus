package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"golang.org/x/net/context"

	"github.com/tkashem/fizbus"
)

type echoHandler struct{}

func (echoHandler) Handle(c context.Context, m fizbus.Message) fizbus.Message {
	return m
}

func main() {
	// This is so that the app can reppond to INT signal and can gracefully shutdown the bus
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	bus := fizbus.New("amqp://guest:guest@192.168.99.100:5672/")
	sender := bus.NewSender("echo-reply")
	bus.Bind("echo", echoHandler{})

	root, cancelFunc := context.WithCancel(context.Background())
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
		receiver := sender.Send(ctx, "echo", fizbus.Message{Payload: "Ping?"})

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
