package installer

import (
	"fmt"
)

// InstallGoInstall handles the installation of tools using `go install source@version`.
func (i *Installer) InstallGoInstall(t Tool, d *Deps) error {
	source := fmt.Sprintf("%s@%s", t.Source, t.Version)
	_, err := d.RunCmd("go", "install", source)
	if err != nil {
		return fmt.Errorf("go install failed: %w", err)
	}
	return nil
}
