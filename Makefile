.PHONY: build test-run clean

build:
	CGO_ENABLED=0 go build -o gocker main.go

test-run: build
	mkdir -p rootfs
	@if [ ! -f alpine.tar.gz ]; then \
		echo "Downloading Alpine minirootfs..."; \
		curl -sSL -o alpine.tar.gz https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-minirootfs-3.19.1-x86_64.tar.gz || \
		wget -O alpine.tar.gz https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-minirootfs-3.19.1-x86_64.tar.gz; \
	fi
	@if [ ! -f rootfs/bin/sh ]; then \
		echo "Unpacking rootfs..."; \
		tar -xzf alpine.tar.gz -C rootfs; \
	fi
	sudo ./gocker run --cpu 512 --mem 50M --rootfs ./rootfs /bin/sh

clean:
	rm -rf gocker alpine.tar.gz rootfs
