package qemu

import "fmt"

type Disk struct {
	File   string
	Format string // qcow2, raw, vmdk, vhdx, etc.
}

func (d Disk) Args(id int) []string {
	return []string{
		"-device", fmt.Sprintf("virtio-blk-pci,drive=drive%d,bootindex=%d", id, id),
		"-drive", fmt.Sprintf("if=none,file=%s,id=drive%d,format=%s", d.File, id, d.Format),
	}
}
