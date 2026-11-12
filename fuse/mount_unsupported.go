//go:build windows || netbsd || openbsd || solaris || (linux && 386) || (linux && arm) || (linux && mips) || (linux && mipsle)

package fuse

import "errors"

func MountFS(_ string, _ []File) (func(), error) {
	return nil, errors.New("not supported on this platform")
}
