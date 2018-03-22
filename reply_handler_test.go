package fizbus

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/streadway/amqp"
)

func TestReplyHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	context := NewMockContext(ctrl)
	sender := NewMocksender(ctrl)
	ack := NewMockAcknowledger(ctrl)

	d := amqp.Delivery{Acknowledger: ack, DeliveryTag: 10} // This is the original amqp delivery we pull off of queue

	gomock.InOrder(
		sender.EXPECT().OnReply(d).Times(1),
		ack.EXPECT().Ack(d.DeliveryTag, false).Return(nil).Times(1),
	)

	target := replyHandler{sender: sender}
	target.Handle(context, d)
}
