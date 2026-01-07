# Stage 1: Build the Go binary
FROM golang:1.22-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum* ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gocker-runtime main.go

# Stage 2: Final Image
FROM debian:bookworm-slim
WORKDIR /root/

# 1. Install system tools (for the HOST view)
RUN apt-get update && apt-get install -y \
    procps \
    iproute2 \
    && rm -rf /var/lib/apt/lists/*

# 2. Copy the Gocker binary from builder
COPY --from=builder /app/gocker-runtime .
RUN chmod +x ./gocker-runtime

# 3. Copy the isolated RootFS folder for the GUEST view
COPY rootfs ./rootfs

# 4. Start the runtime
ENTRYPOINT ["./gocker-runtime"]