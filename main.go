package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/shaikmuskan27/gocker-linux-runtime/pkg/container"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		run()
	case "init":
		initialize()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gocker run --cpu <shares> --mem <limit> [--rootfs <path>] <command> [args...]")
}

// run parses the flags for the host process, then spawns the child process
// in isolated namespaces.
func run() {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	cpuShares := runCmd.Int("cpu", 0, "CPU shares limit")
	memLimit := runCmd.String("mem", "", "Memory limit (e.g., 50M)")
	rootfs := runCmd.String("rootfs", "./rootfs", "Path to rootfs directory")

	if err := runCmd.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	args := runCmd.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no command specified to run inside container\n")
		os.Exit(1)
	}

	parent(*cpuShares, *memLimit, *rootfs, args)
}

// initialize parses the flags for the container's init process, then runs the child.
func initialize() {
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	rootfs := initCmd.String("rootfs", "./rootfs", "Path to rootfs directory")

	if err := initCmd.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	args := initCmd.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no command specified for container init\n")
		os.Exit(1)
	}

	child(*rootfs, args)
}

// parent starts the container process inside namespace clone flags and waits for synchronization.
func parent(cpuShares int, memLimit string, rootfs string, args []string) {
	// Re-exec self with the "init" command.
	cmdArgs := append([]string{"init", "--rootfs", rootfs}, args...)
	cmd := exec.Command("/proc/self/exe", cmdArgs...)

	// Configure namespace isolation clone flags (UTS, PID, Mount, Network).
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Create synchronization pipe to block the child until Cgroups configuration is complete.
	r, w, err := os.Pipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating synchronization pipe: %v\n", err)
		os.Exit(1)
	}
	defer r.Close()
	defer w.Close()

	// Pass the read end of the pipe as file descriptor 3 to the child.
	cmd.ExtraFiles = []*os.File{r}

	// Start the child process in the background.
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting child process: %v\n", err)
		os.Exit(1)
	}

	// Configure Cgroups settings using the child process's host PID.
	if err := container.ConfigureCgroups(cmd.Process.Pid, cpuShares, memLimit); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to configure cgroups: %v\n", err)
	}

	// Close the write end of the pipe, signaling to the child that Cgroups are configured.
	w.Close()

	// Wait for the child process to complete execution and capture its exit status.
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}

// child executes within isolated namespaces, jails itself, and mounts /proc before running the command.
func child(rootfs string, args []string) {
	// 1. Wait for parent to signal that cgroups configuration is complete.
	// The read end of the pipe is passed as the first extra file descriptor (index 3).
	pipe := os.NewFile(3, "pipe")
	if pipe == nil {
		fmt.Fprintf(os.Stderr, "Error: synchronization pipe not found in child process\n")
		os.Exit(1)
	}

	// Read a single byte. This blocks until the parent closes the write end of the pipe.
	buf := make([]byte, 1)
	_, _ = pipe.Read(buf)
	pipe.Close()

	// 2. Set up the namespace: Hostname, Mount Propagation, Chroot jail, Mount /proc.
	if err := container.SetupNamespace(rootfs); err != nil {
		fmt.Fprintf(os.Stderr, "Namespace setup failed: %v\n", err)
		os.Exit(1)
	}

	// 3. Execute the target command within the isolated container.
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Provide standard fallback environment paths.
	cmd.Env = []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"}

	// Run the container payload.
	runErr := cmd.Run()

	// 4. Elegantly unmount /proc filesystem inside the mount namespace.
	if err := syscall.Unmount("/proc", 0); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to unmount /proc on exit: %v\n", err)
	}

	if runErr != nil {
		fmt.Fprintf(os.Stderr, "Container command failed: %v\n", runErr)
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}