package fizbus

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestRun_AMQPConnectionError_ContextCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cancel := NewMockcanceller(ctrl)
	connection := NewMockConnection(ctrl)
	context := NewMockContext(ctrl)
	parent := NewMockContext(ctrl)
	errorCh := NewMockerrorChannel(ctrl)

	notify := make(chan *amqp.Error, 1)
	done := make(chan struct{})

	runtime := &runtimeContext{Connection: connection, Context: context, Parent: parent, done: done, Error: errorCh, CancelFunc: cancel.Cancel}

	gomock.InOrder(
		connection.EXPECT().Notify().Return(notify).Times(1),
		cancel.EXPECT().Cancel().Times(1),
	)

	parent.EXPECT().Done().Return(nil)
	errorCh.EXPECT().Error().Return(nil)

	// simulate an AMQP connection error by sending to this channel
	notify <- &amqp.Error{}
	runner := busRunnerImpl{}
	runner.Run(runtime)

	// need to assert that the done channel is closed
	_, open := <-done
	assert.False(t, open)
}

func TestRun_ParentContextHasBeenCancelled_ContextCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cancel := NewMockcanceller(ctrl)
	connection := NewMockConnection(ctrl)
	context := NewMockContext(ctrl)
	parent := NewMockContext(ctrl)
	errorCh := NewMockerrorChannel(ctrl)

	parentDone := make(chan struct{}, 1)
	done := make(chan struct{})

	runtime := &runtimeContext{Connection: connection, Context: context, Parent: parent, done: done, Error: errorCh, CancelFunc: cancel.Cancel}

	connection.EXPECT().Notify().Return(nil)
	parent.EXPECT().Done().Return(parentDone)
	errorCh.EXPECT().Error().Return(nil)
	cancel.EXPECT().Cancel().Times(1)

	// simulate that parent context has been cancelled
	close(parentDone)

	runner := busRunnerImpl{}
	runner.Run(runtime)

	// need to assert that the done channel is closed
	_, open := <-done
	assert.False(t, open)
}

func TestRun_BusErrorHappened_ContextCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cancel := NewMockcanceller(ctrl)
	connection := NewMockConnection(ctrl)
	context := NewMockContext(ctrl)
	parent := NewMockContext(ctrl)
	errorCh := NewMockerrorChannel(ctrl)

	busError := make(chan error, 1)
	done := make(chan struct{})

	runtime := &runtimeContext{Connection: connection, Context: context, Parent: parent, done: done, Error: errorCh, CancelFunc: cancel.Cancel}

	connection.EXPECT().Notify().Return(nil)
	parent.EXPECT().Done().Return(nil)
	errorCh.EXPECT().Error().Return(busError)
	cancel.EXPECT().Cancel().Times(1)

	// simulate that a bus error has happened
	busError <- errors.New("error")

	runner := busRunnerImpl{}
	runner.Run(runtime)

	// need to assert that the done channel is closed
	_, open := <-done
	assert.False(t, open)
}

func TestRun_WorkerLoopRunning_Wait(t *testing.T) {
	// todo: This test is not very effective, if the function being tested does not call waitGroup.Wait
	// the test will still pass.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cancel := NewMockcanceller(ctrl)
	connection := NewMockConnection(ctrl)
	context := NewMockContext(ctrl)
	parent := NewMockContext(ctrl)
	errorCh := NewMockerrorChannel(ctrl)

	busError := make(chan error, 1)
	done := make(chan struct{})

	runtime := &runtimeContext{Connection: connection, Context: context, Parent: parent, done: done, Error: errorCh, CancelFunc: cancel.Cancel}

	connection.EXPECT().Notify().Return(nil)
	parent.EXPECT().Done().Return(nil)
	errorCh.EXPECT().Error().Return(busError)
	cancel.EXPECT().Cancel().Times(1)

	busError <- errors.New("error")

	// let's spin up a single worker
	runtime.Shutdown.Add(1)
	go func() {
		defer runtime.Shutdown.Done()
		<-time.After(time.Millisecond * 1)
	}()

	runner := busRunnerImpl{}
	runner.Run(runtime)
}
