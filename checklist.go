package installer

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// ShowChecklist displays optional tools and returns the selected indices.
// Uses raw terminal mode to capture ↑/↓/Space/Enter without buffering.
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

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return []int{}
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return []int{}
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	selected := make([]bool, len(optionalTools))
	for i := range selected {
		selected[i] = true
	}
	cursor := 0

	fmt.Print("\033[?25l") // Hide cursor
	defer fmt.Print("\033[?25h") // Show cursor

	render := func() {
		fmt.Print("\r\033[KSelect optional tools to install:\n")
		for i, t := range optionalTools {
			fmt.Print("\r\033[K")
			if i == cursor {
				fmt.Print("> ")
			} else {
				fmt.Print("  ")
			}

			isDisabled := false
			if t.DependsOn != "" {
				// Check if dependency is selected
				depFound := false
				for j, opt := range optionalTools {
					if opt.Name == t.DependsOn {
						depFound = true
						if !selected[j] {
							isDisabled = true
						}
						break
					}
				}
				if !depFound {
					// Check required tools if not found in optional
					// (Assuming for now dependencies are in the same list or required)
				}
			}

			checkbox := "[ ]"
			if isDisabled {
				checkbox = "[-]"
				fmt.Print("\033[2m") // Dim
			} else if selected[i] {
				checkbox = "[x]"
			}

			indent := ""
			if t.DependsOn != "" {
				indent = "  └─ "
			}

			depInfo := ""
			if t.DependsOn != "" {
				depInfo = fmt.Sprintf(" (requires %s)", t.DependsOn)
			}

			fmt.Printf("%s%s %s — %s%s\n", checkbox, indent, t.Name, t.Version, depInfo)
			if isDisabled {
				fmt.Print("\033[0m") // Reset dim
			}
		}
		fmt.Print("\r\033[K\n\r\033[K  ↑/↓: move  Space: toggle  Enter: confirm")
		// Move cursor back up
		fmt.Printf("\033[%dA", len(optionalTools)+2)
	}

	render()

	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}

		if n == 1 {
			if buf[0] == 0x0D { // Enter
				break
			} else if buf[0] == 0x20 { // Space
				t := optionalTools[cursor]
				isDisabled := false
				if t.DependsOn != "" {
					for j, opt := range optionalTools {
						if opt.Name == t.DependsOn && !selected[j] {
							isDisabled = true
							break
						}
					}
				}
				if !isDisabled {
					selected[cursor] = !selected[cursor]
					// If unselecting a dependency, unselect dependents
					if !selected[cursor] {
						for i, opt := range optionalTools {
							if opt.DependsOn == t.Name {
								selected[i] = false
							}
						}
					}
				}
			} else if buf[0] == 0x71 || buf[0] == 0x03 { // q or Ctrl+C
				return []int{}
			}
		} else if n == 3 && buf[0] == 0x1b && buf[1] == '[' {
			if buf[2] == 'A' { // Up
				if cursor > 0 {
					cursor--
				}
			} else if buf[2] == 'B' { // Down
				if cursor < len(optionalTools)-1 {
					cursor++
				}
			}
		}
		render()
	}

	// Move cursor to the end
	fmt.Printf("\033[%dB\n", len(optionalTools)+2)

	var result []int
	for i, s := range selected {
		if s {
			result = append(result, originalIndices[i])
		}
	}
	return result
}
