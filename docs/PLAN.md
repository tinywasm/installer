# PLAN: installer — tinywasm ecosystem orchestrator

## Goal

Single command to set up a complete tinywasm development environment.
Bash/PowerShell scripts **only** install Go. After that, a Go program
handles everything else: spinner, tool installation, verification, uninstall.

```bash
# Linux / macOS
curl -fsSL https://raw.githubusercontent.com/tinywasm/installer/main/scripts/install.sh | bash

# Windows (PowerShell as Admin)
irm https://raw.githubusercontent.com/tinywasm/installer/main/scripts/install.ps1 | iex

# Uninstall specific tool
UNINSTALL=tinygo curl -fsSL .../install.sh | bash
```

## Architecture — Two phases

```
Phase 1: bash / ps1
├── Read Go version from go_version.conf (curl from GitHub)
├── Install Go via goinstall script
└── go install tinywasm/installer → run installer

Phase 2: Go binary (installer)
├── Load embedded go_version.conf
├── Show checklist of optional tools (↑/↓ Space Enter)
│   └── Dependencies shown indented, auto-disabled if parent unchecked
├── Install required + selected (goroutines + spinner)
│   ├── GoInstall mode → go install source@version
│   ├── Binary mode   → download GitHub release
│   └── DependsOn check → skip if dependency failed/skipped
├── Verify each: tool version
└── Summary: ✅ / ❌
```

Go is **never** in the tool registry — it's handled entirely by Phase 1 (bash/ps1).

## Design Decisions

- **Scripts are minimal**: bash/ps1 install Go + `go install` + run — nothing else.
- **All logic in Go**: spinner, tool registry, install modes, verification, uninstall — testable, cross-platform.
- **Orchestrator only**: each tool's repo handles its own install. This repo calls them.
- **2 install modes**: `GoInstall` (`go install`) and `Binary` (GitHub release). More modes added when needed.
- **Checklist selection**: optional tools shown in interactive checklist (↑/↓ Space Enter) before install starts. Installation is fully unattended after confirmation.
- **Goroutines + spinner**: independent tools install concurrently. Tools with `DependsOn` wait for their dependency to finish before starting.
- **Selective uninstall**: `UNINSTALL=<tool>` or `UNINSTALL=all`.
- **Version pinning**: `go_version.conf` — single source of truth. Embedded in Go binary via `//go:embed`, read by scripts via raw GitHub URL.
- **`go install` not `go run`**: binary is cached locally, faster on subsequent runs.
- **Tool dependencies**: `DependsOn` field links a tool to its prerequisite. If the dependency fails or is skipped, the dependent tool is auto-skipped.

## Release naming convention (for `Binary` mode tools)

```
{tool}_v{version}_{os}_{arch}.tar.gz    # Linux / macOS
{tool}_v{version}_{os}_{arch}.zip       # Windows
```

OS: `linux`, `darwin`, `windows`
Arch: `amd64`, `arm64`

## Tool registration

### go_version.conf — Go version only

```
1.25.2
```

Used by bash/ps1 scripts (Phase 1) and embedded in the Go binary via `//go:embed`.

### Go code — tool registry with versions

```go
var tools = []Tool{
    {Mode: GoInstall, Name: "tinygoinstall", Source: "github.com/tinywasm/tinygo/cmd/tinygoinstall", Version: "0.40.1", Required: true},
    {Mode: Binary,    Name: "git",           Source: "...", Version: "2.47.0", Required: false},
    {Mode: Binary,    Name: "gh",            Source: "...", Version: "2.65.0", Required: false, DependsOn: "git"},
}
```

**Adding a new tool = 1 entry in `tools` slice**

### Required vs Optional

```go
var tools = []Tool{
    {Mode: GoInstall, Name: "tinygoinstall", Source: "...", Required: true},
    {Mode: Binary,    Name: "git",           Source: "...", Required: false},
    {Mode: Binary,    Name: "gh",            Source: "...", Required: false, DependsOn: "git"},
}
```

- **Required** (`Required: true`): se instala siempre, no aparece en checklist. Si falla → error fatal, se detiene.
- **Optional** (`Required: false`): aparece en checklist interactivo al inicio. El usuario selecciona antes de que empiece la instalación. Si falla → warning, continua.

Terminal:
```
Select optional tools to install:
  [x] git — 2.47.0
  [x]   └─ gh — 2.65.0  (requires git)
  [ ] lazygit — 0.44.1
  [x] wasmtime — 25.0.0

  ↑/↓: move  Space: toggle  Enter: confirm

⠋ Installing tinygo...
✅ tinygo — 0.40.1
⠋ Installing git...
✅ git — 2.47.0
⠋ Installing gh...
✅ gh — 2.65.0
⠋ Installing wasmtime...
✅ wasmtime — 25.0.0
⏭ lazygit — skipped

Done. 4 installed, 0 failed, 1 skipped.
```

### Modes

| Mode | How it installs | When to use |
|------|----------------|-------------|
| `GoInstall` | `go install source@version` | Simple Go CLI tool |
| `Binary` | Downloads `{name}_v{version}_{os}_{arch}.tar.gz` from GitHub releases | Tool with many deps, pre-built binary is faster |

## Files

| File | Role |
|------|------|
| `scripts/install.sh` | Bash: install Go + `go install installer` + run |
| `scripts/install.ps1` | PowerShell: same for Windows |
| `go_version.conf` | Go version only (embedded + read from GitHub by scripts) |
| `installer.go` | Tool registry, config, Mode types, `//go:embed` |
| `checklist.go` | Interactive checklist for optional tools (`golang.org/x/term`) |
| `spinner.go` | Terminal spinner (goroutine) |
| `mode_go.go` | Mode: `go install source@version` |
| `mode_binary.go` | Mode: download GitHub release |
| `verify.go` | Run `tool version` and check output |
| `uninstall.go` | Selective uninstall per mode |
| `deps.go` | Dependency injection struct (`Deps`) for testability |
| `cmd/installer/main.go` | CLI entry point |

## Stages

| Stage | Description | Dependency | Completed |
|-------|-------------|------------|-----------|
| 1 | [Scripts](stages/stage1_skeleton.md) — bash/ps1: install Go + `go install installer` + run | — | [ ] |
| 2 | [Go skeleton](stages/stage2_go_skeleton.md) — Tool registry, config, CLI, go_version.conf embed | Stage 1 | [ ] |
| 3 | [Checklist](stages/stage3_checklist.md) — interactive selection of optional tools (`golang.org/x/term`) | Stage 2 | [ ] |
| 4 | [Spinner](stages/stage4_spinner.md) — goroutine spinner `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏` | Stage 2 | [ ] |
| 5 | [GoInstall mode](stages/stage5_mode_go.md) — `go install source@version` | Stage 2 | [ ] |
| 6 | [Binary mode](stages/stage6_mode_binary.md) — download release from GitHub | Stage 2 | [ ] |
| 7 | [Verify](stages/stage7_verify.md) — `tool version` check after install | Stage 2 | [ ] |
| 8 | [Uninstall](stages/stage8_uninstall.md) — selective: `UNINSTALL=<tool>` or `all` | Stage 5 | [ ] |
| 9 | [Tests](stages/stage9_tests.md) — dependency injection + mocks covering all diagram paths | Stage 2 | [ ] |
| 10 | [Docs](stages/stage10_docs.md) — README.md: quick start, tools, uninstall, adding tools | Stage 9 | [ ] |

## Flow Diagram

See [docs/diagrams/install_flow.md](diagrams/install_flow.md)
