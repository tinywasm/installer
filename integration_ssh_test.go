//go:build integration

package installer_test

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
)

// waitSSH polls until a real SSH command succeeds or timeout expires.
// Polling TCP port alone is not enough — Windows SSH service may accept
// the connection before PowerShell is ready to run commands.
func waitSSH(t *testing.T, env windowsEnv, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	addr := env.ip + ":22"
	for time.Now().Before(deadline) {
		// First check TCP reachability (cheap).
		conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
		if err != nil {
			t.Logf("waiting for SSH at %s...", addr)
			time.Sleep(1 * time.Second)
			continue
		}
		conn.Close()

		// Then verify PowerShell is actually ready.
		encoded := encodePSCommand("Write-Output ready")
		sshArgs := []string{
			"-o", "StrictHostKeyChecking=no",
			"-o", "ConnectTimeout=5",
			fmt.Sprintf("%s@%s", env.user, env.ip),
			"powershell.exe", "-NoProfile", "-NonInteractive", "-EncodedCommand", encoded,
		}
		var cmd *exec.Cmd
		if _, err := exec.LookPath("sshpass"); err == nil && env.pass != "" {
			cmd = exec.Command("sshpass", append([]string{"-p", env.pass, "ssh"}, sshArgs...)...)
		} else {
			cmd = exec.Command("ssh", sshArgs...)
		}
		if out, err := cmd.Output(); err == nil && strings.Contains(string(out), "ready") {
			t.Logf("SSH reachable at %s ✅", addr)
			return
		}
		t.Logf("SSH port up but PowerShell not ready yet, retrying...")
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("Windows not reachable at %s within %s", addr, timeout)
}

// runSSH executes a PowerShell command on remote Windows via SSH.
// Uses key-based auth, or sshpass if WIN_PASS is set and sshpass is installed.
// The command is encoded as base64 to avoid shell quoting issues with Windows
// OpenSSH (which uses cmd.exe as the default shell).
func runSSH(t *testing.T, env windowsEnv, psCmd string) string {
	t.Helper()

	// Suppress CLIXML progress noise and encode as UTF-16LE base64 to avoid
	// quoting issues when cmd.exe is the default SSH shell on Windows.
	encoded := encodePSCommand("$ProgressPreference='SilentlyContinue'; " + psCmd)

	sshArgs := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "ConnectTimeout=10",
		fmt.Sprintf("%s@%s", env.user, env.ip),
		"powershell.exe", "-NoProfile", "-NonInteractive", "-EncodedCommand", encoded,
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

// encodePSCommand encodes a PowerShell command as UTF-16LE base64,
// suitable for powershell.exe -EncodedCommand. This avoids quoting
// issues when Windows OpenSSH's default shell (cmd.exe) parses the args.
func encodePSCommand(cmd string) string {
	runes := utf16.Encode([]rune(cmd))
	buf := make([]byte, len(runes)*2)
	for i, r := range runes {
		binary.LittleEndian.PutUint16(buf[i*2:], r)
	}
	return base64.StdEncoding.EncodeToString(buf)
}
