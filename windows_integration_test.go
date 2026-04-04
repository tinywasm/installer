//go:build integration

package installer_test

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
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

// TestWindowsConnectivity creates a file on the Windows desktop,
// reverts to the clean KVM snapshot, then verifies the file is gone.
//
// Prerequisites: OpenSSH enabled on Windows, virsh installed on host.
// See TESTING.md for one-time setup steps.
//
// Run:
//
//	go test -v -tags integration -run TestWindowsConnectivity ./...
func TestWindowsConnectivity(t *testing.T) {
	env := loadWindowsEnv(t)

	if _, err := exec.LookPath("virsh"); err != nil {
		t.Skip("virsh not found — install libvirt-clients")
	}

	const testFile = `C:\Users\user\Desktop\test_installer.txt`

	t.Log("Step 1: revert to clean snapshot")
	revertKVM(t, env)

	t.Log("Step 2: wait for Windows SSH")
	waitSSH(t, env.ip, 3*time.Minute)

	t.Log("Step 3: create test file on desktop")
	runSSH(t, env, fmt.Sprintf(`New-Item -Path "%s" -ItemType File -Force | Out-Null`, testFile))

	t.Log("Step 4: verify file exists")
	out := runSSH(t, env, fmt.Sprintf(`(Test-Path "%s").ToString()`, testFile))
	if !strings.Contains(strings.ToLower(out), "true") {
		t.Fatalf("expected file to exist after creation, got: %q", out)
	}
	t.Log("file exists ✅")

	t.Log("Step 5: revert to clean snapshot")
	revertKVM(t, env)

	t.Log("Step 6: wait for Windows SSH")
	waitSSH(t, env.ip, 3*time.Minute)

	t.Log("Step 7: verify file is gone after revert")
	out = runSSH(t, env, fmt.Sprintf(`(Test-Path "%s").ToString()`, testFile))
	if strings.Contains(strings.ToLower(out), "true") {
		t.Fatal("file still exists after snapshot revert — snapshot did not restore clean state")
	}
	t.Log("file is gone after revert ✅ — clean state confirmed")
}

func virsh(args ...string) *exec.Cmd {
	return exec.Command("virsh", append([]string{"-c", "qemu:///system"}, args...)...)
}

func revertKVM(t *testing.T, env windowsEnv) {
	t.Helper()
	virsh("shutdown", env.vmName).Run()
	time.Sleep(5 * time.Second)

	if out, err := virsh("snapshot-revert", env.vmName, env.snapshotName).CombinedOutput(); err != nil {
		t.Fatalf("virsh snapshot-revert: %v\n%s", err, out)
	}
	// snapshot was taken while running → domain already active after revert
	out, _ := virsh("domstate", env.vmName).Output()
	if !strings.Contains(string(out), "running") {
		if out, err := virsh("start", env.vmName).CombinedOutput(); err != nil {
			t.Fatalf("virsh start: %v\n%s", err, out)
		}
	}
	t.Logf("reverted %s → %s", env.vmName, env.snapshotName)
}

// waitSSH polls port 22 until Windows is reachable or timeout expires.
func waitSSH(t *testing.T, ip string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	addr := ip + ":22"
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
		if err == nil {
			conn.Close()
			time.Sleep(15 * time.Second) // extra wait for Windows services
			t.Logf("SSH reachable at %s ✅", addr)
			return
		}
		t.Logf("waiting for SSH at %s...", addr)
		time.Sleep(10 * time.Second)
	}
	t.Fatalf("Windows not reachable at %s within %s", addr, timeout)
}

// runSSH executes a PowerShell command on remote Windows via SSH.
// Uses key-based auth, or sshpass if WIN_PASS is set and sshpass is installed.
func runSSH(t *testing.T, env windowsEnv, psCmd string) string {
	t.Helper()
	sshArgs := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "ConnectTimeout=10",
		fmt.Sprintf("%s@%s", env.user, env.ip),
		"powershell.exe", "-NoProfile", "-NonInteractive", "-Command", psCmd,
	}

	var cmd *exec.Cmd
	if _, err := exec.LookPath("sshpass"); err == nil && env.pass != "" {
		cmd = exec.Command("sshpass", append([]string{"-p", env.pass, "ssh"}, sshArgs...)...)
	} else {
		cmd = exec.Command("ssh", sshArgs...)
	}

	out, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(out))
	if err != nil {
		t.Fatalf("SSH failed: %v\ncmd: %s\nout: %s", err, psCmd, result)
	}
	t.Logf("PS> %s\n→ %q", psCmd, result)
	return result
}
