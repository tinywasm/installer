# Stage 5 — go install mode (Go)

## Goal

Implement `GoInstall` mode: `exec.Command("go", "install", source+"@"+version)`.

## Steps

- [ ] Run `go install <source>@<version>`
- [ ] Capture stdout/stderr
- [ ] Return error if exit code != 0

## Acceptance

```
✅ tinygoinstall — 0.40.1
```
