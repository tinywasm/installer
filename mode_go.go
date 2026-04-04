package installer

import (
	"fmt"
)

// InstallGoInstall installs a tool via `go install source@moduleVersion`,
// then runs the installed binary with `-version t.Version` to perform
// the actual tool installation (e.g. tinygoinstall installs TinyGo).
func (i *Installer) InstallGoInstall(t Tool, d *Deps) error {
	moduleVersion := t.ModuleVersion
	if moduleVersion == "" {
		moduleVersion = "latest"
	}
	source := fmt.Sprintf("%s@%s", t.Source, moduleVersion)
	if _, err := d.RunCmd("go", "install", source); err != nil {
		return fmt.Errorf("go install failed: %w", err)
	}
	if t.Version != "" {
		if _, err := d.RunCmd(t.Name, "-version", t.Version); err != nil {
			return fmt.Errorf("%s -version %s failed: %w", t.Name, t.Version, err)
		}
	}
	return nil
}
