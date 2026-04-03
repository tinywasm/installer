# Stage 3 — Checklist (interactive selection of optional tools)

## Goal

Before installation starts, show an interactive checklist of optional tools.
The user selects which ones to install using keyboard navigation.
After confirmation, the installation runs unattended (no more prompts).

Required tools are NOT shown in the checklist — they always install.

## Dependencies

- `golang.org/x/term` — raw terminal mode for reading keypresses without Enter

## checklist.go

```go
// ShowChecklist displays optional tools and returns the selected indices.
// Uses raw terminal mode to capture ↑/↓/Space/Enter without buffering.
func ShowChecklist(tools []Tool) []int {
    // 1. Filter optional tools (Required == false)
    // 2. Enter raw terminal mode (term.MakeRaw)
    // 3. Render checklist with cursor
    // 4. Read keypresses:
    //    - ↑/↓ (escape sequences \x1b[A / \x1b[B): move cursor
    //    - Space: toggle [x] / [ ]
    //    - Enter: confirm and return selected indices
    // 5. Restore terminal mode (defer)
}
```

## Terminal output

```
Select optional tools to install:
  [x] git — 2.47.0
  [x]   └─ gh — 2.65.0  (requires git)
  [ ] lazygit — 0.44.1
  [x] wasmtime — 25.0.0

  ↑/↓: move  Space: toggle  Enter: confirm
```

## Dependency rendering

Tools with `DependsOn` are indented under their dependency with `└─` prefix.
When a dependency is unchecked, its dependents are auto-unchecked and disabled:

```
  [ ] git — 2.47.0
  [-]   └─ gh — 2.65.0  (requires git)    ← disabled, cannot toggle
```

When the dependency is re-checked, dependents become toggleable again (but stay unchecked — user must opt in).

## Rendering

Each re-render:
1. Move cursor up N lines (`\033[{N}A`)
2. Clear each line (`\033[2K`)
3. Print updated checklist
4. Highlight current row with `>` prefix
5. Disabled tools shown with `[-]` and dimmed text (ANSI `\033[2m`)

```
Select optional tools to install:
> [x] git — 2.47.0            ← cursor here
  [x]   └─ gh — 2.65.0  (requires git)
  [ ] lazygit — 0.44.1
  [x] wasmtime — 25.0.0

  ↑/↓: move  Space: toggle  Enter: confirm
```

## Key mapping

| Key | Escape sequence | Action |
|-----|----------------|--------|
| `↑` | `\x1b[A` | Move cursor up |
| `↓` | `\x1b[B` | Move cursor down |
| `Space` | `0x20` | Toggle selected/unselected |
| `Enter` | `0x0D` | Confirm selection |
| `q` / `Ctrl+C` | `0x71` / `0x03` | Cancel (install required only) |

## Integration with main flow

```go
func main() {
    // ...
    if uninstall {
        // uninstall flow
        return
    }

    // Show checklist for optional tools
    selected := ShowChecklist(tools)

    // Build install list: all required + selected optional
    var toInstall []Tool
    for _, t := range tools {
        if t.Required {
            toInstall = append(toInstall, t)
        }
    }
    for _, idx := range selected {
        toInstall = append(toInstall, optionalTools[idx])
    }

    // Install all — no more prompts
    installAll(toInstall)
}
```

## Edge cases

- **No optional tools**: skip checklist, install required tools directly.
- **Non-interactive terminal** (piped input): install required tools only, skip optional.
- **All optional unchecked**: install required only, mark optional as skipped.

## Acceptance

Running the installer shows the checklist, responds to keyboard input,
and proceeds to install only the selected tools without further prompts.
