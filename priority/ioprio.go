//go:build !linux

package priority

import (
	"errors"
)

// SetIONice does nothing in non-linux OS
func SetIONice(class, value int) error {
	return errors.New("IONice is only supported on Linux")
}
