package fizbus

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestStartupContext(t *testing.T) {
	configuration := &Configuration{PrefetchCount: 3}
	target := starterImpl{}
	startup, err := target.StartupContext(configuration)

	assert.NotNil(t, startup)
	assert.Nil(t, err)
	assert.NotNil(t, startup.Error)
	assert.Equal(t, configuration, startup.configuration)
	assert.IsType(t, (errorChannelImpl)(nil), startup.Error)
}

func TestRuntimeContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dialer := NewMockDialer(ctrl)
	parent := NewMockContext(ctrl)
	connection := NewMockConnection(ctrl)

	dialer.EXPECT().Dial().Return(connection, nil)
	parent.EXPECT().Done().Return(make(chan struct{}))

	target := starterImpl{dialer: dialer}
	runtime, err := target.RuntimeContext(parent)

	assert.NotNil(t, runtime)
	assert.Nil(t, err)
	assert.NotNil(t, runtime.Error)
	assert.IsType(t, (errorChannelImpl)(nil), runtime.Error)
	assert.Equal(t, parent, runtime.Parent)
	assert.Equal(t, connection, runtime.Connection)

	assert.NotNil(t, runtime.Context) // todo: can we do more strict validation than just checking for nil
	assert.NotNil(t, runtime.CancelFunc)

	assert.NotNil(t, runtime.done)
	assert.Equal(t, 0, cap(runtime.done))
}

func TestRuntimeContext_ConnectFailed_ErrorExpected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dialer := NewMockDialer(ctrl)
	parent := NewMockContext(ctrl)

	errorExpected := errors.New("error")
	dialer.EXPECT().Dial().Return(nil, errorExpected)

	target := starterImpl{dialer: dialer}
	runtime, err := target.RuntimeContext(parent)

	assert.NotNil(t, err)
	assert.Nil(t, runtime)
	assert.Equal(t, errorExpected, err)
}

func TestNewBinder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := NewMockHandler(ctrl)

	target := starterImpl{}
	target.NewBinder("myrouting-key", handler)

	assert.Equal(t, 1, len(target.initializers))
	initializer := target.initializers[0]

	assert.NotNil(t, initializer)

	setup, ok := initializer.(*binderSetup)
	assert.True(t, ok)

	assert.Equal(t, handler, setup.userHandler)
	_, ok = setup.consumer.(queueConsumerImpl)
	assert.True(t, ok)
	_, ok = setup.factory.(handlerFactoryImpl)
	assert.True(t, ok)
}

func TestNewSender(t *testing.T) {
	myroutegen := func(s string) string {
		return s
	}

	target := starterImpl{routing: myroutegen}
	actual := target.NewSender("myrouting-key")

	assert.NotNil(t, actual)
	sender, ok := actual.(*senderImpl)
	assert.True(t, ok)
	assert.Equal(t, "myrouting-key", sender.replyTo)
	// assert.Equal(t, newUUID, sender.idgen)
	// assert.Equal(t, replyChGenerator, sender.chgen)
	assert.Equal(t, 0, cap(sender.outgoing))

	_, ok = sender.converter.(converterImpl)
	assert.True(t, ok)
	_, ok = sender.requestChannels.(*requestMapImpl)
	assert.True(t, ok)

	assert.Equal(t, 1, len(target.initializers))
	initializer := target.initializers[0]

	assert.NotNil(t, initializer)

	setup, ok := initializer.(senderSetup)
	assert.True(t, ok)

	assert.Equal(t, sender, setup.sender)
	_, ok = setup.chain.(handlerFactoryImpl)
	assert.True(t, ok)
	_, ok = setup.consumer.(queueConsumerImpl)
	assert.True(t, ok)
	_, ok = setup.senderLoop.(*senderLoopImpl)
	assert.True(t, ok)
}

func TestStart_NoInitializer_NoOP(t *testing.T) {
	target := starterImpl{}

	err := target.Start(nil, nil)

	assert.Nil(t, err)
}

func TestStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	in1 := NewMockinitializer(ctrl)
	in2 := NewMockinitializer(ctrl)
	errorCh := NewMockerrorChannel(ctrl)

	startup := &startupContext{Error: errorCh}
	runtime := &runtimeContext{}

	in1.EXPECT().initialize(startup, runtime)
	in2.EXPECT().initialize(startup, runtime)
	errorCh.EXPECT().Receive().Return(nil)

	target := starterImpl{}
	target.append(in1)
	target.append(in2)

	go func() {
		startup.Ready.Done()
		startup.Ready.Done()
	}()

	err := target.Start(startup, runtime)

	<-time.After(time.Millisecond * 2)
	assert.Nil(t, err)
}

func TestStart_SetupFails_ErrorExpected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	in := NewMockinitializer(ctrl)
	errorCh := NewMockerrorChannel(ctrl)

	startup := &startupContext{Error: errorCh}
	runtime := &runtimeContext{}
	errorExpected := errors.New("error")

	in.EXPECT().initialize(startup, runtime)
	errorCh.EXPECT().Receive().Return(errorExpected)

	target := starterImpl{}
	target.append(in)

	go func() {
		startup.Ready.Done()
	}()

	err := target.Start(startup, runtime)

	<-time.After(time.Millisecond * 2)
	assert.Equal(t, errorExpected, err)
}
