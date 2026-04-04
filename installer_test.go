package installer

import (
	"fmt"
	"os"
	"testing"
)

// --- GoInstall ---

func TestGoInstall_Success(t *testing.T) {
	ins := New()
	tool := Tool{Mode: GoInstall, Name: "tinygoinstall", Source: "github.com/tinywasm/tinygo/cmd/tinygoinstall", ModuleVersion: "latest", Version: "0.40.1", Required: true}
	d := &Deps{
		RunCmd: func(name string, args ...string) ([]byte, error) {
			if name == "go" && args[0] == "install" && args[1] == "github.com/tinywasm/tinygo/cmd/tinygoinstall@latest" {
				return nil, nil
			}
			if name == "tinygoinstall" && args[0] == "-v" && args[1] == "-version" && args[2] == "0.40.1" {
				return nil, nil
			}
			return nil, fmt.Errorf("unexpected command: %s %v", name, args)
		},
	}
	if err := ins.InstallGoInstall(tool, d); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestGoInstall_Fail_Optional(t *testing.T) {
	ins := New()
	tool := Tool{Mode: GoInstall, Name: "sometool", Source: "github.com/x/sometool", Version: "1.0.0", Required: false}
	d := &Deps{
		RunCmd: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("install failed")
		},
	}
	if err := ins.InstallGoInstall(tool, d); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Verify ---

var lookPathNotFound = func(string) (string, error) { return "", fmt.Errorf("not found") }

func TestVerify_Success(t *testing.T) {
	ins := New()
	tool := Tool{Name: "testtool", Version: "1.2.3"}
	d := &Deps{
		LookPath: lookPathNotFound,
		RunCmd: func(name string, args ...string) ([]byte, error) {
			if name == "testtool" && (args[0] == "version" || args[0] == "--version") {
				return []byte("version 1.2.3"), nil
			}
			return nil, fmt.Errorf("command failed")
		},
	}
	if err := ins.Verify(tool, d); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestVerify_DashDashVersion(t *testing.T) {
	ins := New()
	tool := Tool{Name: "testtool", Version: "2.0.0"}
	d := &Deps{
		LookPath: lookPathNotFound,
		RunCmd: func(name string, args ...string) ([]byte, error) {
			if args[0] == "--version" {
				return []byte("testtool 2.0.0"), nil
			}
			return nil, fmt.Errorf("version subcommand not found")
		},
	}
	if err := ins.Verify(tool, d); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestVerify_Mismatch(t *testing.T) {
	ins := New()
	tool := Tool{Name: "testtool", Version: "1.2.3"}
	d := &Deps{
		LookPath: lookPathNotFound,
		RunCmd: func(name string, args ...string) ([]byte, error) {
			return []byte("testtool 9.9.9"), nil
		},
	}
	if err := ins.Verify(tool, d); err == nil {
		t.Fatal("expected error for version mismatch, got nil")
	}
}

// --- Binary ---

func TestBinary_DownloadFail(t *testing.T) {
	ins := New()
	tool := Tool{Mode: Binary, Name: "mytool", Source: "https://github.com/org/mytool", Version: "1.0.0"}
	d := &Deps{
		Download: func(url string) ([]byte, error) {
			return nil, fmt.Errorf("network error")
		},
	}
	if err := ins.InstallBinary(tool, d); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- ResolveTools ---

func TestResolveTools_Empty(t *testing.T) {
	tools := []Tool{
		{Name: "req", Required: true},
		{Name: "opt1", Required: false},
	}
	if got := ResolveTools(tools, ""); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestResolveTools_All(t *testing.T) {
	tools := []Tool{
		{Name: "req", Required: true},
		{Name: "opt1", Required: false},
		{Name: "opt2", Required: false},
	}
	got := ResolveTools(tools, "all")
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("expected [1,2], got %v", got)
	}
}

func TestResolveTools_Specific(t *testing.T) {
	tools := []Tool{
		{Name: "req", Required: true},
		{Name: "opt1", Required: false},
		{Name: "opt2", Required: false},
	}
	got := ResolveTools(tools, "opt2")
	if len(got) != 1 || got[0] != 2 {
		t.Fatalf("expected [2], got %v", got)
	}
}

func TestResolveTools_Unknown(t *testing.T) {
	tools := []Tool{
		{Name: "opt1", Required: false},
	}
	got := ResolveTools(tools, "nonexistent")
	if len(got) != 0 {
		t.Fatalf("expected [], got %v", got)
	}
}

// --- InstallAll: Required vs Optional ---

func TestRequired_AlwaysInstalled(t *testing.T) {
	ins := New()
	installed := []string{}
	tools := []Tool{
		{Mode: GoInstall, Name: "req-tool", Source: "github.com/x/req-tool", Version: "1.0.0", Required: true},
	}
	d := &Deps{
		RunCmd: func(name string, args ...string) ([]byte, error) {
			if name == "go" {
				return nil, nil
			}
			if name == "req-tool" {
				installed = append(installed, name)
				return []byte("req-tool 1.0.0"), nil
			}
			return nil, fmt.Errorf("unexpected: %s", name)
		},
	}
	res := ins.InstallAll(tools, nil, d)
	if res.Installed != 1 {
		t.Fatalf("expected 1 installed, got %d", res.Installed)
	}
	_ = installed
}

func TestOptional_ContinuesOnFail(t *testing.T) {
	ins := New()
	tools := []Tool{
		{Mode: GoInstall, Name: "opt1", Source: "github.com/x/opt1", Version: "1.0.0", Required: false},
		{Mode: GoInstall, Name: "opt2", Source: "github.com/x/opt2", Version: "2.0.0", Required: false},
	}
	d := &Deps{
		RunCmd: func(name string, args ...string) ([]byte, error) {
			if name == "go" && len(args) > 1 && args[1] == "github.com/x/opt1@1.0.0" {
				return nil, fmt.Errorf("install failed")
			}
			if name == "go" {
				return nil, nil
			}
			if name == "opt2" {
				return []byte("opt2 2.0.0"), nil
			}
			return nil, fmt.Errorf("unexpected: %s %v", name, args)
		},
	}
	res := ins.InstallAll(tools, []int{0, 1}, d)
	if res.Failed != 1 {
		t.Fatalf("expected 1 failed, got %d", res.Failed)
	}
	if res.Installed != 1 {
		t.Fatalf("expected 1 installed (opt2), got %d", res.Installed)
	}
}

// --- InstallAll: Dependencies ---

func TestDep_SkipOnFail(t *testing.T) {
	ins := New()
	tools := []Tool{
		{Mode: GoInstall, Name: "git", Source: "github.com/x/git", Version: "2.47.0", Required: false},
		{Mode: GoInstall, Name: "gh", Source: "github.com/x/gh", Version: "2.65.0", Required: false, DependsOn: "git"},
	}
	d := &Deps{
		RunCmd: func(name string, args ...string) ([]byte, error) {
			if name == "go" {
				return nil, fmt.Errorf("git install failed")
			}
			return nil, fmt.Errorf("unexpected: %s", name)
		},
	}
	res := ins.InstallAll(tools, []int{0, 1}, d)
	if res.Failed != 1 {
		t.Fatalf("expected 1 failed (git), got %d", res.Failed)
	}
	if res.Skipped != 1 {
		t.Fatalf("expected 1 skipped (gh), got %d", res.Skipped)
	}
}

func TestDep_SkipOnSkip(t *testing.T) {
	ins := New()
	tools := []Tool{
		{Mode: GoInstall, Name: "git", Source: "github.com/x/git", Version: "2.47.0", Required: false},
		{Mode: GoInstall, Name: "gh", Source: "github.com/x/gh", Version: "2.65.0", Required: false, DependsOn: "git"},
	}
	d := &Deps{}
	// Select only gh (idx 1), not git (idx 0) — gh depends on git which wasn't selected/installed
	res := ins.InstallAll(tools, []int{1}, d)
	if res.Skipped != 1 {
		t.Fatalf("expected 1 skipped (gh because git not in status map), got %d", res.Skipped)
	}
}

func TestDep_Success(t *testing.T) {
	ins := New()
	tools := []Tool{
		{Mode: GoInstall, Name: "git", Source: "github.com/x/git", Version: "2.47.0", Required: false},
		{Mode: GoInstall, Name: "gh", Source: "github.com/x/gh", Version: "2.65.0", Required: false, DependsOn: "git"},
	}
	d := &Deps{
		RunCmd: func(name string, args ...string) ([]byte, error) {
			if name == "go" {
				return nil, nil
			}
			if name == "git" {
				return []byte("git version 2.47.0"), nil
			}
			if name == "gh" {
				return []byte("gh version 2.65.0"), nil
			}
			return nil, fmt.Errorf("unexpected: %s", name)
		},
	}
	res := ins.InstallAll(tools, []int{0, 1}, d)
	if res.Installed != 2 {
		t.Fatalf("expected 2 installed, got %d", res.Installed)
	}
}

// --- Summary ---

func TestSummary_AllSuccess(t *testing.T) {
	ins := New()
	versions := map[string]string{
		"tool1": "1.0.0",
		"tool2": "2.0.0",
		"tool3": "3.0.0",
	}
	tools := []Tool{
		{Mode: GoInstall, Name: "tool1", Source: "github.com/x/tool1", Version: "1.0.0", Required: true},
		{Mode: GoInstall, Name: "tool2", Source: "github.com/x/tool2", Version: "2.0.0", Required: false},
		{Mode: GoInstall, Name: "tool3", Source: "github.com/x/tool3", Version: "3.0.0", Required: false},
	}
	d := &Deps{
		RunCmd: func(name string, args ...string) ([]byte, error) {
			if name == "go" {
				return nil, nil
			}
			if v, ok := versions[name]; ok {
				return []byte(name + " " + v), nil
			}
			return nil, fmt.Errorf("unexpected: %s", name)
		},
	}
	// tool1 is required (always); tool2=idx1, tool3=idx2 are optional selected
	res := ins.InstallAll(tools, []int{1, 2}, d)
	if res.Installed != 3 || res.Failed != 0 || res.Skipped != 0 {
		t.Fatalf("expected 3/0/0 got %d/%d/%d", res.Installed, res.Failed, res.Skipped)
	}
}

func TestSummary_Mixed(t *testing.T) {
	ins := New()
	tools := []Tool{
		{Mode: GoInstall, Name: "tool1", Source: "github.com/x/tool1", Version: "1.0.0", Required: false},
		{Mode: GoInstall, Name: "tool2", Source: "github.com/x/tool2", Version: "2.0.0", Required: false},
		{Mode: GoInstall, Name: "tool3", Source: "github.com/x/tool3", Version: "3.0.0", Required: false},
	}
	callCount := 0
	d := &Deps{
		RunCmd: func(name string, args ...string) ([]byte, error) {
			callCount++
			if name == "go" && callCount <= 1 {
				// First go install (tool1) fails
				return nil, fmt.Errorf("install failed")
			}
			if name == "go" {
				return nil, nil
			}
			if name == "tool2" {
				return []byte("tool2 2.0.0"), nil
			}
			return nil, fmt.Errorf("unexpected: %s", name)
		},
	}
	// Select tool1 (fails), tool2 (ok); tool3 not selected → skipped
	res := ins.InstallAll(tools, []int{0, 1}, d)
	if res.Failed != 1 {
		t.Fatalf("expected 1 failed, got %d", res.Failed)
	}
	if res.Installed != 1 {
		t.Fatalf("expected 1 installed, got %d", res.Installed)
	}
}

func TestSummary_DepSkip(t *testing.T) {
	ins := New()
	tools := []Tool{
		{Mode: GoInstall, Name: "git", Source: "github.com/x/git", Version: "2.47.0", Required: false},
		{Mode: GoInstall, Name: "gh", Source: "github.com/x/gh", Version: "2.65.0", Required: false, DependsOn: "git"},
	}
	d := &Deps{
		RunCmd: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("git install failed")
		},
	}
	res := ins.InstallAll(tools, []int{0, 1}, d)
	if res.Failed != 1 || res.Skipped != 1 || res.Installed != 0 {
		t.Fatalf("expected 0/1/1 got %d/%d/%d", res.Installed, res.Failed, res.Skipped)
	}
}

// --- Uninstall ---

func TestUninstall_Specific(t *testing.T) {
	os.Setenv("UNINSTALL", "tinygoinstall")
	defer os.Unsetenv("UNINSTALL")

	ins := New()
	toolsList := []Tool{
		{Name: "tinygoinstall"},
	}
	d := &Deps{
		LookPath: func(name string) (string, error) {
			if name == "tinygoinstall" {
				return "/bin/tinygoinstall", nil
			}
			return "", fmt.Errorf("not found")
		},
		RemoveFile: func(path string) error {
			if path == "/bin/tinygoinstall" {
				return nil
			}
			return fmt.Errorf("wrong path")
		},
	}
	if err := ins.Uninstall(toolsList, d); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestUninstall_All(t *testing.T) {
	os.Setenv("UNINSTALL", "all")
	defer os.Unsetenv("UNINSTALL")

	ins := New()
	toolsList := []Tool{
		{Name: "tool1"},
		{Name: "tool2"},
	}
	removed := []string{}
	d := &Deps{
		LookPath: func(name string) (string, error) {
			return "/bin/" + name, nil
		},
		RemoveFile: func(path string) error {
			removed = append(removed, path)
			return nil
		},
	}
	if err := ins.Uninstall(toolsList, d); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	// Should be removed in reverse order
	if len(removed) != 2 || removed[0] != "/bin/tool2" || removed[1] != "/bin/tool1" {
		t.Fatalf("expected reverse order removal, got %v", removed)
	}
}

func TestUninstall_NotFound(t *testing.T) {
	os.Setenv("UNINSTALL", "all")
	defer os.Unsetenv("UNINSTALL")

	ins := New()
	toolsList := []Tool{{Name: "missing-tool"}}
	d := &Deps{
		LookPath: func(name string) (string, error) {
			return "", fmt.Errorf("not found")
		},
		RemoveFile: func(path string) error {
			return fmt.Errorf("should not be called")
		},
	}
	// Should not error — just skip tools not found (UNINSTALL=all)
	if err := ins.Uninstall(toolsList, d); err != nil {
		t.Fatalf("expected nil error for missing tool in all-uninstall, got %v", err)
	}
}
