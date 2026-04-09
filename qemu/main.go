package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	ctx := context.Background()
	err := runFreeBSD(ctx, "qemu/disk_images/freebsd-15.0-arm64.qcow2")
	if err != nil {
		log.Fatal(err)
	}
}

// func prepareFreeBSD(ctx context.Context, version string) error {
// 	home, err := os.UserHomeDir()
// 	if err != nil {
// 		return fmt.Errorf("os.UserHomeDir: %w", err)
// 	}
// 	currentDir, err := os.Getwd()
// 	if err != nil {
// 		return fmt.Errorf("os.Getwd: %w", err)
// 	}
// 	err = unXZ(ctx,
// 		filepath.Join(home, "Downloads", version+".xz"),
// 		filepath.Join(currentDir, version),
// 	)
// 	if err != nil {
// 		return fmt.Errorf("unXZ: %w", err)
// 	}
// 	return nil
// }

// func unXZ(ctx context.Context, source, target string) error {
// 	cmd := exec.CommandContext(ctx, "xz",
// 		"--verbose",
// 		"--keep",
// 		"--decompress",
// 		source,
// 	)
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	err := cmd.Run()
// 	if err != nil {
// 		return fmt.Errorf("xz: %w", err)
// 	}

// 	cmd = exec.CommandContext(ctx, "mv", "-v", strings.TrimSuffix(source, ".xz"), target)
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	err = cmd.Run()
// 	if err != nil {
// 		return fmt.Errorf("mv: %w", err)
// 	}
// 	return nil
// }

func runFreeBSD(ctx context.Context, diskImage string) error {
	cmd := exec.CommandContext(ctx, "qemu-system-aarch64",
		"-m", "4096M", "-M", "virt,accel=hvf",
		"-cpu", "cortex-a57",
		"-bios", "edk2-aarch64-code.fd",
		"-rtc", "base=localtime,clock=rt",
		"-nographic",
		"-serial", "mon:stdio",
		"-device", "qemu-xhci",
		"-device", "usb-kbd",
		"-device", "usb-tablet",
		"-device", "virtio-net,netdev=net0",
		"-netdev", "user,id=net0,hostfwd=tcp::2222-:22,hostfwd=tcp::3389-:3389",
		"-drive", "if=virtio,file="+diskImage+",format=qcow2,cache=writethrough",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("qemu-system-aarch64: %w", err)
	}
	return nil
}
