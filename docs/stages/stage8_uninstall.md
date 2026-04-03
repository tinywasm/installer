# Stage 8 — Selective uninstall (Go)

## Goal

Remove tools via `UNINSTALL` env var.

## Usage

```bash
UNINSTALL=tinygo installer
UNINSTALL=all installer
```

## Uninstall per mode

| Mode | Action |
|------|--------|
| GoInstall | `rm $(exec.LookPath(name))` |
| Binary | `rm $(exec.LookPath(name))` |

## Steps

- [ ] Read `UNINSTALL` env var
- [ ] If `all`: iterate tools in reverse, remove each
- [ ] If specific name: find and remove that tool only
- [ ] Verify: `exec.LookPath(name)` should fail after removal
- [ ] Print: `✅ tinygo — removed`

## Acceptance

```
$ UNINSTALL=tinygo installer
✅ tinygo — removed
```
