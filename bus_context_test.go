package fizbus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDone(t *testing.T) {
	expected := make(chan struct{})
	target := runtimeContext{done: expected}

	done := target.Done()

	assert.NotNil(t, done)
	assert.Equal(t, 0, cap(done))

	// todo: need to find a way to compare channel identity or equality
	// assert.Equal(t, expected, done)
}
