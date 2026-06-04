package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// EXPECTED: the catalog installs the `tinywasm` binary from the public
// distribution repo tinywasm/app (not the outdated tinywasm-cli/tinywasm-server
// pointing at tinywasm/tinywasm). Single tool, Binary mode.
func TestTools_TinywasmFromAppRepo(t *testing.T) {
	var found *Tool
	for i := range Tools {
		if Tools[i].Name == "tinywasm" {
			found = &Tools[i]
			break
		}
	}
	if found == nil {
		t.Fatal("expected a 'tinywasm' tool installing the app binary")
	}
	if found.Mode != Binary {
		t.Errorf("tinywasm must use Binary mode, got %v", found.Mode)
	}
	if found.Source != "https://github.com/tinywasm/app" {
		t.Errorf("tinywasm must be sourced from tinywasm/app, got %q", found.Source)
	}
}

// EXPECTED: InstallBinary downloads the RAW gorelease asset
// (tinywasm-{os}-{arch}[.exe]) directly — no archive, no version-in-name,
// matching what `gorelease` publishes and the low-level curl fallback.
func TestInstallBinary_DownloadsRawGoreleaseAsset(t *testing.T) {
	ins := New()
	tool := Tool{
		Mode:    Binary,
		Name:    "tinywasm",
		Source:  "https://github.com/tinywasm/app",
		Version: "0.3.0",
	}

	var gotURL string
	d := &Deps{
		Download: func(url string) ([]byte, error) {
			gotURL = url
			// Stop here: we only assert the URL shape, not the install.
			return nil, fmt.Errorf("captured url")
		},
	}

	_ = ins.InstallBinary(tool, d)

	asset := fmt.Sprintf("tinywasm-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		asset += ".exe"
	}
	want := fmt.Sprintf("https://github.com/tinywasm/app/releases/download/v0.3.0/%s", asset)

	if gotURL != want {
		t.Errorf("expected raw gorelease asset URL\n want: %s\n  got: %s", want, gotURL)
	}
}

// EXPECTED: with no pinned Version, the installer must query the latest published
// release (GitHub API releases/latest) and use that tag for the download URL —
// so publishing a new tinywasm/app version does NOT require updating the installer.
func TestInstallBinary_ResolvesLatestVersion(t *testing.T) {
	ins := New()
	tool := Tool{
		Mode:    Binary,
		Name:    "tinywasm",
		Source:  "https://github.com/tinywasm/app",
		Version: "", // empty = resolve latest at runtime, do not hardcode
	}

	var assetURL string
	d := &Deps{
		Download: func(url string) ([]byte, error) {
			if strings.Contains(url, "api.github.com") && strings.Contains(url, "releases/latest") {
				return []byte(`{"tag_name":"v0.9.9"}`), nil
			}
			assetURL = url
			return nil, fmt.Errorf("captured asset url")
		},
	}

	_ = ins.InstallBinary(tool, d)

	if !strings.Contains(assetURL, "v0.9.9") {
		t.Errorf("expected asset URL to use the resolved latest tag v0.9.9, got: %q", assetURL)
	}
}

// #1 SECURITY: Test verifyChecksum helper directly
func TestVerifyChecksum_RejectsMismatch(t *testing.T) {
	data := []byte("binary-bytes")
	// SHA256 of "binary-bytes" is 4458...
	sums := "deadbeef  tinywasm-linux-amd64\n" // intentional mismatch
	if err := verifyChecksum("tinywasm-linux-amd64", data, []byte(sums)); err == nil {
		t.Fatal("expected error on checksum mismatch")
	}

	// Correct checksum
	sum := sha256.Sum256(data)
	correctSums := fmt.Sprintf("%s  tinywasm-linux-amd64\n", hex.EncodeToString(sum[:]))
	if err := verifyChecksum("tinywasm-linux-amd64", data, []byte(correctSums)); err != nil {
		t.Errorf("expected success on correct checksum, got: %v", err)
	}
}
