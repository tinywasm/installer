package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Verify checks if the tool is correctly installed and has the expected version.
// If VerifyBinary is set, it is used instead of Name to run the version check.
// Resolves the binary via LookPath first; on Windows also checks Scoop shims
// so verification works even when shims are not in the current process PATH.
func (i *Installer) Verify(t Tool, d *Deps) error {
	binary := t.VerifyBinary
	if binary == "" {
		binary = t.Name
	}

	binary = resolveBinary(binary, d)

	expectedVersion := t.Version
	for _, arg := range []string{"version", "--version"} {
		out, err := d.RunCmd(binary, arg)
		if err == nil && strings.Contains(string(out), expectedVersion) {
			return nil
		}
	}
	return fmt.Errorf("%s: version %s not found", binary, expectedVersion)
}

// resolveBinary returns the absolute path for binary, trying LookPath first
// then the Scoop shims directory for Windows sessions where PATH may not include shims.
func resolveBinary(binary string, d *Deps) string {
	if d.LookPath != nil {
		if path, err := d.LookPath(binary); err == nil {
			return path
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		scoopBin := filepath.Join(home, "scoop", "shims", binary+".exe")
		if _, err := os.Stat(scoopBin); err == nil {
			return scoopBin
		}
	}
	return binary
}
