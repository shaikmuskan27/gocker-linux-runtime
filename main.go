package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"path/filepath"
)

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("Use: run <command>")
	}
}

func run() {
	fmt.Printf("Parent: Running %v as PID %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running child: %v\n", err)
		os.Exit(1)
	}
}

func child() {
	fmt.Printf("Child: My PID is %d\n", os.Getpid())

	// 1. Set Hostname
	syscall.Sethostname([]byte("amazon-internal-core"))

	// --- CGROUPS: RESOURCE LIMITING ---
	// --- CGROUPS: THE ROBUST VERSION ---
    // On modern Linux (Cgroups v2), everything is in /sys/fs/cgroup
    cgroups := "/sys/fs/cgroup"
    containerGroup := filepath.Join(cgroups, "amazon")
    
    // 1. Create the 'amazon' group folder
    os.Mkdir(containerGroup, 0755)
    
    // 2. Limit the number of processes (PIDs)
    // We use pids.max in the main cgroup folder for v2
    // If this fails, we try the v1 path
    err := os.WriteFile(filepath.Join(containerGroup, "pids.max"), []byte("50"), 0700)
    if err != nil {
        // Fallback for older systems (Cgroups v1)
        os.MkdirAll("/sys/fs/cgroup/pids/amazon", 0755)
        must(os.WriteFile("/sys/fs/cgroup/pids/amazon/pids.max", []byte("50"), 0700))
        must(os.WriteFile("/sys/fs/cgroup/pids/amazon/cgroup.procs", []byte(fmt.Sprintf("%d", os.Getpid())), 0700))
    } else {
        // For Cgroups v2
        must(os.WriteFile(filepath.Join(containerGroup, "cgroup.procs"), []byte(fmt.Sprintf("%d", os.Getpid())), 0700))
    }
    // ----------------------------------

	// 2. CHROOT Logic
	// We get the absolute path to your 'rootfs' folder
	pwd, _ := os.Getwd()
	rootfsPath := filepath.Join(pwd, "rootfs")


	// Change the root to the Alpine folder
	if err := syscall.Chroot(rootfsPath); err != nil {
		panic(fmt.Sprintf("Chroot failed: %v", err))
	}
	if err := syscall.Chdir("/"); err != nil {
		panic(err)
	}

	// 3. Mount the private /proc (This fixes your 'ps' issue!)
	// This tells the kernel: "Only show processes belonging to THIS container"
	syscall.Mount("proc", "proc", "proc", 0, "")

	// 4. Run the command
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error inside child: %v\n", err)
	}

	// Cleanup: Unmount proc when done
	syscall.Unmount("/proc", 0)
}
// Helper function to handle errors. 
// If an error occurs, it prints the error and stops the program (panic).
func must(err error) {
	if err != nil {
		panic(err)
	}
}