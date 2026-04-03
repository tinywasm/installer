package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/tinywasm/installer"
)

func main() {
	ins := installer.New()
	deps := &installer.Deps{
		RunCmd: func(name string, args ...string) ([]byte, error) {
			return exec.Command(name, args...).CombinedOutput()
		},
		Download: installer.DefaultDownload,
		WriteFile: func(path string, data []byte, perm os.FileMode) error {
			return os.WriteFile(path, data, perm)
		},
		RemoveFile: os.Remove,
		LookPath: exec.LookPath,
		Checklist: installer.ShowChecklist,
	}

	// 1. Check UNINSTALL env
	if os.Getenv("UNINSTALL") != "" {
		err := ins.Uninstall(installer.Tools, deps)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 2. Show checklist for optional tools
	selectedIdx := deps.Checklist(installer.Tools)

	// 3. Build install list: all required + selected optional
	var toInstall []installer.Tool
	for _, t := range installer.Tools {
		if t.Required {
			toInstall = append(toInstall, t)
		}
	}
	for _, idx := range selectedIdx {
		toInstall = append(toInstall, installer.Tools[idx])
	}

	// 4. Install all
	fmt.Printf("go — %s (script)\n", installer.GoVersion)
	installAll(ins, toInstall, deps)

	fmt.Println("\nDone.")
}

func installAll(ins *installer.Installer, tools []installer.Tool, deps *installer.Deps) {
	// Track completion status for dependency checks
	status := make(map[string]bool)

	for _, t := range tools {
		// Check dependency
		if t.DependsOn != "" {
			if !status[t.DependsOn] {
				fmt.Printf("⏭ %s — skipped (requires %s)\n", t.Name, t.DependsOn)
				status[t.Name] = false
				continue
			}
		}

		err := installer.RunWithSpinner(t.Name, func() error {
			var err error
			switch t.Mode {
			case installer.GoInstall:
				err = ins.InstallGoInstall(t, deps)
			case installer.Binary:
				err = ins.InstallBinary(t, deps)
			default:
				return fmt.Errorf("unknown mode")
			}
			if err != nil {
				return err
			}
			return ins.Verify(t, deps)
		})

		if err != nil {
			status[t.Name] = false
			if t.Required {
				fmt.Fprintf(os.Stderr, "FATAL: required tool %s failed to install\n", t.Name)
				os.Exit(1)
			}
		} else {
			status[t.Name] = true
		}
	}
}
