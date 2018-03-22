package fizbus

import (
	"golang.org/x/net/context"

	"github.com/streadway/amqp"
	"github.com/tkashem/fizbus/upstream"
)

type idgen func() string
type chgen func() chan reply

// generates a reply channel for each request
func replyChGenerator() chan reply {
	return make(chan reply, 1)
}

// reply we get off of a reply queue
type reply struct {
	message *Message
	err     error
}

type send struct {
	sender  sender
	request *upstream.Request
}

type senderImpl struct {
	idgen    idgen
	chgen    chgen
	replyTo  string
	outgoing chan *send

	requestChannels requestMap
	converter       converter
}

func (s *senderImpl) Outgoing() <-chan *send {
	return s.outgoing
}

func (s *senderImpl) ReplyTo() string {
	return s.replyTo
}

func (s *senderImpl) Cleanup(correlationID string) bool {
	replyCh, exists := s.requestChannels.Get(correlationID)
	if !exists {
		return false
	}

	s.requestChannels.Remove(correlationID)
	close(replyCh)

	return true
}

func (s *senderImpl) ReplyChannel(correlationID string) <-chan reply {
	replyCh, exists := s.requestChannels.Get(correlationID)
	if !exists {
		return nil
	}

	return replyCh
}

func (s *senderImpl) Send(context context.Context, to string, msg Message) Receiver {
	cid := s.idgen()
	reply := s.chgen()
	s.requestChannels.Add(cid, reply)

	request := &upstream.Request{Body: msg.Payload, Headers: msg.Headers, CorrelationID: cid, RoutingKey: to, ReplyTo: s.replyTo}
	out := &send{sender: s, request: request}

	s.outgoing <- out

	return receiverImpl{context: context, request: out}
}

func (s *senderImpl) OnReply(d amqp.Delivery) {
	ch, exists := s.requestChannels.Get(d.CorrelationId)
	if exists {
		message := s.converter.Convert(&d)
		ch <- reply{message: &message, err: nil}
	}
}

func (s *senderImpl) Publish(channel upstream.Channel, request *upstream.Request) {
	replyCh, exists := s.requestChannels.Get(request.CorrelationID)

	if !exists {
		// maybe we should error management here?
		return
	}

	err := channel.Publish(request)
	if err != nil {
		s.requestChannels.Remove(request.CorrelationID)
		replyCh <- reply{message: nil, err: err}
	}
}
