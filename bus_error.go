package fizbus

func newErrorChannel() errorChannel {
	ch := make(errorChannelImpl, 1)
	return ch
}

// This is a encapsulation around a channel of error, it allows to send and receive in a non blocking way
// This is an unbounded channel and any number of errors can be sent to this channel by any number of go routines
// By necessity, we dont want the send operation to be blocking.
// We dont want a go routine to block while sending to this channel
type errorChannelImpl chan error

func (e errorChannelImpl) Error() <-chan error {
	return e
}

func (e errorChannelImpl) Send(err error) bool {
	select {
	case e <- err:
		return true
	default:
		return false
	}
}

func (e errorChannelImpl) Receive() error {
	select {
	case err := <-e:
		return err
	default:
		return nil
	}
}
