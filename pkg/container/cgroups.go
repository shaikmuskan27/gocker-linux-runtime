package container

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigureCgroups limits the number of processes to prevent fork-bombs
func ConfigureCgroups() {
	cgroups := "/sys/fs/cgroup"
	pids := filepath.Join(cgroups, "pids/gocker")

	os.Mkdir(pids, 0755)
	must(os.WriteFile(filepath.Join(pids, "pids.max"), []byte("50"), 0700))
	// Remove the process from the cgroup when it exits
	must(os.WriteFile(filepath.Join(pids, "notify_on_release"), []byte("1"), 0700))
	must(os.WriteFile(filepath.Join(pids, "cgroup.procs"), []byte(fmt.Sprintf("%d", os.Getpid())), 0700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
