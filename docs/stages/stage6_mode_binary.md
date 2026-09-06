# Stage 6 — Binary download mode (Go)

## Goal

Implement `Binary` mode: download pre-built binary from GitHub releases.

## URL pattern

```
https://github.com/webtyp/{name}/releases/download/v{version}/{name}_v{version}_{os}_{arch}.tar.gz
```

## Steps

- [ ] Build download URL from tool name, version, `runtime.GOOS`, `runtime.GOARCH`
- [ ] Download with `net/http`
- [ ] Extract: `archive/tar` + `compress/gzip` (Linux/macOS) or `archive/zip` (Windows)
- [ ] Move binary to `$GOPATH/bin/` or `/usr/local/bin/`
- [ ] Set permissions `0755`
- [ ] Clean up temp files

## Acceptance

```
✅ app — v1.0.0
```
