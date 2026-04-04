//go:build integration

package installer_test

// TestWindowsConnectivity was a sanity check that verified SSH connectivity
// and KVM snapshot revert. Kept for reference but not run in CI.
//
// To re-enable: remove the comment markers around the function body and
// restore the imports (fmt, os/exec, strings, testing, time).
//
// func TestWindowsConnectivity(t *testing.T) {
// 	env := loadWindowsEnv(t)
//
// 	if _, err := exec.LookPath("virsh"); err != nil {
// 		t.Skip("virsh not found — install libvirt-clients")
// 	}
//
// 	const testFile = `C:\Users\user\Desktop\test_installer.txt`
//
// 	t.Log("Step 1: wait for Windows SSH")
// 	waitSSH(t, env, 1*time.Minute)
//
// 	t.Log("Step 2: create test file on desktop")
// 	runSSH(t, env, fmt.Sprintf(`New-Item -Path "%s" -ItemType File -Force | Out-Null`, testFile))
//
// 	t.Log("Step 3: verify file exists")
// 	out := runSSH(t, env, fmt.Sprintf(`(Test-Path "%s").ToString()`, testFile))
// 	if !strings.Contains(strings.ToLower(out), "true") {
// 		t.Fatalf("expected file to exist after creation, got: %q", out)
// 	}
// 	t.Log("file exists ✅")
//
// 	t.Log("Step 4: revert to clean snapshot")
// 	revertKVM(t, env)
//
// 	t.Log("Step 5: wait for Windows SSH")
// 	waitSSH(t, env, 1*time.Minute)
//
// 	t.Log("Step 6: verify file is gone after revert")
// 	out = runSSH(t, env, fmt.Sprintf(`(Test-Path "%s").ToString()`, testFile))
// 	if strings.Contains(strings.ToLower(out), "true") {
// 		t.Fatal("file still exists after snapshot revert — snapshot did not restore clean state")
// 	}
// 	t.Log("file is gone after revert ✅ — clean state confirmed")
// }
