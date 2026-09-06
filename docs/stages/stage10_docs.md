# Stage 10 — Documentation (README.md)

## Goal

README.md del repo con quick start, requisitos, lista de herramientas, uninstall, y cómo agregar nuevas tools.

## README.md

```markdown
# webtyp/installer

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

1. Installs Go (version pinned in `go_version.conf`)
2. Builds and runs the Go installer binary
3. Shows an interactive checklist of optional tools
4. Installs all required + selected tools with progress spinner
5. Verifies each tool after installation

## Tools

| Tool | Mode | Required | Depends on |
|------|------|----------|------------|
| tinygoinstall | GoInstall | yes | — |
| git | Binary | no | — |
| gh | Binary | no | git |

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
    {Mode: Binary, Name: "mytool", Source: "github.com/org/mytool", Version: "1.0.0", Required: false},
}
```

If the tool depends on another:

```go
{Mode: Binary, Name: "gh", Source: "...", Version: "2.65.0", Required: false, DependsOn: "git"},
```

## How it works

```
Phase 1: bash / ps1
├── Read Go version from go_version.conf
├── Install Go via goinstall script
└── go install webtyp/installer → run installer

Phase 2: Go binary
├── Show checklist of optional tools
├── Install required + selected tools
├── Verify each tool
└── Print summary
```
```

## Acceptance

README.md exists in repo root with all sections above. Content matches
current tool registry and install modes.
