package fizbus_test

import (
	"fmt"
	"golang.org/x/net/context"
	"os"
	"os/signal"
	"time"

	"github.com/tkashem/fizbus"
)

type echoHandler struct{}

func (echoHandler) Handle(c context.Context, m fizbus.Message) fizbus.Message {
	return m
}

func Example() {
	// This is so that the app can reppond to INT signal and can gracefully shutdown the bus
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	bus := fizbus.New("amqp://guest:guest@localhost:5672/")
	bus.Bind("echo", echoHandler{})
	sender := bus.NewSender("echo-reply")

	root, cancelFunc := context.WithCancel(context.Background())

	done, err := bus.Start(root, fizbus.Default())
	if err != nil {
		fmt.Printf("app: type: %T; value: %q\n", err.Error(), err.Error())
		return
	}

	defer func() {
		cancelFunc()
		<-done()
	}()

	this, _ := context.WithTimeout(root, 1*time.Second)
	receiver := sender.Send(this, "echo", fizbus.Message{Payload: "fiz in a bus!"})

	reply, err := receiver.Receive()

	if err != nil {
		fmt.Println("reply=[" + reply.Payload + "]")
	}
}
