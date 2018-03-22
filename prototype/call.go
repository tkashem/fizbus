package simple

import (
  "net/context"
)
type caller interface {
  call(ctx context.Context)
}