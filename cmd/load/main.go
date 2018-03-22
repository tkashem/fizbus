package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"golang.org/x/net/context"

	"github.com/tkashem/fizbus"
)

var (
	url         string
	mode        string
	senderCount int
	thinkTime   int
	concurrency int
)

func init() {
	flag.StringVar(&url, "endpoint", "amqp://guest:guest@192.168.99.100:5672/", "rabbitmq URL")
	flag.StringVar(&mode, "mode", "", "mode - sender, binder. The default value does both")
	flag.IntVar(&senderCount, "sender.count", 1, "The number sender go routines")
	flag.IntVar(&thinkTime, "sender.think-time", 0, "think time in millisecond(s)")
	flag.IntVar(&concurrency, "binder.concurrency", 1, "number of handler go routines")

	flag.Parse()
}

type echoHandler struct{}

func (echoHandler) Handle(c context.Context, m fizbus.Message) fizbus.Message {
	return m
}

func newBus() fizbus.Bus {
	return fizbus.New(url)
}

func main() {
	// This is so that the app can reppond to INT signal and can gracefully shutdown the bus
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	root, cancelFunc := context.WithCancel(context.Background())

	var (
		err  error
		done fizbus.Done
	)

	switch mode {
	case "sender":
		done, err = sender(root)
	case "binder":
		done, err = binder(root)
	}

	if err != nil {
		fmt.Printf("app: type: %T; value: %q\n", err.Error(), err.Error())
		return
	}

	defer func() {
		// wait until the bus has completely shutdown
		<-done()
		fmt.Println("app: got done notification from bus, exiting")
	}()

	select {
	case s := <-c:
		fmt.Println("app: time to quit, cancelling root context", s)
		cancelFunc()
		return

	case <-done():
		return
	}
}

func sender(root context.Context) (fizbus.Done, error) {
	bus := newBus()

	sender := bus.NewSender("echo-reply")

	done, err := bus.Start(root, fizbus.Default())
	if err != nil {
		fmt.Printf("app: type: %T; value: %q\n", err.Error(), err.Error())
		return nil, err
	}

	for i := 1; i <= senderCount; i++ {
		go sendLoop(root, sender, "echo")
	}

	return done, nil
}

func binder(root context.Context) (fizbus.Done, error) {
	bus := newBus()
	bus.Bind("echo", echoHandler{})

	configuration := &fizbus.Configuration{PrefetchCount: concurrency}
	done, err := bus.Start(root, configuration)
	if err != nil {
		return nil, err
	}

	return done, nil
}

func sendLoop(root context.Context, sender fizbus.Sender, destination string) {
	// This function sends a request, waits for a reply with a timeout specified by the context
	// Then, it checks to see if it is time to quit, otherwise waits for a second before sending another request
	for {
		ctx, _ := context.WithTimeout(root, 2*time.Second)
		receiver := sender.Send(ctx, "echo", fizbus.Message{Payload: "Ping?"})

		_, err := receiver.Receive()

		if err != nil {
			fmt.Println(err.Error())
		}

		select {
		case <-time.After(time.Millisecond * time.Duration(thinkTime)):
		case <-root.Done():
			fmt.Println("app: echo loop exiting")
			return
		}
	}
}
