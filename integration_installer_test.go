//go:build integration

package installer_test

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestInstallScript reverts to a clean Windows snapshot, runs install.ps1
// via SSH, and verifies that Go and the installer binary are present afterward.
//
// Prerequisites: OpenSSH enabled on Windows, virsh installed on host.
//
// Run:
//
//	go test -v -tags integration -run TestInstallScript ./...
func TestInstallScript(t *testing.T) {
	env := loadWindowsEnv(t)

	if _, err := exec.LookPath("virsh"); err != nil {
		t.Skip("virsh not found — install libvirt-clients")
	}

	t.Log("Step 1: revert to clean snapshot")
	revertKVM(t, env)

	t.Log("Step 2: wait for Windows SSH")
	waitSSH(t, env, 1*time.Minute)

	t.Log("Step 3: run install.ps1")
	// irm fetches the script from GitHub; iex executes it.
	// We use -EncodedCommand (via runSSH) so cmd.exe quoting is not an issue.
	runSSH(t, env, `irm https://raw.githubusercontent.com/tinywasm/installer/main/scripts/install.ps1 | iex`)

	t.Log("Step 4: verify Go is installed")
	out := runSSH(t, env, `& "C:\Program Files\Go\bin\go.exe" version`)
	if !strings.Contains(out, "go version") {
		t.Fatalf("Go not installed — got: %q", out)
	}
	t.Logf("Go installed: %s", strings.TrimSpace(out))

	t.Log("Step 5: verify installer binary is reachable")
	out = runSSH(t, env, `(Get-Command installer -ErrorAction SilentlyContinue) -ne $null`)
	if !strings.Contains(strings.ToLower(out), "true") {
		t.Fatalf("installer binary not found in PATH after install — got: %q", out)
	}
	t.Log("installer binary found in PATH ✅")

	t.Log("Step 6: verify TinyGo is installed")
	out = runSSH(t, env, `tinygo version`)
	if !strings.Contains(out, "0.40.1") {
		t.Fatalf("expected TinyGo 0.40.1, got: %q", out)
	}
	t.Logf("TinyGo installed: %s", strings.TrimSpace(out))
}
