package main

import (
	"fmt"      // <--- Make sure this is here!
	"os"
	"os/exec"
	"syscall"
	"github.com/shaikmuskan27/gocker-linux-runtime/pkg/container"
)

// ... rest of your code ...
// ... rest of your code
func main() {
	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	default:
		panic("help")
	}
}

func parent() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func child() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	container.SetupNamespace()
	// container.ConfigureCgroups() // Uncomment if you fixed the paths

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// This allows the shell to find 'ps' and other tools
	cmd.Env = []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"}

	must(cmd.Run())
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}