package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/tinywasm/installer"
	"github.com/tinywasm/update"
)

func main() {
	tools := flag.String("tools", "", "optional tools to install: all, or comma-separated names (e.g. tinywasm-cli,tinywasm-server)")
	flag.Parse()

	ins := installer.New()
	deps := &installer.Deps{
		RunCmd: func(name string, args ...string) ([]byte, error) {
			return exec.Command(name, args...).CombinedOutput()
		},
		Download: update.DefaultDownload,
		WriteFile: func(path string, data []byte, perm os.FileMode) error {
			return os.WriteFile(path, data, perm)
		},
		RemoveFile: os.Remove,
		LookPath:   exec.LookPath,
	}

	if os.Getenv("UNINSTALL") != "" {
		if err := ins.Uninstall(installer.Tools, deps); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	selectedIdx := installer.ResolveTools(installer.Tools, *tools)
	fmt.Printf("go — %s (script)\n", installer.GoVersion)

	res := ins.InstallAll(installer.Tools, selectedIdx, deps)
	fmt.Printf("\nDone. %d installed, %d failed, %d skipped.\n", res.Installed, res.Failed, res.Skipped)
}

