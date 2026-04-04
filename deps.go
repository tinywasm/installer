package installer

import (
	"os"
)

// Deps — all external operations behind interfaces for testability
type Deps struct {
	RunCmd      func(name string, args ...string) ([]byte, error)      // exec.Command
	Download    func(url string) ([]byte, error)                       // net/http GET
	WriteFile   func(path string, data []byte, perm os.FileMode) error // os.WriteFile
	RemoveFile  func(path string) error                                // os.Remove
	LookPath    func(name string) (string, error)                      // exec.LookPath
}
