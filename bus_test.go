package fizbus

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBind_ValidInput_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := NewMockHandler(ctrl)
	starter := NewMockstarter(ctrl)
	starter.EXPECT().NewBinder("request-queue", handler)

	b := bus{starter: starter}
	b.Bind("request-queue", handler)
}

func TestNewSender_ValidInput_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sender := NewMockSender(ctrl)
	starter := NewMockstarter(ctrl)
	starter.EXPECT().NewSender("reply-to").Return(sender)

	b := bus{starter: starter}
	senderActual := b.NewSender("reply-to")

	assert.Equal(t, sender, senderActual)
}

func TestStart_HappyPath_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parent := NewMockContext(ctrl)
	starter := NewMockstarter(ctrl)
	runner := NewMockbusRunner(ctrl)
	startup := &startupContext{}
	runtime := &runtimeContext{}

	configuration := &Configuration{PrefetchCount: 1}
	gomock.InOrder(
		starter.EXPECT().StartupContext(configuration).Return(startup, nil),
		starter.EXPECT().RuntimeContext(parent).Return(runtime, nil),
		starter.EXPECT().Start(startup, runtime).Return(nil),
		runner.EXPECT().Run(runtime),
	)

	b := bus{starter: starter, runner: runner}
	done, err := b.Start(parent, configuration)

	// yielding so that other go routines spawned by the bus can execute.
	// todo: is there a better way to handle this?
	<-time.After(time.Millisecond * 1)

	assert.Nil(t, err)
	assert.NotNil(t, done)
}

func TestStart_CalledMultipleTimes_OnceExpected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parent := NewMockContext(ctrl)
	starter := NewMockstarter(ctrl)
	runner := NewMockbusRunner(ctrl)
	startup := &startupContext{}
	runtime := &runtimeContext{}

	configuration := &Configuration{PrefetchCount: 1}
	starter.EXPECT().StartupContext(configuration).Return(startup, nil).Times(1)
	starter.EXPECT().RuntimeContext(parent).Return(runtime, nil).Times(1)
	starter.EXPECT().Start(startup, runtime).Return(nil).Times(1)
	runner.EXPECT().Run(runtime).Times(1)

	b := bus{starter: starter, runner: runner}
	b.Start(parent, configuration)
	done, err := b.Start(parent, configuration)

	// yielding so that other go routines spawned by the bus can execute.
	// todo: is there a better way to handle this?
	<-time.After(time.Millisecond * 1)

	assert.Nil(t, err)
	assert.Nil(t, done)
}

func TestStart_StartupContextFailed_ErrorExpected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parent := NewMockContext(ctrl)
	starter := NewMockstarter(ctrl)
	errorExpected := errors.New("failed")

	starter.EXPECT().StartupContext(gomock.Any()).Return(nil, errorExpected)

	b := bus{starter: starter}
	done, err := b.Start(parent, nil)

	assert.Nil(t, done)
	assert.Equal(t, errorExpected, err)
}

func TestStart_RuntimeContextFailed_ErrorExpected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parent := NewMockContext(ctrl)
	starter := NewMockstarter(ctrl)
	startup := &startupContext{}
	errorExpected := errors.New("failed")

	starter.EXPECT().StartupContext(gomock.Any()).Return(startup, nil)
	starter.EXPECT().RuntimeContext(parent).Return(nil, errorExpected)

	b := bus{starter: starter}
	done, err := b.Start(parent, nil)

	assert.Nil(t, done)
	assert.Equal(t, errorExpected, err)
}

func TestStart_StartFailed_ErrorExpected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parent := NewMockContext(ctrl)
	starter := NewMockstarter(ctrl)
	startup := &startupContext{}
	runtime := &runtimeContext{}
	errorExpected := errors.New("failed")

	configuration := &Configuration{PrefetchCount: 1}
	starter.EXPECT().StartupContext(configuration).Return(startup, nil)
	starter.EXPECT().RuntimeContext(parent).Return(runtime, nil)
	starter.EXPECT().Start(startup, runtime).Return(errorExpected)

	b := bus{starter: starter}
	done, err := b.Start(parent, configuration)

	assert.Nil(t, done)
	assert.Equal(t, errorExpected, err)
}
