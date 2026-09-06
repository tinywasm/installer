package installer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"webtyp.com/update"
)

// InstallBinary handles the installation of tools from GitHub releases.
func (ins *Installer) InstallBinary(t Tool, d *Deps) error {
	version := t.Version
	if version == "" {
		v, err := update.ResolveLatestVersion(t.Source, d.Download)
		if err != nil {
			return fmt.Errorf("failed to resolve latest version: %w", err)
		}
		version = strings.TrimPrefix(v, "v")
	}

	osStr := runtime.GOOS
	archStr := runtime.GOARCH

	asset := fmt.Sprintf("%s-%s-%s", t.Name, osStr, archStr)
	if osStr == "windows" {
		asset += ".exe"
	}

	base := fmt.Sprintf("%s/releases/download/v%s", t.Source, version)
	url := fmt.Sprintf("%s/%s", base, asset)

	data, err := d.Download(url)
	if err != nil {
		return fmt.Errorf("download failed from %s: %w", url, err)
	}

	// #1 SECURITY: Verify checksum
	sums, err := d.Download(base + "/checksums.txt")
	if err != nil {
		return fmt.Errorf("failed to download checksums: %w", err)
	}
	if err := update.VerifyChecksum(asset, data, sums); err != nil {
		return err
	}

	// Move to bin
	targetPath := getInstallPath(t.Name)

	// Ensure the parent directory exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to open destination path %s: %w", targetPath, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	return nil
}

func getInstallPath(name string) string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, _ := os.UserHomeDir()
		gopath = filepath.Join(home, "go")
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(gopath, "bin", name+".exe")
	}
	return filepath.Join(gopath, "bin", name)
}
