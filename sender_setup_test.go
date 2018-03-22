package fizbus

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/streadway/amqp"
)

// todo: this test fails occasionally, fix this.
func TestSenderSetup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	connection := NewMockConnection(ctrl)
	channel := NewMockChannel(ctrl)
	errorCh := NewMockerrorChannel(ctrl)
	chain := NewMockhandlerFactory(ctrl)
	handler := NewMockhandler(ctrl)
	consumer := NewMockqueueConsumer(ctrl)
	senderLoop := NewMocksenderLoop(ctrl)
	sender := NewMocksender(ctrl)

	startup := &startupContext{Error: errorCh}
	runtime := &runtimeContext{Connection: connection}
	deliveries := make(chan amqp.Delivery)
	gomock.InOrder(
		connection.EXPECT().Channel().Return(channel, nil).Times(1),
		sender.EXPECT().ReplyTo().Return("reply-to").Times(1),
		channel.EXPECT().Consume("reply-to", true).Return(deliveries, nil).Times(1),
		chain.EXPECT().NewReplyHandler(sender, channel).Return(handler).Times(1),
		consumer.EXPECT().Consume(runtime, gomock.Any(), handler).Times(1),
		senderLoop.EXPECT().Send(runtime, channel, sender).Times(1),
	)

	startup.Ready.Add(1)
	setup := senderSetup{sender: sender, chain: chain, consumer: consumer, senderLoop: senderLoop}
	setup.initialize(startup, runtime)

	// This gives the go routine a chance to executed
	// todo: is there a better way to do this?
	<-time.After(time.Millisecond * 10)
}

func TestSenderSetup_ChannelConsumeError_BusNotified(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	connection := NewMockConnection(ctrl)
	channel := NewMockChannel(ctrl)
	errorCh := NewMockerrorChannel(ctrl)
	sender := NewMocksender(ctrl)

	errorExpected := errors.New("error")
	startup := &startupContext{Error: errorCh}
	runtime := &runtimeContext{Connection: connection}

	connection.EXPECT().Channel().Return(channel, nil)
	sender.EXPECT().ReplyTo().Return("reply-to")
	channel.EXPECT().Consume("reply-to", true).Return(nil, errorExpected)
	errorCh.EXPECT().Send(errorExpected)

	startup.Ready.Add(1)
	setup := senderSetup{sender: sender}
	setup.initialize(startup, runtime)
}

func TestSenderSetup_ChannelCreateError_BusNotified(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	connection := NewMockConnection(ctrl)
	errorCh := NewMockerrorChannel(ctrl)

	errorExpected := errors.New("error")
	startup := &startupContext{Error: errorCh}
	runtime := &runtimeContext{Connection: connection}

	connection.EXPECT().Channel().Return(nil, errorExpected)
	errorCh.EXPECT().Send(errorExpected)

	startup.Ready.Add(1)
	setup := senderSetup{}
	setup.initialize(startup, runtime)
}
