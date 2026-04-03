package installer

import (
	"fmt"
	"strings"
)

// Verify checks if the tool is correctly installed and has the expected version.
func (i *Installer) Verify(t Tool, d *Deps) error {
	expectedVersion := t.Version
	// Try "name version" first, then "name --version"
	for _, arg := range []string{"version", "--version"} {
		out, err := d.RunCmd(t.Name, arg)
		if err == nil && strings.Contains(string(out), expectedVersion) {
			return nil
		}
	}
	return fmt.Errorf("%s: version %s not found", t.Name, expectedVersion)
}
