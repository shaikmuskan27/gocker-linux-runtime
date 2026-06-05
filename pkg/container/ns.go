package container

import (
	"fmt"
	"os"
	"syscall"
)

// SetupNamespace sets up the hostname, chroot jail, mount isolation, and mounts /proc.
func SetupNamespace(rootfs string) error {
	// 1. Set Hostname
	if err := syscall.Sethostname([]byte("gocker-runtime")); err != nil {
		return fmt.Errorf("setting hostname failed: %w", err)
	}

	// 2. Make the mount namespace private to prevent container mounts from leaking to host.
	// This is a critical isolation step before performing any chroot or mounting.
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("making mount namespace private failed: %w", err)
	}

	// 3. Lock the process into the rootfs folder via chroot jail.
	if err := syscall.Chroot(rootfs); err != nil {
		return fmt.Errorf("chroot to %s failed: %w", rootfs, err)
	}
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("chdir to / inside jail failed: %w", err)
	}

	// 4. Mount /proc INSIDE the new jail so that tools like ps or top function properly.
	if err := os.MkdirAll("/proc", 0755); err != nil {
		return fmt.Errorf("mkdir /proc failed: %w", err)
	}
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("mounting /proc failed: %w", err)
	}

	return nil
}