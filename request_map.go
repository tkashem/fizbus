package fizbus

import (
	"sync"
)

type requestMapImpl struct {
	sync.Mutex
	muxer map[string]chan reply
}

func (m *requestMapImpl) Add(key string, ch chan reply) {
	m.Lock()
	defer m.Unlock()

	if m.muxer == nil {
		m.muxer = make(map[string]chan reply)
	}
	m.muxer[key] = ch
}

func (m *requestMapImpl) Remove(key string) bool {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.muxer[key]; ok {
		delete(m.muxer, key)
		return true
	}

	return false
}

func (m *requestMapImpl) Get(key string) (chan reply, bool) {
	m.Lock()
	defer m.Unlock()

	ch, ok := m.muxer[key]
	return ch, ok
}
