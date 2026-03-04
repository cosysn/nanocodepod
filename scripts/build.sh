#!/bin/bash
# Build script for multi-platform release
# Builds CLI, agent, and server for all platforms

set -e

COMMIT_ID=${1:-$(git rev-parse --short HEAD)}
OUTPUT_DIR=${2:-./release}

echo "Building CodePod release: $COMMIT_ID"
echo "Output directory: $OUTPUT_DIR"

# Platform configurations
PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64")

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    OS=$(echo $PLATFORM | cut -d'/' -f1)
    ARCH=$(echo $PLATFORM | cut -d'/' -f2)

    # Determine platform directory name
    case "$OS-$ARCH" in
        linux-amd64) DIR="linux-x86" ;;
        linux-arm64) DIR="linux-arm" ;;
        darwin-amd64) DIR="macos-x86" ;;
        darwin-arm64) DIR="macos-arm" ;;
        *) DIR="$OS-$ARCH" ;;
    esac

    echo "Building for $DIR..."

    # Create platform directory
    PLATFORM_DIR="$OUTPUT_DIR/$DIR"
    mkdir -p "$PLATFORM_DIR"

    # Build CLI
    GOOS=$OS GOARCH=$ARCH go build -o "$PLATFORM_DIR/codepod" ./codepod

    # Build agent
    GOOS=$OS GOARCH=$ARCH go build -o "$PLATFORM_DIR/codepod-agent" ./codepod/cmd/agent

    # Build server
    GOOS=$OS GOARCH=$ARCH go build -o "$PLATFORM_DIR/codepod-server" ./server/cmd/wsl-server
done

# Create archive
ARCHIVE_NAME="codepod-${COMMIT_ID}.tar.gz"
echo "Creating archive: $ARCHIVE_NAME"

cd "$OUTPUT_DIR"
tar -czvf "../$ARCHIVE_NAME" */

echo "Build complete: $ARCHIVE_NAME"
ls -la "../$ARCHIVE_NAME"
