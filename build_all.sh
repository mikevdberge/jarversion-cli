#!/bin/bash

# Check if version string is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <version>"
  exit 1
fi

VERSION=$1
COMMIT_HASH=$(git rev-parse --short HEAD)
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "Building jarversion CLI tool with version: $VERSION"
echo "Commit hash: $COMMIT_HASH"
echo "Build time: $BUILD_TIME"

LDFLAGS="-X 'main.toolVersion=$VERSION' -X 'main.commitHash=$COMMIT_HASH' -X 'main.buildTime=$BUILD_TIME'"

# Build for Linux (amd64)
GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o ./build/package/jarversion-linux ./cmd/jarversion/main.go

# Build for Windows (amd64)
GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o ./build/package/jarversion-windows.exe ./cmd/jarversion/main.go

# Build for macOS (arm64)
GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o ./build/package/jarversion-macos-arm64 ./cmd/jarversion/main.go

echo "✅ Build completed for all targets.🚀"