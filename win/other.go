// +build !windows

package win

import "errors"

// RunElevated returns and error if not on windows platform
func RunElevated(port int) error {
	return errors.New("only available on windows platform")
}
