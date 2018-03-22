package fizbus

import (
	"golang.org/x/net/context"

	"github.com/streadway/amqp"
)

// Once we receive a reply, what needs to happen is encapsulated here. This applies to a sender
type replyHandler struct {
	sender sender
}

func (rh *replyHandler) Handle(c context.Context, d amqp.Delivery) {
	rh.sender.OnReply(d)
	d.Ack(false)
}
