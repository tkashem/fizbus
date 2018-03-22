package upstream

import (
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestTo_MapHasItems_ValidTableExpected(t *testing.T) {
	expected := map[string]string{
		"request-id": "X3456",
		"deadline":   "12.00AM",
		"url":        "/identities/",
	}

	c := amqpTableConverterImpl{}
	actual := c.To(expected)

	if len(expected) != len(actual) {
		t.Error("length mismatch")
	}

	for ek, ev := range expected {
		av, _ := actual[ek]
		str, _ := av.(string)
		if ev != str {
			t.Error("item value mismatch mismatch")
		}
	}
}

func TestTo_MapIsNil_EmptyTableExpected(t *testing.T) {
	c := amqpTableConverterImpl{}
	actual := c.To(nil)

	if len(actual) != 0 {
		t.Error("length mismatch")
	}
}

func TestTo_MapIsEmpty_EmptyTableExpected(t *testing.T) {
	c := amqpTableConverterImpl{}
	actual := c.To(map[string]string{})

	if len(actual) != 0 {
		t.Error("length mismatch")
	}
}

func TestFrom_TableHasItems_ValidMapExpected(t *testing.T) {
	expected := amqp.Table{
		"request-id": "X3456",
		"deadline":   "12.00AM",
		"url":        "/identities/",
	}

	c := amqpTableConverterImpl{}
	actual := c.From(expected)

	if len(expected) != len(actual) {
		t.Error("length mismatch")
	}

	for ek, ev := range expected {
		av, _ := actual[ek]
		if ev != av {
			t.Error("item value mismatch mismatch")
		}
	}
}

func TestFrom_TableIsNil_EmptyMapExpected(t *testing.T) {
	c := amqpTableConverterImpl{}
	actual := c.From(nil)

	if len(actual) != 0 {
		t.Error("length mismatch")
	}
}

func TestFrom_TableIsEmpty_EmptyMapExpected(t *testing.T) {
	c := amqpTableConverterImpl{}
	actual := c.From(amqp.Table{})

	if len(actual) != 0 {
		t.Error("length mismatch")
	}
}

func TestFrom_ItemNotString_Panic(t *testing.T) {
	expected := amqp.Table{
		"request-id": 1234,
	}

	c := amqpTableConverterImpl{}
	assert.Panics(t, func() {
		c.From(expected)
	})
}
