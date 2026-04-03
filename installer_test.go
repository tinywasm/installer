package installer

import (
	"fmt"
	"os"
	"testing"
)

func TestGoInstall_Success(t *testing.T) {
	ins := New()
	tool := Tool{Mode: GoInstall, Name: "tinygoinstall", Source: "github.com/tinywasm/tinygo/cmd/tinygoinstall", Version: "0.40.1", Required: true}
	d := &Deps{
		RunCmd: func(name string, args ...string) ([]byte, error) {
			if name == "go" && args[0] == "install" && args[1] == "github.com/tinywasm/tinygo/cmd/tinygoinstall@0.40.1" {
				return nil, nil
			}
			return nil, fmt.Errorf("unexpected command: %s %v", name, args)
		},
	}

	err := ins.InstallGoInstall(tool, d)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestVerify_Success(t *testing.T) {
	ins := New()
	tool := Tool{Name: "testtool", Version: "1.2.3"}
	d := &Deps{
		RunCmd: func(name string, args ...string) ([]byte, error) {
			if name == "testtool" && (args[0] == "version" || args[0] == "--version") {
				return []byte("version 1.2.3"), nil
			}
			return nil, fmt.Errorf("command failed")
		},
	}

	err := ins.Verify(tool, d)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

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

	err := ins.Uninstall(toolsList, d)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestChecklist_SelectNone(t *testing.T) {
	toolsList := []Tool{
		{Name: "git", Required: false},
	}

	// mock ShowChecklist as Checklist in Deps
	d := &Deps{
		Checklist: func(tools []Tool) []int {
			return []int{}
		},
	}

	selected := d.Checklist(toolsList)
	if len(selected) != 0 {
		t.Fatalf("expected 0 selected tools, got %d", len(selected))
	}
}
