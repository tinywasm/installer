package installer

import (
	"fmt"
	"os"
)

// Uninstall removes tools according to the UNINSTALL environment variable.
func (ins *Installer) Uninstall(tools []Tool, d *Deps) error {
	uninstallValue := os.Getenv("UNINSTALL")
	if uninstallValue == "" {
		return nil
	}

	if uninstallValue == "all" {
		// iterate in reverse order
		for i := len(tools) - 1; i >= 0; i-- {
			t := tools[i]
			err := removeOne(t, d)
			if err == nil {
				fmt.Printf("✅ %s — removed\n", t.Name)
			} else {
				fmt.Printf("⏭ %s — skip removal (not found)\n", t.Name)
			}
		}
		return nil
	}

	for _, t := range tools {
		if t.Name == uninstallValue {
			err := removeOne(t, d)
			if err != nil {
				return fmt.Errorf("failed to remove %s: %w", t.Name, err)
			}
			fmt.Printf("✅ %s — removed\n", t.Name)
			return nil
		}
	}

	return fmt.Errorf("tool %s not found in registry", uninstallValue)
}

func removeOne(t Tool, d *Deps) error {
	path, err := d.LookPath(t.Name)
	if err != nil {
		return err
	}
	return d.RemoveFile(path)
}
