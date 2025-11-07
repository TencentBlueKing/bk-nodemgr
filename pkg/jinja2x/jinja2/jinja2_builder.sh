#!/bin/bash

# make sure buildx is available
echo "Make sure Docker buildx is available..."
docker buildx version || { echo "Please install Docker buildx"; exit 1; }
docker buildx create --name jinja2_exec-multiarch --platform linux/amd64,linux/arm64,linux/arm/v7 --use --bootstrap || true
docker buildx use jinja2_exec-multiarch

mkdir -p binaries

# AMD64
echo "Start to build for AMD64"
docker buildx build --platform linux/amd64 --tag jinja2_exec-builder-amd64 --load .

echo "Creating container for AMD64 and extracting binary..."
docker create --name jinja2_exec-amd64 jinja2_exec-builder-amd64
docker cp jinja2_exec-amd64:/app/dist/jinja2_exec-static ./binaries/jinja2_exec-amd64
docker rm jinja2_exec-amd64
chmod +x ./binaries/jinja2_exec-amd64

# Clean up buildx builder
docker buildx rm jinja2_exec-multiarch

echo "Successfully built！Generated the following files:"
echo "- binaries/jinja2_exec-amd64"