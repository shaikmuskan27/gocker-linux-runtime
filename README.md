# Gocker: A Custom Linux Container Runtime in Go

A low-level container engine implementation built from scratch to explore Linux Kernel primitives. This project demonstrates how tools like Docker use the Linux API to provide process isolation and resource management.

## 🚀 Key Features
- **Namespaces (Isolation):** Implements UTS, PID, and Mount namespaces using `syscall` to isolate the container identity and process tree.
- **Filesystem Jailing:** Uses `chroot` and `pivot_root` concepts to isolate the container's root filesystem using an Alpine Linux rootfs.
- **Resource Control (Cgroups):** Implements Control Groups (v2) to enforce a hard limit on the number of processes (PIDs) to prevent fork-bomb attacks.
- **VFS Management:** Manually mounts a private `/proc` to ensure the container only sees its own processes.

## 🏗️ Architecture
The runtime follows the **Parent-Child Re-exec** pattern:
1. **Parent:** Invokes `/proc/self/exe` with `Cloneflags` to create a new namespace "bubble."
2. **Child:** Configures the internal environment (hostname, chroot, mounts, cgroups).
3. **Exec:** Replaces the child process with the target binary (e.g., `/bin/sh`).



## 🛠️ Usage
1. **Download the rootfs:**
   ```bash
   mkdir rootfs
   curl -Lo alpine.tar.gz [https://dl-cdn.alpinelinux.org/alpine/v3.18/releases/x86_64/alpine-minirootfs-3.18.4-x86_64.tar.gz](https://dl-cdn.alpinelinux.org/alpine/v3.18/releases/x86_64/alpine-minirootfs-3.18.4-x86_64.tar.gz)
   tar -xf alpine.tar.gz -C rootfs