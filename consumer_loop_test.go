package fizbus

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/streadway/amqp"
)

func TestConsume_RuntimeContextHasBeenCancelled_LoopShouldEnd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	context := NewMockContext(ctrl)
	handler := NewMockhandler(ctrl)

	deliveries := make(chan amqp.Delivery)
	done := make(chan struct{})
	runtime := &runtimeContext{Context: context}

	context.EXPECT().Done().Return(done).Times(1)

	// simulate that the context has been cancelled
	close(done)
	runtime.Shutdown.Add(1)

	consumer := queueConsumerImpl{}
	consumer.Consume(runtime, deliveries, handler)
}

func TestConsume_ItemInQueue_HandlerInvoked(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	context := NewMockContext(ctrl)
	handler := NewMockhandler(ctrl)

	delivery := amqp.Delivery{}
	deliveries := make(chan amqp.Delivery, 1)
	notCancelled := make(chan struct{})
	runtime := &runtimeContext{Context: context}

	context.EXPECT().Done().Return(notCancelled)
	handler.EXPECT().Handle(context, delivery).Times(1)

	// After the delivery is handled we need the loop to exit
	cancelled := make(chan struct{})
	close(cancelled)
	context.EXPECT().Done().Return(cancelled)

	// simulate that we have an AMQP delivery to process
	deliveries <- delivery
	runtime.Shutdown.Add(1)

	consumer := queueConsumerImpl{}
	consumer.Consume(runtime, deliveries, handler)
}
