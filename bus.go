package fizbus

import (
	"golang.org/x/net/context"
	"sync"
)

type bus struct {
	once    sync.Once
	starter starter
	runner  busRunner
}

func (b *bus) Bind(queue string, handler Handler) {
	b.starter.NewBinder(queue, handler)
}

func (b *bus) NewSender(queue string) Sender {
	return b.starter.NewSender(queue)
}

func (b *bus) Start(parent context.Context, cfg *Configuration) (Done, error) {
	// todo: on subsequent execution it always returns nil
	var (
		err  error
		done Done
	)

	b.once.Do(
		func() {
			done, err = b.start(parent, cfg)
		})

	return done, err
}

func (b *bus) start(parent context.Context, cfg *Configuration) (Done, error) {
	var (
		startup *startupContext
		runtime *runtimeContext
		err     error
	)

	if startup, err = b.starter.StartupContext(cfg); err != nil {
		return nil, err
	}

	if runtime, err = b.starter.RuntimeContext(parent); err != nil {
		return nil, err
	}

	if err := b.starter.Start(startup, runtime); err != nil {
		return nil, err
	}

	// bus has initialized successfully, let's enter into bus run loop
	go b.runner.Run(runtime)

	return runtime.Done, nil
}
