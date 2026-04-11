package qemu

type Architecture struct {
	Name          string
	Firmware      string
	CPU           string
	MachineType   string
	Hypervisor    string
	NetworkDevice string
}

func (a Architecture) Args() []string {
	args := []string{
		"-machine", "type=" + a.MachineType + ",accel=hvf:kvm:tcg",
		"-cpu", a.CPU,
		"-device", a.NetworkDevice + ",netdev=net0",
		"-bios", a.Firmware,
	}
	return args
}

var (
	ArchitectureARM64 = Architecture{
		Name:          "arm64",
		Firmware:      "edk2-aarch64-code.fd",
		CPU:           "cortex-a57",
		MachineType:   "virt",
		Hypervisor:    "qemu",
		NetworkDevice: "virtio-net-pci",
	}

	ArchitectureAMD64 = Architecture{
		Name:          "x86_64",
		Firmware:      "",
		CPU:           "qemu64",
		MachineType:   "q35",
		Hypervisor:    "qemu",
		NetworkDevice: "e1000",
	}
)
