package fizbus

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestAck_BelowThreshold_NotAcked(t *testing.T) {
	target := multiAcker{threshold: 2, maxDelivery: &amqp.Delivery{}}

	delivery := &amqp.Delivery{DeliveryTag: 100}
	err := target.ack(delivery)

	assert.NoError(t, err)
	assert.Equal(t, 1, target.current)
	assert.Equal(t, delivery, target.maxDelivery)
}

func TestAck_AboveThreshold_Acked(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ack := NewMockAcknowledger(ctrl)

	target := multiAcker{threshold: 1, maxDelivery: &amqp.Delivery{}}
	delivery := &amqp.Delivery{DeliveryTag: 100, Acknowledger: ack}
	ack.EXPECT().Ack(delivery.DeliveryTag, true).Return(nil).Times(1)

	err := target.ack(delivery)

	assert.NoError(t, err)
	assert.Equal(t, 0, target.current)
	assert.Nil(t, target.maxDelivery)
}

func TestAck_Multi_Acked(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ack := NewMockAcknowledger(ctrl)

	target := multiAcker{threshold: 5, maxDelivery: &amqp.Delivery{}}

	deliveries := []*amqp.Delivery{
		&amqp.Delivery{DeliveryTag: 2, Acknowledger: ack},
		&amqp.Delivery{DeliveryTag: 1, Acknowledger: ack},
		&amqp.Delivery{DeliveryTag: 5, Acknowledger: ack},
		&amqp.Delivery{DeliveryTag: 3, Acknowledger: ack},
		&amqp.Delivery{DeliveryTag: 4, Acknowledger: ack},
	}

	ack.EXPECT().Ack(uint64(5), true).Return(nil).Times(1)

	for _, d := range deliveries {
		err := target.ack(d)
		assert.NoError(t, err)
	}

	assert.Equal(t, 0, target.current)
	assert.Nil(t, target.maxDelivery)
}

func TestAck_AckError_ErrorExpected(t *testing.T) {
	target := multiAcker{threshold: 1, maxDelivery: &amqp.Delivery{}}
	delivery := &amqp.Delivery{DeliveryTag: 100, Acknowledger: nil}

	err := target.ack(delivery)

	assert.EqualError(t, err, "can't ACK delivery, acknowledger set to nil")
	assert.Equal(t, 1, target.current)
	assert.Equal(t, delivery, target.maxDelivery)
}
