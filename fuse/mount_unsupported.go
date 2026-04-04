//go:build windows || netbsd || openbsd || solaris

package fuse

import "errors"

func MountFS(_ string, _ []File) (func(), error) {
	return nil, errors.New("not supported on this platform")
}
