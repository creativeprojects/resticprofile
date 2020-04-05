//+build linux

package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

const ioPrioClassShift = 13
const ioPrioMask = (1 << ioPrioClassShift) - 1

type ioPrioClass int

const (
	ioPrioClassRT ioPrioClass = iota + 1
	ioPrioClassBE
	ioPrioClassIdle
)

const (
	ioPrioWhoProcess = iota + 1
	ioPrioWhoPGRP
	ioPrioWhoUser
)

// This is only displaying the priority of the current process (for testing)
func main() {

	getPriority()
	getIOPriority()

}

func getPriority() {
	pid := unix.Getpid()
	pri, err := unix.Getpriority(unix.PRIO_PROCESS, pid)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Priority: %d\n", pri)
}

func getIOPriority() {
	r1, _, err := unix.Syscall(
		unix.SYS_IOPRIO_GET,
		uintptr(ioPrioWhoPGRP),
		uintptr(os.Getpid()),
		0,
	)
	if err != 0 {
		fmt.Println(err)
		os.Exit(1)
	}
	class := r1 >> ioPrioClassShift
	value := r1 & ioPrioMask
	fmt.Printf("IOPriority: class = %d, value = %d\n", class, value)
}
