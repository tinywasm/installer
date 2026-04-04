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
		LookPath:   exec.LookPath,
		Checklist:  installer.ShowChecklist,
	}

	if os.Getenv("UNINSTALL") != "" {
		if err := ins.Uninstall(installer.Tools, deps); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	selectedIdx := deps.Checklist(installer.Tools)
	fmt.Printf("go — %s (script)\n", installer.GoVersion)

	res := ins.InstallAll(installer.Tools, selectedIdx, deps)
	fmt.Printf("\nDone. %d installed, %d failed, %d skipped.\n", res.Installed, res.Failed, res.Skipped)
}
