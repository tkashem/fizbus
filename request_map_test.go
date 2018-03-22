package fizbus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	requests := requestMapImpl{}

	expected := make(chan reply)
	requests.Add("1", expected)

	actual, ok := requests.muxer["1"]
	assert.True(t, ok)
	assert.Equal(t, expected, actual)
}

func TestAdd_Exists_MostRecentValueExpected(t *testing.T) {
	requests := requestMapImpl{}

	ch1 := make(chan reply)
	expected := make(chan reply)
	requests.Add("1", ch1)
	requests.Add("1", expected)

	actual, ok := requests.muxer["1"]
	assert.True(t, ok)
	assert.Equal(t, expected, actual)
}

func TestRemove(t *testing.T) {
	requests := requestMapImpl{}

	expected := make(chan reply)
	requests.Add("1", expected)

	success := requests.Remove("1")

	assert.True(t, success)
	actual, ok := requests.muxer["1"]
	assert.False(t, ok)
	assert.Nil(t, actual)
}

func TestRemove_NotFound_FalseExpected(t *testing.T) {
	requests := requestMapImpl{}

	success := requests.Remove("1")

	assert.False(t, success)
}

func TestRemove_InvokedMultipleTimes_Idempotence(t *testing.T) {
	requests := requestMapImpl{}

	expected := make(chan reply)
	requests.Add("1", expected)

	requests.Remove("1")
	success := requests.Remove("1")

	assert.False(t, success)
}

func TestGet(t *testing.T) {
	requests := requestMapImpl{}

	expected := make(chan reply)
	requests.Add("1", expected)

	actual, ok := requests.Get("1")

	assert.True(t, ok)
	assert.Equal(t, expected, actual)
}

func TestGet_NotFound_FalseExpected(t *testing.T) {
	requests := requestMapImpl{}

	actual, ok := requests.Get("1")

	assert.False(t, ok)
	assert.Nil(t, actual)
}
