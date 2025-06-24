#!/bin/bash

# make sure buildx is available
echo "Make sure Docker buildx is available..."
docker buildx version || { echo "Please install Docker buildx"; exit 1; }
docker buildx create --name wmiexec-multiarch --platform linux/amd64,linux/arm64,linux/arm/v7 --use --bootstrap || true
docker buildx use wmiexec-multiarch

mkdir -p binaries

# AMD64
echo "Start to build for AMD64"
docker buildx build --platform linux/amd64 --tag wmiexec-builder-amd64 --load .

echo "Creating container for AMD64 and extracting binary..."
docker create --name wmiexec-amd64 wmiexec-builder-amd64
docker cp wmiexec-amd64:/app/dist/wmiexec-static ./binaries/wmiexec-amd64
docker rm wmiexec-amd64
chmod +x ./binaries/wmiexec-amd64

# ARM64
echo "Start to build for ARM64"
docker buildx build --platform linux/arm64 --tag wmiexec-builder-arm64 --load .

echo "Creating container for ARM64 and extracting binary..."
docker create --name wmiexec-arm64 wmiexec-builder-arm64
docker cp wmiexec-arm64:/app/dist/wmiexec-static ./binaries/wmiexec-arm64
docker rm wmiexec-arm64
chmod +x ./binaries/wmiexec-arm64

# Clean up buildx builder
docker buildx rm wmiexec-multiarch

echo "Successfully built！Generated the following files:"
echo "- binaries/wmiexec-amd64"
echo "- binaries/wmiexec-arm64"
