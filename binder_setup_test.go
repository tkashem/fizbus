package fizbus

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/streadway/amqp"
)

func TestBinderSetup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	connection := NewMockConnection(ctrl)
	channel := NewMockChannel(ctrl)
	factory := NewMockhandlerFactory(ctrl)
	handler := NewMockhandler(ctrl)
	consumer := NewMockqueueConsumer(ctrl)
	userHandler := NewMockHandler(ctrl)

	startup := &startupContext{configuration: &Configuration{PrefetchCount: 3}}
	runtime := &runtimeContext{Connection: connection}
	deliveries := make(chan amqp.Delivery)

	gomock.InOrder(
		connection.EXPECT().Channel().Return(channel, nil).Times(1),
		channel.EXPECT().Prefetch(startup.configuration.PrefetchCount).Return(nil).Times(1),
		channel.EXPECT().Consume("ping", false).Return(deliveries, nil).Times(1),
		factory.EXPECT().NewRequestHandler(runtime, channel, userHandler).Return(handler).Times(1),
		consumer.EXPECT().Consume(runtime, gomock.Any(), handler).Times(3), // todo: remove gomock.Any()?
	)

	startup.Ready.Add(1)
	target := binderSetup{routingKey: "ping", factory: factory, consumer: consumer, userHandler: userHandler}
	target.initialize(startup, runtime)

	// This gives the go routine a chance to executed
	// todo: is there a better way to do this?
	<-time.After(time.Millisecond * 1)
}

func TestBinderSetup_ConsumeFailed_BusNotified(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	connection := NewMockConnection(ctrl)
	errorCh := NewMockerrorChannel(ctrl)
	channel := NewMockChannel(ctrl)

	startup := &startupContext{Error: errorCh, configuration: &Configuration{PrefetchCount: 3}}
	runtime := &runtimeContext{Connection: connection}
	errorExpected := errors.New("error")

	connection.EXPECT().Channel().Return(channel, nil)
	channel.EXPECT().Prefetch(startup.configuration.PrefetchCount).Return(nil)
	channel.EXPECT().Consume("ping", false).Return(nil, errorExpected)
	errorCh.EXPECT().Send(errorExpected)

	startup.Ready.Add(1)
	target := binderSetup{routingKey: "ping"}
	target.initialize(startup, runtime)
}

func TestBinderSetup_PrefetchSetFailed_BusNotified(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	connection := NewMockConnection(ctrl)
	errorCh := NewMockerrorChannel(ctrl)
	channel := NewMockChannel(ctrl)

	startup := &startupContext{Error: errorCh, configuration: &Configuration{PrefetchCount: 3}}
	runtime := &runtimeContext{Connection: connection}
	errorExpected := errors.New("error")

	connection.EXPECT().Channel().Return(channel, nil)
	channel.EXPECT().Prefetch(startup.configuration.PrefetchCount).Return(errorExpected)
	errorCh.EXPECT().Send(errorExpected)

	startup.Ready.Add(1)
	target := binderSetup{routingKey: "ping"}
	target.initialize(startup, runtime)
}

func TestBinderSetup_ChannelCreateFailed_BusNotified(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	connection := NewMockConnection(ctrl)
	errorCh := NewMockerrorChannel(ctrl)

	startup := &startupContext{Error: errorCh}
	runtime := &runtimeContext{Connection: connection}
	errorExpected := errors.New("error")

	connection.EXPECT().Channel().Return(nil, errorExpected)
	errorCh.EXPECT().Send(errorExpected)

	startup.Ready.Add(1)
	target := binderSetup{routingKey: "ping"}
	target.initialize(startup, runtime)
}
