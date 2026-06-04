package installer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// NOTE (self-update refactor): verifyChecksum, resolveLatestVersion and
// DefaultDownload were MOVED to github.com/tinywasm/update. This file no longer
// compiles until InstallBinary and cmd/tinywasm-installer are rewired to call
// update.VerifyChecksum / update.ResolveLatestVersion / update.DefaultDownload.
// See docs/PLAN.md.

// InstallBinary handles the installation of tools from GitHub releases.
func (ins *Installer) InstallBinary(t Tool, d *Deps) error {
	version := t.Version
	if version == "" {
		v, err := resolveLatestVersion(t.Source, d)
		if err != nil {
			return fmt.Errorf("failed to resolve latest version: %w", err)
		}
		version = v
		// Strip 'v' prefix if present for URL construction
		version = strings.TrimPrefix(version, "v")
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
	if err := verifyChecksum(asset, data, sums); err != nil {
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
