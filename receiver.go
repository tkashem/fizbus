package fizbus

import (
	"errors"
	"golang.org/x/net/context"
)

// This is a helper for sender, upon "send" an instance of this type is returned to the caller
// so that it can invoke the Receive method to wait for the reply.
type receiverImpl struct {
	context context.Context
	request *send
}

func (r receiverImpl) Receive() (*Message, error) {
	request := r.request.request
	sender := r.request.sender

	replyCh := sender.ReplyChannel(request.CorrelationID)
	if replyCh == nil {
		return nil, errors.New("reply channel associated with the request has been removed")
	}

	defer func() {
		sender.Cleanup(request.CorrelationID)
	}()

	select {
	case <-r.context.Done():
		// the sender has timed out or cancelled the request
		return nil, errors.New("timed out")

	case reply := <-replyCh:
		return reply.message, reply.err
	}
}
