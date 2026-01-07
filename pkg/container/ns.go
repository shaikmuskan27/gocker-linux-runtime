package container

import (
	"os"
	"syscall"
)

// SetupNamespace sets up the hostname, chroot, and mounts
func SetupNamespace() {
    // 1. Set Hostname
    must(syscall.Sethostname([]byte("gocker-runtime")))

    // 2. Lock the process into the rootfs folder
    must(syscall.Chroot("/root/rootfs"))
    must(os.Chdir("/"))

    // 3. Mount /proc INSIDE the new jail
    os.MkdirAll("/proc", 0755)
    must(syscall.Mount("proc", "/proc", "proc", 0, ""))
}