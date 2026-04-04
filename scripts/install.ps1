$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

$repoRaw = "https://raw.githubusercontent.com/tinywasm/installer/main"

# 1. Ask which optional tools to install — before any download begins
$toolsFlag = "-tools=all"
try {
    $savedPref = $ErrorActionPreference
    $ErrorActionPreference = "SilentlyContinue"
    $ans = Read-Host "Install optional tools (tinywasm-cli, tinywasm-server)? [Y/n]"
    $ErrorActionPreference = $savedPref
    if ($ans -eq 'n' -or $ans -eq 'N') { $toolsFlag = "" }
} catch {
    # Non-interactive session (e.g. SSH, iex): default to install all
}

# 2. Read Go version from go_version.conf
$goVersion = (Invoke-RestMethod "$repoRaw/go_version.conf").Trim()

# 3. Install Go
$env:GO_VERSION = $goVersion
irm https://raw.githubusercontent.com/tinywasm/goinstall/main/scripts/install.ps1 | iex

# 4. Refresh PATH so go.exe and installed binaries are available in this session
$goBin = "C:\Program Files\Go\bin"
$userGoBin = "$env:USERPROFILE\go\bin"
$machinePath = [System.Environment]::GetEnvironmentVariable("PATH", "Machine")
$env:PATH = "$goBin;$userGoBin;$machinePath"

# 5. Verify
go version

# 6. Install and run the Go orchestrator
go install github.com/tinywasm/installer/cmd/installer@v0.0.29
$installerBin = "$env:USERPROFILE\go\bin\installer.exe"
& $installerBin $toolsFlag
