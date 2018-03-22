package fizbus

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestnewErrorChannel(t *testing.T) {
	ech := newErrorChannel()

	channel := ech.Error()
	assert.Equal(t, 1, len(channel))
}

func TestSend(t *testing.T) {
	ech := newErrorChannel()
	err1 := errors.New("error")

	ok := ech.Send(err1)

	err2 := <-ech.Error()

	assert.True(t, ok)
	assert.Equal(t, err1, err2)
}

func TestReceive_NoSender_NilExpected(t *testing.T) {
	ech := newErrorChannel()

	err := ech.Receive()

	assert.Nil(t, err)
}

func TestReceive_ChannelNotEmpy_Success(t *testing.T) {
	ech := newErrorChannel()
	err1 := errors.New("error")

	ech.Send(err1)
	err2 := ech.Receive()

	assert.Equal(t, err1, err2)
}

func TestReceive_ChannelEmpy_NoBlockExpected(t *testing.T) {
	ech := newErrorChannel()

	ech.Receive()
	ech.Receive()
	err := ech.Receive()

	assert.Nil(t, err)
}

func TestSend_NoListener_NoBlockExpected(t *testing.T) {
	ech := newErrorChannel()
	err := errors.New("error")

	ech.Send(err)
	ech.Send(err)
	ok := ech.Send(err)

	assert.False(t, ok)
}
