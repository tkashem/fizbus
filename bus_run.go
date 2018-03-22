package fizbus

import (
	"fmt"
)

// While the bus is in running mode, it enters into the following block as it needs to accomodate the following -
// 1. If the underlying connection to AMQP broker is closed then we want to notify all worker go routines and initiate a shutdown process
// 2. Check if the app has initiated a shut down process, in that case we want to initiate a shutdown process
// 3. If an error has occurred, the initiates a shutdown process
// todo: right now the bus shuts down on any type of error, maybe we can inspect the error and if initiate shutdown only if the error is fatal
type busRunnerImpl struct{}

func (br *busRunnerImpl) Run(runtime *runtimeContext) {
	// the bus will wait until all workers have exited
	defer func() {
		runtime.Shutdown.Wait()
		close(runtime.done)
		fmt.Println("bus: bus run loop exited")
	}()

	// If the connection is closed or any error happens then we want to initiate a shut down process
	connectionErrorChannel := runtime.Connection.Notify()

	select {
	case <-connectionErrorChannel:
		// cancel the bus context so all workers exit their loop.
		runtime.CancelFunc()
		fmt.Println("bus: connection closed, due to error; shutting down")
		return
	case <-runtime.Parent.Done():
		// the application context has been cancelled, so let's notify all workers
		runtime.CancelFunc()
		fmt.Println("bus: parent has timed out; shutting down")
		return
	case err := <-runtime.Error.Error():
		// A worker has notified of an error, let's shutdown
		// todo: shutting down on all errors may be aggressive, need to improve it later
		fmt.Printf("app: type: %T; value: %q\n", err.Error(), err.Error())
		runtime.CancelFunc()
		return
	}
}
