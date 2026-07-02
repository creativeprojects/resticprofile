//go:build unix

package shell

import "syscall"

type waitStatus = syscall.WaitStatus
