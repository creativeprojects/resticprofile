package qemu

import (
	"fmt"
	"strconv"
)

type VM struct {
	Architecture Architecture
	CPUCores     int
	Memory       int // in MB
	Disks        []Disk
	Console      bool
}

func (vm VM) Args() []string {
	serial := "file:serial.txt"
	if vm.Console {
		serial = "mon:stdio"
	}
	args := []string{
		// "-daemonize",
		"-m", fmt.Sprintf("%dM", vm.Memory),
		"-smp", strconv.Itoa(vm.CPUCores),
		"-netdev", "user,id=net0,hostfwd=tcp::2222-:22,hostfwd=tcp::3389-:3389",
		"-rtc", "base=localtime,clock=rt",
		"-serial", serial,
		"-boot", "strict=off",
		"-device", "qemu-xhci",
		"-device", "usb-kbd",
		"-device", "usb-tablet",
	}
	args = append(args, vm.Architecture.Args()...)
	for i, disk := range vm.Disks {
		args = append(args, disk.Args(i)...)
	}
	return args
}
