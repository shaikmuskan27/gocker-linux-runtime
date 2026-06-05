package container

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ConfigureCgroups configures the CPU and memory limits for the given process PID.
// It directly interacts with the v1 or v2 cgroups filesystem under /sys/fs/cgroup.
func ConfigureCgroups(pid int, cpuShares int, memLimit string) error {
	// Parse memory limit to bytes if provided.
	var memBytes string
	if memLimit != "" {
		var err error
		memBytes, err = parseMemoryLimit(memLimit)
		if err != nil {
			return err
		}
	}

	// Determine if cgroups v1 is active.
	isV1Memory := false
	isV1CPU := false

	memPath := "/sys/fs/cgroup/memory"
	if _, err := os.Stat(memPath); err == nil {
		isV1Memory = true
	}

	cpuPath := "/sys/fs/cgroup/cpu"
	if _, err := os.Stat(cpuPath); err != nil {
		// Fallback to unified cpu,cpuacct controller in some v1 configurations.
		if _, err := os.Stat("/sys/fs/cgroup/cpu,cpuacct"); err == nil {
			cpuPath = "/sys/fs/cgroup/cpu,cpuacct"
			isV1CPU = true
		}
	} else {
		isV1CPU = true
	}

	pidStr := strconv.Itoa(pid)

	// If at least one cgroups v1 controller path exists, use v1.
	if isV1Memory || isV1CPU {
		// Configure CPU Shares (v1)
		if isV1CPU && cpuShares > 0 {
			gockerCPUPath := filepath.Join(cpuPath, "gocker")
			if err := os.MkdirAll(gockerCPUPath, 0755); err != nil {
				return fmt.Errorf("failed to create cgroup cpu dir: %w", err)
			}
			if err := os.WriteFile(filepath.Join(gockerCPUPath, "cpu.shares"), []byte(strconv.Itoa(cpuShares)), 0644); err != nil {
				return fmt.Errorf("failed to write cpu.shares: %w", err)
			}
			// Write process PID to tasks or cgroup.procs to apply limit.
			if err := os.WriteFile(filepath.Join(gockerCPUPath, "cgroup.procs"), []byte(pidStr), 0644); err != nil {
				_ = os.WriteFile(filepath.Join(gockerCPUPath, "tasks"), []byte(pidStr), 0644)
			}
		}

		// Configure Memory Limits (v1)
		if isV1Memory && memBytes != "" {
			gockerMemPath := filepath.Join(memPath, "gocker")
			if err := os.MkdirAll(gockerMemPath, 0755); err != nil {
				return fmt.Errorf("failed to create cgroup memory dir: %w", err)
			}
			if err := os.WriteFile(filepath.Join(gockerMemPath, "memory.limit_in_bytes"), []byte(memBytes), 0644); err != nil {
				return fmt.Errorf("failed to write memory.limit_in_bytes: %w", err)
			}
			if err := os.WriteFile(filepath.Join(gockerMemPath, "cgroup.procs"), []byte(pidStr), 0644); err != nil {
				_ = os.WriteFile(filepath.Join(gockerMemPath, "tasks"), []byte(pidStr), 0644)
			}
		}
		return nil
	}

	// Fallback to Cgroups v2 unified filesystem if cgroups v1 is not detected.
	v2Path := "/sys/fs/cgroup"
	if _, err := os.Stat(filepath.Join(v2Path, "cgroup.controllers")); err == nil {
		gockerV2Path := filepath.Join(v2Path, "gocker")
		if err := os.MkdirAll(gockerV2Path, 0755); err != nil {
			return fmt.Errorf("failed to create cgroup v2 dir: %w", err)
		}

		// Enable memory and cpu controllers in the subtree control of parent group.
		_ = os.WriteFile(filepath.Join(v2Path, "cgroup.subtree_control"), []byte("+cpu +memory"), 0644)

		if cpuShares > 0 {
			// Convert CPU shares to v2 weight. v1 shares (2-262144, default 1024) map to v2 weight (1-10000, default 100).
			weight := ((cpuShares - 2) * 9999 / 262142) + 1
			if weight < 1 {
				weight = 1
			}
			if weight > 10000 {
				weight = 10000
			}
			_ = os.WriteFile(filepath.Join(gockerV2Path, "cpu.weight"), []byte(strconv.Itoa(weight)), 0644)
		}

		if memBytes != "" {
			// Write memory max limit for Cgroups v2.
			_ = os.WriteFile(filepath.Join(gockerV2Path, "memory.max"), []byte(memBytes), 0644)
		}

		// Write process PID to cgroup.procs to enforce limits.
		if err := os.WriteFile(filepath.Join(gockerV2Path, "cgroup.procs"), []byte(pidStr), 0644); err != nil {
			return fmt.Errorf("failed to write PID to cgroup.procs in v2: %w", err)
		}
		return nil
	}

	return fmt.Errorf("no cgroups filesystem detected")
}

// parseMemoryLimit parses memory limit strings like "50M", "100M", "1G" into bytes string.
func parseMemoryLimit(memStr string) (string, error) {
	memStr = strings.TrimSpace(strings.ToUpper(memStr))
	if memStr == "" {
		return "", nil
	}
	var multiplier int64 = 1
	var numStr string

	if strings.HasSuffix(memStr, "K") {
		multiplier = 1024
		numStr = memStr[:len(memStr)-1]
	} else if strings.HasSuffix(memStr, "KB") {
		multiplier = 1024
		numStr = memStr[:len(memStr)-2]
	} else if strings.HasSuffix(memStr, "M") {
		multiplier = 1024 * 1024
		numStr = memStr[:len(memStr)-1]
	} else if strings.HasSuffix(memStr, "MB") {
		multiplier = 1024 * 1024
		numStr = memStr[:len(memStr)-2]
	} else if strings.HasSuffix(memStr, "G") {
		multiplier = 1024 * 1024 * 1024
		numStr = memStr[:len(memStr)-1]
	} else if strings.HasSuffix(memStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		numStr = memStr[:len(memStr)-2]
	} else {
		numStr = memStr
	}

	val, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid memory limit format: %s", memStr)
	}
	return strconv.FormatInt(val*multiplier, 10), nil
}
