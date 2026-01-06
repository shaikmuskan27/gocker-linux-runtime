package container

import (
	"os"
	"syscall"
)

// SetupNamespace sets up the hostname, chroot, and mounts
func SetupNamespace() {
	must(syscall.Sethostname([]byte("amazon-internal-core")))
	must(syscall.Chroot("rootfs"))
	must(os.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))
}