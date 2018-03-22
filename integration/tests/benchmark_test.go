package tests

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tkashem/fizbus"
	"golang.org/x/net/context"
)

func BenchmarkSend(b *testing.B) {

	// setup
	t := &testing.T{}
	bus := fizbus.New(endpoint)
	sender := bus.NewSender("echo-reply")

	root, cancelFunc := context.WithCancel(context.Background())
	done, err := bus.Start(root, fizbus.Default())
	defer func() {
		cancelFunc()
		<-done()
	}()

	require.Nil(t, err)
	require.NotNil(t, done)

	this, _ := context.WithCancel(root)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sender.Send(this, "echo", fizbus.Message{Payload: "hello!"})
		}
	})
}

func BenchmarkSendAndReceive(b *testing.B) {
	// setup
	t := &testing.T{}
	bus := fizbus.New(endpoint)
	sender := bus.NewSender("echo-reply")
	bus.Bind("echo", echoHandler{})

	root, cancelFunc := context.WithCancel(context.Background())
	done, err := bus.Start(root, fizbus.Default())
	defer func() {
		cancelFunc()
		<-done()
	}()

	require.Nil(t, err)
	require.NotNil(t, done)

	log.Printf("bus setup complete, initiating benchmark")

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			this, _ := context.WithTimeout(root, 2*time.Second)
			receiver := sender.Send(this, "echo", fizbus.Message{Payload: "hello!"})
			_, err := receiver.Receive()

			if err != nil {
				log.Printf("%s ", err.Error())
			}
		}
	})
}
