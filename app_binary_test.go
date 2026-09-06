package installer

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// EXPECTED: the catalog installs the `webtyp` binary from the public
// distribution repo webtyp/app (not the outdated webtyp-cli/webtyp-server
// pointing at webtyp/webtyp). Single tool, Binary mode.
func TestTools_WebTypFromAppRepo(t *testing.T) {
	var found *Tool
	for i := range Tools {
		if Tools[i].Name == "webtyp" {
			found = &Tools[i]
			break
		}
	}
	if found == nil {
		t.Fatal("expected a 'webtyp' tool installing the app binary")
	}
	if found.Mode != Binary {
		t.Errorf("webtyp must use Binary mode, got %v", found.Mode)
	}
	if found.Source != "https://github.com/webtyp/app" {
		t.Errorf("webtyp must be sourced from webtyp/app, got %q", found.Source)
	}
}

// EXPECTED: InstallBinary downloads the RAW gorelease asset
// (webtyp-{os}-{arch}[.exe]) directly — no archive, no version-in-name,
// matching what `gorelease` publishes and the low-level curl fallback.
func TestInstallBinary_DownloadsRawGoreleaseAsset(t *testing.T) {
	ins := New()
	tool := Tool{
		Mode:    Binary,
		Name:    "webtyp",
		Source:  "https://github.com/webtyp/app",
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

	asset := fmt.Sprintf("webtyp-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		asset += ".exe"
	}
	want := fmt.Sprintf("https://github.com/webtyp/app/releases/download/v0.3.0/%s", asset)

	if gotURL != want {
		t.Errorf("expected raw gorelease asset URL\n want: %s\n  got: %s", want, gotURL)
	}
}

// EXPECTED: with no pinned Version, the installer must query the latest published
// release (GitHub API releases/latest) and use that tag for the download URL —
// so publishing a new webtyp/app version does NOT require updating the installer.
func TestInstallBinary_ResolvesLatestVersion(t *testing.T) {
	ins := New()
	tool := Tool{
		Mode:    Binary,
		Name:    "webtyp",
		Source:  "https://github.com/webtyp/app",
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

// NOTE (self-update refactor): TestVerifyChecksum_RejectsMismatch was MOVED to
// webtyp.com/update (checksum_test.go: TestVerifyChecksum). The installer
// now consumes update.VerifyChecksum, so the contract is tested there.
