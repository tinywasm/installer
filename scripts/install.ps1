$ErrorActionPreference = "Stop"

$repoRaw = "https://raw.githubusercontent.com/tinywasm/installer/main"

# 1. Read Go version from go_version.conf
$goVersion = (Invoke-RestMethod "$repoRaw/go_version.conf").Trim()

# 2. Install Go
$env:GO_VERSION = $goVersion
irm https://raw.githubusercontent.com/tinywasm/goinstall/main/scripts/install.ps1 | iex

# 3. Refresh PATH so go.exe and installed binaries are available in this session
$goBin = "C:\Program Files\Go\bin"
$userGoBin = "$env:USERPROFILE\go\bin"
$machinePath = [System.Environment]::GetEnvironmentVariable("PATH", "Machine")
$env:PATH = "$goBin;$userGoBin;$machinePath"

# 4. Verify
go version

# 5. Install and run the Go orchestrator
go install github.com/tinywasm/installer/cmd/installer@latest
$installerBin = "$env:USERPROFILE\go\bin\installer.exe"
& $installerBin @args
