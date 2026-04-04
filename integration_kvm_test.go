//go:build integration

package installer_test

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

func revertKVM(t *testing.T, env windowsEnv) {
	t.Helper()
	virsh("shutdown", env.vmName).Run()
	time.Sleep(2 * time.Second)

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

func virsh(args ...string) *exec.Cmd {
	return exec.Command("virsh", append([]string{"-c", "qemu:///system"}, args...)...)
}
