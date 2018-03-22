package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkashem/fizbus"
	"golang.org/x/net/context"
)

type echoHandler struct{}

func (echoHandler) Handle(c context.Context, m fizbus.Message) fizbus.Message {
	return m
}

func TestBus_Send_ReplyExpected(t *testing.T) {
	bus := fizbus.New(endpoint)
	bus.Bind("echo", echoHandler{})
	sender := bus.NewSender("echo-reply")

	root, cancelFunc := context.WithCancel(context.Background())
	done, err := bus.Start(root, fizbus.Default())
	defer func() {
		cancelFunc()
		<-done()
	}()

	require.Nil(t, err)
	require.NotNil(t, done)

	this, timeoutFunc := context.WithTimeout(root, 1*time.Second)
	defer timeoutFunc()

	receiver := sender.Send(this, "echo", fizbus.Message{Payload: "hello!"})
	received, err := receiver.Receive()

	assert.NotNil(t, received)
	assert.Nil(t, err)
	assert.Equal(t, "hello!", received.Payload)
}

func TestBus_NoBinder_TimeoutExpected(t *testing.T) {
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

	this, timeoutFunc := context.WithTimeout(root, 1*time.Second)
	defer timeoutFunc()

	receiver := sender.Send(this, "echo", fizbus.Message{Payload: "hello!"})
	received, err := receiver.Receive()

	assert.Nil(t, received)
	assert.EqualError(t, err, "timed out")
}
