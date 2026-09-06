$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

# 1. Detect Architecture
$arch = "amd64"
if ($PSVersionTable.OS -eq "Windows") {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { $arch = "arm64" }
}

$binary = "webtyp-installer-windows-$arch.exe"
$url = "https://github.com/webtyp/installer/releases/latest/download/$binary"

# 2. Download
Write-Host "Downloading $binary..."
Invoke-WebRequest -Uri $url -OutFile $binary

# 3. Execute
& .\"$binary" $args

# 4. Cleanup
Remove-Item "$binary"
