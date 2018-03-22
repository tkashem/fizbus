package fizbus

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestConvert(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	table := NewMockTable(ctrl)
	target := converterImpl{table: table}

	headerExpected := map[string]string{
		"foo": "1",
		"bar": "2",
	}

	delivery := amqp.Delivery{
		Body: []byte{'h', 'e', 'l', 'l', 'o'},
		Headers: amqp.Table{
			"foo": "1",
			"bar": "2",
		},
	}

	table.EXPECT().From(delivery.Headers).Return(headerExpected)

	message := target.Convert(&delivery)

	assert.Equal(t, headerExpected, message.Headers)
	assert.Equal(t, "hello", message.Payload)
}
