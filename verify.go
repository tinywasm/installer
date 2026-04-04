package installer

import (
	"fmt"
	"strings"
)

// Verify checks if the tool is correctly installed and has the expected version.
// If VerifyBinary is set, it is used instead of Name to run the version check.
func (i *Installer) Verify(t Tool, d *Deps) error {
	binary := t.VerifyBinary
	if binary == "" {
		binary = t.Name
	}
	expectedVersion := t.Version
	for _, arg := range []string{"version", "--version"} {
		out, err := d.RunCmd(binary, arg)
		if err == nil && strings.Contains(string(out), expectedVersion) {
			return nil
		}
	}
	return fmt.Errorf("%s: version %s not found", binary, expectedVersion)
}
