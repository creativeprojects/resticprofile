package write

import (
	"time"
)

type AsyncOption func(writer *Async)

// WithWriteInterval sets the interval at which writes happen at least when data is pending
func WithWriteInterval(duration time.Duration) AsyncOption {
	return func(writer *Async) { writer.interval = duration }
}
