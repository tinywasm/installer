#!/bin/bash
set -e

REPO_RAW="https://raw.githubusercontent.com/tinywasm/installer/main"

# 1. Read Go version from go_version.conf (single source of truth)
GO_VERSION=$(curl -fsSL "$REPO_RAW/go_version.conf" | tr -d '[:space:]')

# 2. Install Go using goinstall script (script handles sudo internally)
curl -fsSL https://raw.githubusercontent.com/tinywasm/goinstall/main/scripts/install.sh | bash -s "$GO_VERSION"
hash -r

# 3. Verify Go is available
if ! command -v go &>/dev/null; then
    echo "Error: Go installation failed"
    exit 1
fi

# 4. Install and run the Go orchestrator
go install github.com/tinywasm/installer/cmd/installer@latest
"$(go env GOPATH)/bin/installer" "$@"
