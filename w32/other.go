// +build !windows

package w32

import "errors"

// AttachParentConsole returns and error if not on windows platform
func AttachParentConsole() error {
	return errors.New("only available on windows platform")
}
