# webtyp/installer
<img src="docs/img/badges.svg">

Single command to set up a complete webtyp development environment.

## Quick start

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/webtyp/installer/main/scripts/install.sh | bash
```

### Windows (PowerShell as Admin)

```powershell
irm https://raw.githubusercontent.com/webtyp/installer/main/scripts/install.ps1 | iex
```

## What it does

1. Downloads the latest `webtyp-installer` binary for your platform
2. Shows an interactive checklist of optional tools
3. Installs all required + selected tools (including Go/TinyGo) with progress spinner
4. Verifies each tool after installation

## Tools

| Tool | Mode | Required | Depends on |
|------|------|----------|------------|
| tinygoinstall | GoInstall | yes | — |
| webtyp | Binary | yes | — |

Required tools install automatically. Optional tools are shown in an
interactive checklist before installation begins.

## Uninstall

```bash
# Remove a specific tool
UNINSTALL=tinygo curl -fsSL .../install.sh | bash

# Remove all tools
UNINSTALL=all curl -fsSL .../install.sh | bash
```

## Requirements

- **Linux / macOS**: bash, curl
- **Windows**: PowerShell 5.1+
- Internet connection

Go is installed automatically — no need to have it pre-installed.

## Adding a new tool

Add one entry to the `tools` slice in `installer.go`:

```go
var tools = []Tool{
    // existing tools...
    {Mode: Binary, Name: "mytool", Source: "https://github.com/org/mytool", Version: "1.0.0", Required: false},
}
```

If the tool depends on another:

```go
{Mode: Binary, Name: "gh", Source: "https://github.com/cli/cli", Version: "2.65.0", Required: false, DependsOn: "git"},
```

## How it works

Version resolution, checksum verification, and network downloads are powered by [webtyp/update](https://github.com/webtyp/update).

```
Phase 1: bash / ps1
├── Detect OS/Arch
└── Download webtyp-installer-{os}-{arch} → run installer

Phase 2: Go binary (webtyp-installer)
├── Show checklist of optional tools
├── Install required + selected tools
├── Verify each tool
└── Print summary
```


## Testing

- [Windows integration test setup](docs/TEST_WINDOWS.md)
