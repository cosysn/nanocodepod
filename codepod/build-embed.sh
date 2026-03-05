#!/bin/bash
# Build script for CodePod with embedded agents
# Builds agent binaries for all platforms and embeds them into the codepod binary

set -e

OUTPUT_DIR="dist"
EMBED_DIR="pkg/embed"
TEMP_DIR="/tmp/codepod-embed-build"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== CodePod Embedded Agent Build ===${NC}"

# Check Go version
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

GO_VERSION=$(go version)
echo -e "Found: ${GREEN}$GO_VERSION${NC}"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Create temp directory
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Define platforms to build
PLATFORMS=(
    "linux,amd64,agent-linux-amd64"
    "linux,arm64,agent-linux-arm64"
    "darwin,amd64,agent-darwin-amd64"
    "darwin,arm64,agent-darwin-arm64"
    "windows,amd64,agent-windows-amd64.exe"
)

echo -e "\n${YELLOW}Building agent binaries...${NC}"

# Build agent for each platform
for platform in "${PLATFORMS[@]}"; do
    IFS=',' read -r GOOS GOARCH OUTPUT_NAME <<< "$platform"

    echo -n "  Building $OUTPUT_NAME ($GOOS/$GOARCH)... "

    # Set environment
    export GOOS="$GOOS"
    export GOARCH="$GOARCH"

    # Build agent
    go build -ldflags "-s -w" -o "$TEMP_DIR/$OUTPUT_NAME" ./cmd/agent

    echo -e "${GREEN}✓${NC}"
done

# Reset Go environment
unset GOOS
unset GOARCH

# Create agents.zip
echo -e "\n${YELLOW}Creating agents.zip...${NC}"
cd "$TEMP_DIR"
zip -0 "$SCRIPT_DIR/$EMBED_DIR/agents.zip" agent-linux-amd64 agent-linux-arm64 agent-darwin-amd64 agent-darwin-arm64 agent-windows-amd64.exe
cd "$SCRIPT_DIR"

# Show zip contents
echo -e "  Contents:"
unzip -l "$EMBED_DIR/agents.zip" | grep -v "Archive:" | grep -v "Length" | grep -v "^$" | while read line; do
    echo "    $line"
done

# Clean up temp files
rm -rf "$TEMP_DIR"

echo -e "\n${GREEN}Embedded agents created successfully!${NC}"

# Build final codepod binary with embedded agents
echo -e "\n${YELLOW}Building codepod binary with embedded agents...${NC}"

# Build for current platform
go build -ldflags "-s -w" -o "$OUTPUT_DIR/codepod" .

echo -e "${GREEN}✓${NC}"

echo -e "\n${GREEN}=== Build Complete ===${NC}"
echo "Output: $OUTPUT_DIR/codepod"
