package installer

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// InstallBinary handles the installation of tools from GitHub releases.
func (ins *Installer) InstallBinary(t Tool, d *Deps) error {
	osStr := runtime.GOOS
	archStr := runtime.GOARCH

	ext := "tar.gz"
	if osStr == "windows" {
		ext = "zip"
	}

	url := fmt.Sprintf("%s/releases/download/v%s/%s_v%s_%s_%s.%s",
		t.Source, t.Version, t.Name, t.Version, osStr, archStr, ext)

	data, err := d.Download(url)
	if err != nil {
		return fmt.Errorf("download failed from %s: %w", url, err)
	}

	tempDir, err := os.MkdirTemp("", "tinywasm-installer-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	binPath := ""
	if osStr == "windows" {
		binPath, err = extractZip(data, tempDir, t.Name)
	} else {
		binPath, err = extractTarGz(data, tempDir, t.Name)
	}
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Move to bin
	targetPath := getInstallPath(t.Name)
	src, err := os.Open(binPath)
	if err != nil {
		return fmt.Errorf("failed to open binary: %w", err)
	}
	defer src.Close()

	// Ensure the parent directory exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to open destination path %s: %w", targetPath, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
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

func extractTarGz(data []byte, dest, name string) (string, error) {
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if header.Typeflag == tar.TypeReg && (header.Name == name || strings.HasSuffix(header.Name, "/"+name)) {
			target := filepath.Join(dest, name)
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return "", err
			}
			f.Close()
			return target, nil
		}
	}
	return "", fmt.Errorf("binary %s not found in archive", name)
}

func extractZip(data []byte, dest, name string) (string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	for _, f := range r.File {
		if !f.FileInfo().IsDir() && (f.Name == name+".exe" || strings.HasSuffix(f.Name, "/"+name+".exe")) {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			target := filepath.Join(dest, name+".exe")
			dstFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return "", err
			}
			defer dstFile.Close()

			if _, err := io.Copy(dstFile, rc); err != nil {
				return "", err
			}
			return target, nil
		}
	}
	return "", fmt.Errorf("binary %s.exe not found in archive", name)
}

// DefaultDownload is a production implementation of Download
func DefaultDownload(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
