package fizbus

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/streadway/amqp"
	"github.com/tkashem/fizbus/upstream"
)

func TestRequestHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	context := NewMockContext(ctrl)
	next := NewMockHandler(ctrl)
	converter := NewMockconverter(ctrl)
	channel := NewMockChannel(ctrl)
	acker := NewMockacker(ctrl)

	d := amqp.Delivery{ReplyTo: "me", CorrelationId: "id-1", DeliveryTag: 10} // This is the original amqp delivery we pull off of queue
	in := Message{Payload: "ping"}
	out := Message{Payload: "pong"}
	replyToPublish := &upstream.Request{RoutingKey: d.ReplyTo, CorrelationID: d.CorrelationId, Body: out.Payload, ReplyTo: ""}

	gomock.InOrder(
		converter.EXPECT().Convert(&d).Return(in).Times(1),
		next.EXPECT().Handle(context, in).Return(out).Times(1),
		acker.EXPECT().ack(&d).Return(nil).Times(1),
		channel.EXPECT().Publish(replyToPublish).Return(nil).Times(1),
	)

	runtime := &runtimeContext{}
	target := requestHandler{runtime: runtime, channel: channel, converter: converter, next: next, acker: acker}
	target.Handle(context, d)
}

func TestRequestHandler_PublishError_BusNotified(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	context := NewMockContext(ctrl)
	next := NewMockHandler(ctrl)
	converter := NewMockconverter(ctrl)
	channel := NewMockChannel(ctrl)
	errorCh := NewMockerrorChannel(ctrl)
	acker := NewMockacker(ctrl)

	errorExpected := errors.New("error")
	d := amqp.Delivery{}
	in := Message{Payload: "ping"}
	out := Message{Payload: "pong"}

	converter.EXPECT().Convert(&d).Return(in)
	next.EXPECT().Handle(context, in).Return(out)
	acker.EXPECT().ack(&d).Return(nil)
	channel.EXPECT().Publish(gomock.Any()).Return(errorExpected)
	errorCh.EXPECT().Send(errorExpected)

	runtime := &runtimeContext{Error: errorCh}
	target := requestHandler{runtime: runtime, channel: channel, converter: converter, next: next, acker: acker}
	target.Handle(context, d)
}
