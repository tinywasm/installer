# Stage 4 — Spinner (Go)

## Goal

Spinner implemented in Go using goroutines.

## Implementation

```go
// spinner.go
func runWithSpinner(name string, fn func() error) error {
    done := make(chan error, 1)
    go func() { done <- fn() }()

    chars := []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")
    i := 0
    for {
        select {
        case err := <-done:
            if err != nil {
                fmt.Printf("\r❌ %s — %v\n", name, err)
            } else {
                fmt.Printf("\r✅ %s\n", name)
            }
            return err
        default:
            fmt.Printf("\r%c Installing %s...", chars[i%len(chars)], name)
            i++
            time.Sleep(100 * time.Millisecond)
        }
    }
}
```

## Acceptance

```
⠹ Installing tinygo...
✅ tinygo — 0.40.1
```
