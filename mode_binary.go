package installer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

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

func resolveLatestVersion(source string, d *Deps) (string, error) {
	// Source is expected to be https://github.com/owner/repo
	parts := strings.Split(strings.TrimSuffix(source, "/"), "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid source URL: %s", source)
	}
	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	data, err := d.Download(apiURL)
	if err != nil {
		return "", err
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(data, &release); err != nil {
		return "", fmt.Errorf("failed to parse GitHub API response: %w", err)
	}

	if release.TagName == "" {
		return "", fmt.Errorf("no tag_name found in GitHub API response")
	}

	return release.TagName, nil
}

func verifyChecksum(asset string, data []byte, sums []byte) error {
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])

	lines := strings.Split(string(sums), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		if parts[1] == asset {
			if parts[0] == got {
				return nil
			}
			return fmt.Errorf("checksum mismatch for %s: want %s, got %s", asset, parts[0], got)
		}
	}

	return fmt.Errorf("checksum for %s not found in checksums.txt", asset)
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

// DefaultDownload is a production implementation of Download
func DefaultDownload(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
