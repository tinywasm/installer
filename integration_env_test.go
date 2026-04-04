//go:build integration

package installer_test

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

// windowsEnv holds config for the Windows integration test.
// Values come from .env (gitignored) or environment variables.
// See .env.example for reference.
type windowsEnv struct {
	ip           string
	user         string
	pass         string
	vmName       string // virsh VM name   (win10)
	snapshotName string // snapshot name   (windows-update)
}

// loadEnvFile reads key=value pairs from a file into the process environment.
// Lines starting with # and blank lines are ignored. Does not overwrite existing vars.
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok || os.Getenv(k) != "" {
			continue
		}
		// strip inline comments (e.g. "value  # comment")
		if i := strings.Index(v, "#"); i >= 0 {
			v = v[:i]
		}
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		os.Setenv(k, v)
	}
}

func loadWindowsEnv(t *testing.T) windowsEnv {
	t.Helper()
	loadEnvFile(".env")

	e := windowsEnv{
		ip:           os.Getenv("WIN_IP"),
		user:         os.Getenv("WIN_USER"),
		pass:         os.Getenv("WIN_PASS"),
		vmName:       os.Getenv("WIN_VM_NAME"),
		snapshotName: os.Getenv("WIN_SNAPSHOT_NAME"),
	}
	if e.ip == "" || e.user == "" || e.vmName == "" || e.snapshotName == "" {
		t.Skip("WIN_IP / WIN_USER / WIN_VM_NAME / WIN_SNAPSHOT_NAME not set — copy .env.example to .env")
	}
	return e
}
