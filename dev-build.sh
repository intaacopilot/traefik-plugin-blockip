#!/bin/bash

# Development build script
set -e

echo "Building Traefik BlockIP Plugin (Development)..."

# Download dependencies
echo "Downloading dependencies..."
go mod download
go mod tidy

# Run tests
echo "Running tests..."
go test -v -cover ./...

# Build
echo "Building..."
go build -v -o traefik-plugin-blockip . 

echo "✓ Build successful!"
echo "Binary: traefik-plugin-blockip"
ls -lh traefik-plugin-blockip