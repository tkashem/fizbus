package fizbus

import (
	"golang.org/x/net/context"

	"github.com/streadway/amqp"
	"github.com/tkashem/fizbus/upstream"
)

// Once we receive a request off a queue what should happen is encapsulated below.
// This applies to a binder.
type requestHandler struct {
	runtime   *runtimeContext
	channel   upstream.Channel
	converter converter
	acker     acker
	next      Handler
}

func (rh *requestHandler) Handle(c context.Context, d amqp.Delivery) {
	message := rh.converter.Convert(&d)

	reply := rh.next.Handle(c, message)

	// todo: we are ignoring the error here
	_ = rh.acker.ack(&d)

	out := &upstream.Request{RoutingKey: d.ReplyTo, CorrelationID: d.CorrelationId, Body: reply.Payload, ReplyTo: ""}
	err := rh.channel.Publish(out)
	if err != nil {
		rh.runtime.Error.Send(err)
	}
}
