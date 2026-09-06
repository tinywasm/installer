# Stage 2 — Go skeleton

## Goal

Go program with tool registry (versions inline), `go_version.conf` embed, and CLI entry point.
No install logic yet — just the structure.

## Files

### installer.go — Tool registry + types

```go
type Mode int

const (
    Script Mode = iota    // download + exec external script
    GoInstall             // go install source@version
    Binary                // download GitHub release
)

type Tool struct {
    Mode      Mode
    Name      string
    Source    string  // module path (GoInstall) or GitHub repo (Binary)
    Version   string  // pinned in tool definition
    Required  bool    // true = always install; false = prompt user
    DependsOn string  // "" = no dependency; "git" = requires git installed first
}

var tools = []Tool{
    {Mode: GoInstall, Name: "tinygoinstall", Source: "webtyp.com/tinygo/cmd/tinygoinstall", Required: true},
    {Mode: Binary,    Name: "git",           Source: "...", Required: false},
    {Mode: Binary,    Name: "gh",            Source: "...", Required: false, DependsOn: "git"},
}
```

### cmd/installer/main.go — CLI

```go
func main() {
    // 1. Tools already defined with versions in code
    // 2. Check UNINSTALL env
    // 3. Install tools with spinner
    // 4. Print summary
}
```

## Acceptance

Running `go run ./cmd/installer` prints the registry with versions loaded:
```
go — 1.25.2 (script)
tinygoinstall — 0.40.1 (go)
```
