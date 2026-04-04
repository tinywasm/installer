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
