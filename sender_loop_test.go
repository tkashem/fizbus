package fizbus

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/tkashem/fizbus/upstream"
)

func TestSenderLoop_OutgoingItem_HandlerInvoked(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	channel := NewMockChannel(ctrl)
	sender := NewMocksender(ctrl)
	context := NewMockContext(ctrl)

	outgoing := make(chan *send, 1)
	notCancelled := make(chan struct{})
	sendItem := &send{sender: sender, request: &upstream.Request{}}
	runtime := &runtimeContext{Context: context}

	context.EXPECT().Done().Return(notCancelled)
	sender.EXPECT().Outgoing().Return(outgoing)
	sender.EXPECT().Publish(channel, sendItem.request)

	// After the request has been published, we need to exit the loop
	cancelled := make(chan struct{})
	close(cancelled)
	context.EXPECT().Done().Return(cancelled)   // this will return a closed channel to indicate that the context has timed out
	sender.EXPECT().Outgoing().Return(outgoing) // at this point the outgoing channel will not have any item

	// simulate that we have a request to send out
	outgoing <- sendItem
	runtime.Shutdown.Add(1)

	target := senderLoopImpl{}
	target.Send(runtime, channel, sender)
}

func TestSenderLoop_OutgoingChannelIsClosed_LoopTerminated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	channel := NewMockChannel(ctrl)
	sender := NewMocksender(ctrl)
	context := NewMockContext(ctrl)
	errorCh := NewMockerrorChannel(ctrl)

	errorExpected := errors.New("send loop channel has been closed")
	outgoing := make(chan *send)
	close(outgoing)
	notCancelled := make(chan struct{})
	runtime := &runtimeContext{Context: context, Error: errorCh}

	context.EXPECT().Done().Return(notCancelled)
	sender.EXPECT().Outgoing().Return(outgoing)
	errorCh.EXPECT().Send(gomock.Eq(errorExpected))

	runtime.Shutdown.Add(1)

	target := senderLoopImpl{}
	target.Send(runtime, channel, sender)
}

func TestSenderLoop_ContextTimesOut_LoopTerminated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sender := NewMocksender(ctrl)
	context := NewMockContext(ctrl)

	outgoing := make(chan *send)
	cancelled := make(chan struct{})
	close(cancelled)
	runtime := &runtimeContext{Context: context}

	context.EXPECT().Done().Return(cancelled)
	sender.EXPECT().Outgoing().Return(outgoing)

	runtime.Shutdown.Add(1)

	target := senderLoopImpl{}
	target.Send(runtime, nil, sender)
}
