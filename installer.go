package installer

import (
	_ "embed"
	"fmt"
	"os"
)

//go:embed go_version.conf
var GoVersion string

type Mode int

const (
	Script Mode = iota // download + exec external script
	GoInstall          // go install source@version
	Binary             // download GitHub release
)

type Tool struct {
	Mode          Mode
	Name          string
	Source        string // module path (GoInstall) or GitHub repo (Binary)
	ModuleVersion string // Go module version for go install (GoInstall only); defaults to "latest"
	Version       string // tool version to install and verify (e.g. TinyGo 0.40.1)
	VerifyBinary  string // binary to run for verification; defaults to Name if empty
	Required      bool   // true = always install; false = prompt user
	DependsOn     string // "" = no dependency; "git" = requires git installed first
}

var Tools = []Tool{
	{Mode: GoInstall, Name: "tinygoinstall", Source: "github.com/tinywasm/tinygo/cmd/tinygoinstall", ModuleVersion: "v0.0.6", Version: "0.40.1", VerifyBinary: "tinygo", Required: true},
	{Mode: Binary, Name: "tinywasm-cli", Source: "https://github.com/tinywasm/tinywasm", Version: "0.1.0", Required: false},
	{Mode: Binary, Name: "tinywasm-server", Source: "https://github.com/tinywasm/tinywasm", Version: "0.1.0", Required: false, DependsOn: "tinywasm-cli"},
}

type Installer struct{}

func New() *Installer {
	return &Installer{}
}

// InstallResult holds counts for the summary line.
type InstallResult struct {
	Installed int
	Failed    int
	Skipped   int
}

// InstallAll installs required tools + the optional tools at the given indices.
// Returns counts for summary output. On required tool failure, writes to stderr and exits.
func (ins *Installer) InstallAll(tools []Tool, selectedIdx []int, deps *Deps) InstallResult {
	// Build install list: required first, then selected optional
	var toInstall []Tool
	for _, t := range tools {
		if t.Required {
			toInstall = append(toInstall, t)
		}
	}
	for _, idx := range selectedIdx {
		toInstall = append(toInstall, tools[idx])
	}

	status := make(map[string]bool) // true = success
	var res InstallResult

	for _, t := range toInstall {
		if t.DependsOn != "" && !status[t.DependsOn] {
			fmt.Printf("⏭ %s — skipped (requires %s)\n", t.Name, t.DependsOn)
			status[t.Name] = false
			res.Skipped++
			continue
		}

		err := RunWithSpinner(t.Name, t.Version, func() error {
			var e error
			switch t.Mode {
			case GoInstall:
				e = ins.InstallGoInstall(t, deps)
			case Binary:
				e = ins.InstallBinary(t, deps)
			default:
				return fmt.Errorf("unknown mode")
			}
			if e != nil {
				return e
			}
			return ins.Verify(t, deps)
		})

		if err != nil {
			status[t.Name] = false
			res.Failed++
			if t.Required {
				fmt.Fprintf(os.Stderr, "FATAL: required tool %s failed: %v\n", t.Name, err)
				os.Exit(1)
			}
		} else {
			status[t.Name] = true
			res.Installed++
		}
	}

	return res
}
