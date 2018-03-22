package fizbus

import (
	"fmt"
	"github.com/satori/go.uuid"
)

func newUUID() string {
	return fmt.Sprintf("%s", uuid.NewV4())
}
