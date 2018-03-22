package fizbus

import (
	"sync"

	"golang.org/x/net/context"

	"github.com/tkashem/fizbus/upstream"
)

// This context is passed around to functions that need to synchronize with the bus while in running mode
// This type is not to be confused with net/context type.
// This is an internal type that works as a placeholder for objects that are needed while the bus is in running mode
type runtimeContext struct {
	// Any error encountered is sent to this channel so that the bus can take appropriate measure
	Error errorChannel

	// The bus passes this net/context object to its worker go routines
	// when the bus shuts down, it cancels this context so that all worker go routines are notified.
	Context context.Context

	// This is the cancel function associated with the Context object, when invoked it cancels the context.
	CancelFunc context.CancelFunc

	// This is the net/context owned by the app, it is passed down to the bus when it is started.
	// if this context times out or is cancelled the bus initiates a shut down process
	Parent context.Context

	// The bus waits on this and allows all worker go routines to finish
	Shutdown sync.WaitGroup

	// The underlying connection to AMQP broker
	Connection upstream.Connection

	// When the bus has completely shut down, this channel is closed so that app waiting can gracefully shutdown
	done chan struct{}
}

// Done returns a channel that will be closed when the bus and all it worker go routines have completely shut down
func (rc *runtimeContext) Done() <-chan struct{} {
	return rc.done
}

// This context is passed around to functions that need to synchronize with the bus while in starting mode
// This type is not to be confused with net/context type.
// Once the bus has started successfully there is no more use of this object
type startupContext struct {
	// Any error encountered is sent to this channel while in starting mode
	// so that the bus can take appropriate measure
	Error errorChannel

	// Before the bus enters into running mode, it waits on all initializers to finish
	Ready sync.WaitGroup

	// Bus start up configuration
	configuration *Configuration
}
