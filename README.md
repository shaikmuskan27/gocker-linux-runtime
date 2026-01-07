Gocker: A Custom Linux Container Runtime in Go
A low-level container engine implementation built from scratch to explore Linux Kernel primitives. This project demonstrates how tools like Docker use the Linux API to provide process isolation and resource management.

🏗️ Architecture
The runtime follows the Parent-Child Re-exec pattern:

Parent: Invokes /proc/self/exe with Cloneflags to create a new namespace "bubble."

Child: Configures the internal environment (hostname, chroot, mounts, cgroups).

Exec: Replaces the child process with the target binary (e.g., /bin/sh).

🚀 Key Features
Namespaces (Isolation): Implements UTS, PID, and Mount namespaces using syscall to isolate the container identity and process tree.

Filesystem Jailing: Uses chroot to lock the process into an Alpine Linux rootfs, ensuring the host filesystem remains invisible.

Resource Control (Cgroups): Implements Control Groups (v2) to enforce a hard limit on the number of processes (PIDs) to prevent fork-bomb attacks.

VFS Management: Manually mounts a private /proc to ensure the container only sees its own processes.

🛠️ Execution & Workflow
1. Preparing the Environment
We use a multi-stage Docker build managed by Terraform to deploy the runtime into a "Fake Azure" local environment.

[INSERT SCREENSHOT: Your VS Code terminal showing terraform apply success]

2. Building the RootFS
To provide the container with its own world, we extract a mini-rootfs from Alpine Linux:

Bash

mkdir rootfs
docker export $(docker create alpine) | tar -C rootfs -xf -
3. Launching the Runtime
The runtime is executed inside a privileged container to allow the Go binary to manipulate host namespaces.

Bash

docker run --privileged -it gocker-final run /bin/sh
[INSERT SCREENSHOT: Your VS Code terminal showing the "Running [/bin/sh] as 1" output]

🧪 Verification & Results
Once inside the Gocker shell, we can verify the isolation:

PID Isolation
Running ps confirms that the shell is running as PID 1, and the container cannot see host processes.

[INSERT SCREENSHOT: Terminal showing ps command with PID 1]

Filesystem Isolation
Running ls / shows the Alpine filesystem, and attempts to access the host's gocker-runtime binary fail, proving the chroot jail is active.

[INSERT SCREENSHOT: Terminal showing ls / and the failed attempt to see host files]

🔧 Technical Stack
Language: Go (Golang)

Infrastructure: Terraform

Containerization: Docker (Multi-stage builds)

OS Interface: Linux Syscalls (CLONE_NEWPID, CLONE_NEWUTS, CLONE_NEWNS)