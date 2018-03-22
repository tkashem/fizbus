package fizbus

import (
	"sync"

	"github.com/streadway/amqp"
)

func newMultiAcker(threshold int) acker {
	ma := &multiAcker{threshold: threshold}
	ma.reset()

	return ma
}

type multiAcker struct {
	sync.Mutex

	threshold   int
	current     int
	maxDelivery *amqp.Delivery
}

func (ma *multiAcker) ack(d *amqp.Delivery) error {
	ma.Lock()
	defer ma.Unlock()

	ma.current++
	if d.DeliveryTag > ma.maxDelivery.DeliveryTag {
		ma.maxDelivery = d
	}

	// Is it time to send an ACK? If not, we are done for now.
	if (ma.current % ma.threshold) != 0 {
		return nil
	}

	defer ma.reset()
	return ma.maxDelivery.Ack(true)
}

func (ma *multiAcker) flush() {
	defer ma.reset()

	if ma.maxDelivery != nil {
		ma.maxDelivery.Ack(true)
	}
}

func (ma *multiAcker) reset() {
	ma.current = 0
	ma.maxDelivery = &amqp.Delivery{DeliveryTag: 0}
}
