# Stage 1 — Scripts (bash + PowerShell)

## Goal

Minimal scripts that ONLY install Go and then run the Go orchestrator.
Go version is read from `go_version.conf` (single source of truth).

## install.sh (Linux + macOS)

```bash
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
installer "$@"
```

## install.ps1 (Windows)

```powershell
$ErrorActionPreference = "Stop"

$repoRaw = "https://raw.githubusercontent.com/tinywasm/installer/main"

# 1. Read Go version from go_version.conf
$goVersion = (Invoke-RestMethod "$repoRaw/go_version.conf").Trim()

# 2. Install Go
$env:GO_VERSION = $goVersion
irm https://raw.githubusercontent.com/tinywasm/goinstall/main/scripts/install.ps1 | iex

# 3. Verify
& "C:\Program Files\Go\bin\go.exe" version

# 4. Install and run the Go orchestrator
go install github.com/tinywasm/installer/cmd/installer@latest
installer @args
```

## Acceptance

Scripts install Go and hand off to the Go program. Nothing else.
