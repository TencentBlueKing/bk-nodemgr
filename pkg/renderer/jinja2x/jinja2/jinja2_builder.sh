#!/bin/bash

# make sure buildx is available
echo "Make sure Docker buildx is available..."
docker buildx version || { echo "Please install Docker buildx"; exit 1; }
docker buildx create --name jinja2_exec_multiarch --platform linux/amd64,linux/arm64,linux/arm/v7 --use --bootstrap || true
docker buildx use jinja2_exec_multiarch

mkdir -p binaries

# AMD64
echo "Start to build for AMD64"
docker buildx build --platform linux/amd64 --tag jinja2_exec_builder_amd64 --load .

echo "Creating container for AMD64 and extracting binary..."
docker create --name jinja2_exec_amd64 jinja2_exec_builder_amd64
docker cp jinja2_exec_amd64:/app/dist/jinja2_exec_static ./binaries/jinja2_exec_amd64
docker rm jinja2_exec_amd64
chmod +x ./binaries/jinja2_exec_amd64

# ARM64
echo "Start to build for ARM64"
docker buildx build --platform linux/arm64 --tag jinja2_exec_builder_arm64 --load .

echo "Creating container for ARM64 and extracting binary..."
docker create --name jinja2_exec_arm64 jinja2_exec_builder_arm64
docker cp jinja2_exec_arm64:/app/dist/jinja2_exec_static ./binaries/jinja2_exec_arm64
docker rm jinja2_exec_arm64
chmod +x ./binaries/jinja2_exec_arm64

# Clean up buildx builder
docker buildx rm jinja2_exec_multiarch

echo "Successfully built！Generated the following files:"
echo "- binaries/jinja2_exec_amd64"
echo "- binaries/jinja2_exec_arm64"