package upstream

import (
	"github.com/streadway/amqp"
)

// NewTable returns a new instance of Table
func NewTable() Table {
	return amqpTableConverterImpl{}
}

type amqpTableConverterImpl struct{}

// To converts standard map of <string, string> to AMQP message header
func (amqpTableConverterImpl) To(h map[string]string) amqp.Table {
	headers := make(amqp.Table)
	for k, v := range h {
		headers[k] = v
	}

	return headers
}

// From converts AMQP message header to standard map of <string, string>
func (amqpTableConverterImpl) From(t amqp.Table) map[string]string {
	headers := make(map[string]string)
	for k, v := range t {
		av, ok := v.(string)

		if !ok {
			// I dont think we can recover from this, violation of protocol which is header value must be encoded as string
			panic("AMQP message header violation: value must be string")
		}

		headers[k] = av
	}

	return headers
}
