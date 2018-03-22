package fizbus

import (
	"github.com/tkashem/fizbus/upstream"
)

// Returns a chain of handler as appropriate
type handlerFactoryImpl struct{}

func (handlerFactoryImpl) NewRequestHandler(runtime *runtimeContext, channel upstream.Channel, appHandler Handler) handler {
	return &requestHandler{
		converter: converterImpl{table: upstream.NewTable()},
		channel:   channel,
		next:      appHandler,
		runtime:   runtime,
	}
}

func (handlerFactoryImpl) NewReplyHandler(sender sender, channel upstream.Channel) handler {
	return &replyHandler{
		sender: sender,
	}
}
