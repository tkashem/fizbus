package tests

import (
	"flag"
	"os"
	"testing"
)

var (
	endpoint string
)

func setup() {
	endpointFlagPtr := flag.String("endpoint", "amqp://guest:guest@192.168.99.100:5672/", "rabbitmq endpoint")
	flag.Parse()

	endpoint = *endpointFlagPtr
}

func shutdown() {
}

// TestMain has custom setup and shutdown
func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()

	os.Exit(code)
}
