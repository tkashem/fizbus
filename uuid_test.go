package fizbus

import (
	"github.com/satori/go.uuid"
	"testing"
)

func Test_newUUID(t *testing.T) {
	_, err := uuid.FromString(newUUID())
	if err != nil {
		t.Errorf("expected a valid uuid back from newUUID")
	}
}
