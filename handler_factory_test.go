package fizbus

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewRequestHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	channel := NewMockChannel(ctrl)
	appHandler := NewMockHandler(ctrl)

	runtime := &runtimeContext{}
	target := handlerFactoryImpl{}
	handler := target.NewRequestHandler(runtime, channel, appHandler)

	assert.NotNil(t, handler)
	assert.IsType(t, new(requestHandler), handler)

	rh, _ := handler.(*requestHandler)

	assert.Equal(t, channel, rh.channel)
	assert.Equal(t, appHandler, rh.next)
	assert.Equal(t, runtime, rh.runtime)

	_, ok := rh.converter.(converterImpl)
	assert.True(t, ok)
}

func TestNewReplyHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	channel := NewMockChannel(ctrl)
	sender := NewMocksender(ctrl)

	target := handlerFactoryImpl{}
	handler := target.NewReplyHandler(sender, channel)

	assert.NotNil(t, handler)
	assert.IsType(t, new(replyHandler), handler)

	rh, _ := handler.(*replyHandler)
	assert.Equal(t, sender, rh.sender)
}
