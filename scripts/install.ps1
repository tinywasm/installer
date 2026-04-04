$ErrorActionPreference = "Stop"

$repoRaw = "https://raw.githubusercontent.com/tinywasm/installer/main"

# 1. Read Go version from go_version.conf
$goVersion = (Invoke-RestMethod "$repoRaw/go_version.conf").Trim()

# 2. Install Go
$env:GO_VERSION = $goVersion
irm https://raw.githubusercontent.com/tinywasm/goinstall/main/scripts/install.ps1 | iex

# 3. Refresh PATH so go.exe is available in this session
$env:PATH = "C:\Program Files\Go\bin;" + [System.Environment]::GetEnvironmentVariable("PATH", "Machine")

# 4. Verify
go version

# 5. Install and run the Go orchestrator
go install github.com/tinywasm/installer/cmd/installer@latest
installer @args
