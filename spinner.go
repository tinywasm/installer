package installer

import "fmt"

// RunWithSpinner runs fn printing a simple status line.
// Uses plain text output — no ANSI codes — for compatibility with
// PowerShell, Windows Terminal, and SSH sessions.
func RunWithSpinner(label, version string, fn func() error) error {
	fmt.Printf("Installing %s...\n", label)
	err := fn()
	if err != nil {
		fmt.Printf("❌ %s — failed\n", label)
	} else {
		fmt.Printf("✅ %s — %s\n", label, version)
	}
	return err
}
