package installer

import (
	"fmt"
	"time"
)

// RunWithSpinner runs the given function while displaying a spinner.
func RunWithSpinner(name string, fn func() error) error {
	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	chars := []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")
	i := 0
	for {
		select {
		case err := <-done:
			if err != nil {
				fmt.Printf("\r\033[K❌ %s — %v\n", name, err)
			} else {
				fmt.Printf("\r\033[K✅ %s\n", name)
			}
			return err
		default:
			fmt.Printf("\r\033[K%c Installing %s...", chars[i%len(chars)], name)
			i++
			time.Sleep(100 * time.Millisecond)
		}
	}
}
