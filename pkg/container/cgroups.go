package container

import (
	"os"
	"path/filepath"
	"strconv" // <--- Add this
	// "fmt"    <--- Remove or comment this out
)

// ... rest of your code
// ConfigureCgroups limits the number of processes to prevent fork-bombs
func ConfigureCgroups() {
	// In Cgroup V2, everything is under /sys/fs/cgroup/
	cgroups := "/sys/fs/cgroup/gocker"

	// 1. Create the gocker cgroup directory
	os.Mkdir(cgroups, 0755)

	// 2. Set the max number of processes (PIDs)
	// Notice the path: /sys/fs/cgroup/gocker/pids.max
	must(os.WriteFile(filepath.Join(cgroups, "pids.max"), []byte("20"), 0700))

	// 3. Optional: Remove the cgroup when the process finishes
	// We notify the kernel that this process belongs to this cgroup
	must(os.WriteFile(filepath.Join(cgroups, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
