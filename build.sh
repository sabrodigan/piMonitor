#!/bin/bash
echo "Building piMonitor for Raspberry Pi 5 (ARM64)..."
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.build=$(git rev-parse --short HEAD)" -o piMonitor .

echo "Build complete. Executable is named 'piMonitor'."
echo "Transfer this file and the 'templates' folder to your Pi."
