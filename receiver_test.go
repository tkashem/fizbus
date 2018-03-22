package fizbus

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/tkashem/fizbus/upstream"
)

func TestReceive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	context := NewMockContext(ctrl)
	sender := NewMocksender(ctrl)

	request := &upstream.Request{CorrelationID: "id-1"}
	send := send{sender: sender, request: request}
	errExpected := errors.New("error")
	replyExpected := reply{message: &Message{Payload: "reply"}, err: errExpected}
	replyCh := make(chan reply, 1)
	done := make(chan struct{})

	gomock.InOrder(
		sender.EXPECT().ReplyChannel(request.CorrelationID).Return(replyCh).Times(1),
		context.EXPECT().Done().Return(done).Times(1),
		sender.EXPECT().Cleanup(request.CorrelationID).Times(1),
	)

	// simulate that we have the reply ready to be received
	replyCh <- replyExpected

	target := receiverImpl{context: context, request: &send}
	reply, err := target.Receive()

	assert.Equal(t, errExpected, err)
	assert.Equal(t, replyExpected.message, reply)
}

func TestReceive_ContextCancelled_TimeOutExpected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	context := NewMockContext(ctrl)
	sender := NewMocksender(ctrl)

	request := &upstream.Request{CorrelationID: "id-1"}
	send := send{sender: sender, request: request}
	errExpected := errors.New("timed out")
	replyCh := make(chan reply)
	done := make(chan struct{})

	sender.EXPECT().ReplyChannel(request.CorrelationID).Return(replyCh)
	context.EXPECT().Done().Return(done)
	sender.EXPECT().Cleanup(request.CorrelationID)

	// simulate that the context has timed out
	close(done)

	target := receiverImpl{context: context, request: &send}
	reply, err := target.Receive()

	assert.Nil(t, reply)
	assert.Equal(t, errExpected, err)
}

func TestReceive_ReplyChannelIsNil_ErrorExpected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	context := NewMockContext(ctrl)
	sender := NewMocksender(ctrl)

	request := &upstream.Request{CorrelationID: "id-1"}
	send := send{sender: sender, request: request}
	errExpected := errors.New("reply channel associated with the request has been removed")

	sender.EXPECT().ReplyChannel(request.CorrelationID).Return(nil)

	target := receiverImpl{context: context, request: &send}
	reply, err := target.Receive()

	assert.Nil(t, reply)
	assert.Equal(t, errExpected, err)
}
