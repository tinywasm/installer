#!/bin/bash
set -e

# 1. Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case $OS in
    linux) ;;
    darwin) ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# 2. Download the installer binary from GitHub releases
BINARY="webtyp-installer-$OS-$ARCH"
URL="https://github.com/webtyp/installer/releases/latest/download/$BINARY"

echo "Downloading $BINARY..."
curl -fsSL "$URL" -o "$BINARY"
chmod +x "$BINARY"

# 3. Execute the installer
./"$BINARY" "$@"

# 4. Cleanup
rm "$BINARY"
