package fizbus

// This is to mock CancelFunc for net/context
type canceller interface {
	Cancel()
}
