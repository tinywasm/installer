package installer

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ShowChecklist displays optional tools as a numbered list and returns the
// selected indices. Reads plain text input — no ANSI codes or raw mode —
// so it works correctly in PowerShell, SSH sessions, and any terminal.
func ShowChecklist(tools []Tool) []int {
	var optionalTools []Tool
	var originalIndices []int
	for i, t := range tools {
		if !t.Required {
			optionalTools = append(optionalTools, t)
			originalIndices = append(originalIndices, i)
		}
	}

	if len(optionalTools) == 0 {
		return []int{}
	}

	if !isInteractive() {
		return []int{}
	}

	fmt.Println("Select optional tools to install:")
	for i, t := range optionalTools {
		indent := ""
		dep := ""
		if t.DependsOn != "" {
			indent = "  └─ "
			dep = fmt.Sprintf(" (requires %s)", t.DependsOn)
		}
		fmt.Printf("  [%d] %s%s — %s%s\n", i+1, indent, t.Name, t.Version, dep)
	}
	fmt.Print("\nEnter numbers to install (e.g. 1,2), or press Enter to install all: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	line := strings.TrimSpace(scanner.Text())

	// Empty input → install all.
	if line == "" {
		return originalIndices
	}

	selected := map[int]bool{}
	for _, part := range strings.Split(line, ",") {
		part = strings.TrimSpace(part)
		n, err := strconv.Atoi(part)
		if err != nil || n < 1 || n > len(optionalTools) {
			continue
		}
		selected[n-1] = true
	}

	// Enforce dependencies: if a tool is selected but its dependency is not,
	// skip it and warn.
	var result []int
	for i, t := range optionalTools {
		if !selected[i] {
			continue
		}
		if t.DependsOn != "" {
			depSelected := false
			for j, opt := range optionalTools {
				if opt.Name == t.DependsOn && selected[j] {
					depSelected = true
					break
				}
			}
			if !depSelected {
				fmt.Printf("  Skipping %s: requires %s\n", t.Name, t.DependsOn)
				continue
			}
		}
		result = append(result, originalIndices[i])
	}
	return result
}

// isInteractive returns true when stdin is attached to a real terminal.
func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
