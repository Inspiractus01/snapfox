#!/usr/bin/env bash
set -e

echo "Building Snapfox for all platforms..."

mkdir -p builds

echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -o builds/snapfox-darwin-arm64 .

echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -o builds/snapfox-darwin-amd64 .

echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o builds/snapfox-linux-amd64 .

echo "Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build -o builds/snapfox-linux-arm64 .

echo "All builds complete:"
ls -lh builds/
