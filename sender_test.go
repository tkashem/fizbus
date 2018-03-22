package fizbus

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/tkashem/fizbus/upstream"
)

func Test_replyChGenerator(t *testing.T) {
	ch := replyChGenerator()

	assert.NotNil(t, ch)
	assert.Equal(t, 1, cap(ch))
}

func TestOutgoing(t *testing.T) {
	s := senderImpl{outgoing: make(chan *send)}
	outgoing := s.Outgoing()

	// todo: is there a way to compare equality or identuty of two channels?
	assert.NotNil(t, outgoing)
}

func TestReplyTo(t *testing.T) {
	s := senderImpl{replyTo: "foo"}
	replyTo := s.ReplyTo()

	assert.Equal(t, s.replyTo, replyTo)
}

func TestCleanup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	channels := NewMockrequestMap(ctrl)

	reply := make(chan reply)
	channels.EXPECT().Get("id-1").Return(reply, true).Times(1)
	channels.EXPECT().Remove("id-1").Return(true)

	sender := senderImpl{requestChannels: channels}
	result := sender.Cleanup("id-1")

	_, open := <-reply
	assert.True(t, result)
	assert.False(t, open)
}

func TestCleanup_CorrelationIDNotFound_FalseExpected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	channels := NewMockrequestMap(ctrl)
	channels.EXPECT().Get("id-1").Return(nil, false).Times(1)

	sender := senderImpl{requestChannels: channels}
	result := sender.Cleanup("id-1")

	assert.False(t, result)
}

func TestReplyChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	channels := NewMockrequestMap(ctrl)

	expected := make(chan reply)
	channels.EXPECT().Get("id-1").Return(expected, true).Times(1)

	sender := senderImpl{requestChannels: channels}
	actual := sender.ReplyChannel("id-1")

	assert.NotNil(t, actual)
	assert.EqualValues(t, expected, actual)
}

func TestReplyChannel_CorrelationIDNotFound_NilExpected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	channels := NewMockrequestMap(ctrl)
	channels.EXPECT().Get("id-1").Return(nil, false).Times(1)

	sender := senderImpl{requestChannels: channels}
	ch := sender.ReplyChannel("id-1")

	assert.Nil(t, ch)
}

func TestSender_Send(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	context := NewMockContext(ctrl)
	channels := NewMockrequestMap(ctrl)

	idgen := func() string { return "id-1" }
	myreply := make(chan reply)
	chgen := func() chan reply { return myreply }

	channels.EXPECT().Add("id-1", myreply).Times(1)

	message := Message{Payload: "payload", Headers: map[string]string{"request-id": "X3456"}}
	outgoing := make(chan *send, 1)
	sender := senderImpl{replyTo: "reply-to", idgen: idgen, chgen: chgen, requestChannels: channels, outgoing: outgoing}
	result := sender.Send(context, "to", message)

	assert.NotNil(t, result)
	receiver, ok := result.(receiverImpl)
	assert.True(t, ok)
	assert.Equal(t, context, receiver.context)

	request := &upstream.Request{Body: message.Payload, Headers: message.Headers, CorrelationID: "id-1", RoutingKey: "to", ReplyTo: sender.replyTo}
	assert.Equal(t, request, receiver.request.request)

	expected := &send{sender: &sender, request: request}
	assert.Equal(t, expected, receiver.request)

	// now verify that the request has been sent to the outgoing channels
	actual := <-sender.Outgoing()
	assert.Equal(t, expected, actual)
}

func TestOnReply(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	channels := NewMockrequestMap(ctrl)
	converter := NewMockconverter(ctrl)

	// the channel length set to 1 so that message can be sent to it immediately without being blocked
	ch := make(chan reply, 1)
	d := amqp.Delivery{CorrelationId: "id-1"}
	expected := Message{Payload: "payload"}
	channels.EXPECT().Get("id-1").Return(ch, true).Times(1)
	converter.EXPECT().Convert(&d).Return(expected).Times(1)

	sender := senderImpl{requestChannels: channels, converter: converter}
	sender.OnReply(d)

	reply := <-ch
	assert.NotNil(t, reply)
	assert.Nil(t, reply.err)
	assert.Equal(t, &expected, reply.message)
}

func TestOnreply_CorrelationIDNotFound_NoOP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	channels := NewMockrequestMap(ctrl)
	channels.EXPECT().Get("id-1").Return(nil, false)

	d := amqp.Delivery{CorrelationId: "id-1"}
	sender := senderImpl{requestChannels: channels}
	sender.OnReply(d)
}

func TestPublish_CorrelationIDNotFound_NoOP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	channels := NewMockrequestMap(ctrl)
	channels.EXPECT().Get("id-1").Return(nil, false)

	request := &upstream.Request{CorrelationID: "id-1"}
	sender := senderImpl{requestChannels: channels}
	sender.Publish(nil, request)
}

func TestPublish_AMQPError_ErrorExpected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	channels := NewMockrequestMap(ctrl)
	amqpChannel := NewMockChannel(ctrl)

	// length set to 1 so that it can be sent to immediately
	replyCh := make(chan reply, 1)
	err := errors.New("error")
	request := &upstream.Request{CorrelationID: "id-1"}

	channels.EXPECT().Get("id-1").Return(replyCh, true)
	amqpChannel.EXPECT().Publish(request).Return(err)
	channels.EXPECT().Remove("id-1")

	sender := senderImpl{requestChannels: channels}
	sender.Publish(amqpChannel, request)

	reply := <-replyCh
	assert.NotNil(t, reply)
	assert.Nil(t, reply.message)
	assert.Equal(t, err, reply.err)
}

func TestPublish(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	channels := NewMockrequestMap(ctrl)
	amqpChannel := NewMockChannel(ctrl)

	replyCh := make(chan reply)
	request := &upstream.Request{CorrelationID: "id-1"}

	channels.EXPECT().Get("id-1").Return(replyCh, true)
	amqpChannel.EXPECT().Publish(request).Return(nil)

	sender := senderImpl{requestChannels: channels}
	sender.Publish(amqpChannel, request)
}
