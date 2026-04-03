# Stage 7 — Verify (Go)

## Goal

After each tool installs, verify it works by running `tool version` or `tool --version`.

## Implementation

```go
func verify(name, expectedVersion string) error {
    // Try "name version" first, then "name --version"
    for _, arg := range []string{"version", "--version"} {
        out, err := exec.Command(name, arg).CombinedOutput()
        if err == nil && strings.Contains(string(out), expectedVersion) {
            return nil
        }
    }
    return fmt.Errorf("%s: version %s not found", name, expectedVersion)
}
```

## Acceptance

After install, spinner line updates to:
```
✅ tinygo — 0.40.1
```
