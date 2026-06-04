# PLAN — `tinywasm/installer`: rewire a `tinywasm/update` (código ya movido)

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Zero-context agent: todo lo necesario está aquí.

`verifyChecksum`, `resolveLatestVersion` y `DefaultDownload` **ya fueron MOVIDOS** desde este
módulo a `github.com/tinywasm/update`. El installer **no compila** hasta reconectar las
llamadas. **No reimplementes esas funciones; llama a las de `update`.**

Puntos rotos a reconectar (ya marcados con NOTE en el código):
- `mode_binary.go` → `InstallBinary` llama a `resolveLatestVersion` y `verifyChecksum` (inexistentes).
- `cmd/tinywasm-installer/main.go` → usa `installer.DefaultDownload` (inexistente).

---

## Development Rules (MANDATORY)

- Comentarios y docs en **inglés**.
- **Sin duplicación**: no recrear las funciones movidas; importar `update`.
- Sin nuevas dependencias externas más allá de `github.com/tinywasm/update`.
- Comportamiento y URLs **idénticos** a antes (los tests `TestInstallBinary_*` son el contrato).

---

## Dependencia

`go get github.com/tinywasm/update && go mod tidy`. API usada:
```go
update.ResolveLatestVersion(source string, download func(string)([]byte,error)) (string, error)
update.VerifyChecksum(asset string, data, sums []byte) error
update.DefaultDownload(url string) ([]byte, error)
```
**Dato clave**: `Deps.Download` tiene firma `func(string)([]byte,error)` — se pasa directo como
el `download` de `ResolveLatestVersion`, sin wrapper.

## Stage 1 — `mode_binary.go`

En `InstallBinary`, reconecta las dos llamadas (todo lo demás del flujo se mantiene):
```go
version := t.Version
if version == "" {
    v, err := update.ResolveLatestVersion(t.Source, d.Download) // antes: resolveLatestVersion(t.Source, d)
    if err != nil {
        return fmt.Errorf("failed to resolve latest version: %w", err)
    }
    version = strings.TrimPrefix(v, "v")
}
// ...asset, base, url, d.Download(...) sin cambios...
if err := update.VerifyChecksum(asset, data, sums); err != nil { // antes: verifyChecksum(...)
    return err
}
```
Añade `"github.com/tinywasm/update"` al import. La cabecera NOTE del refactor puede borrarse una
vez reconectado.

## Stage 2 — `cmd/tinywasm-installer/main.go`

```go
Download: update.DefaultDownload, // antes: installer.DefaultDownload
```
Añade el import `"github.com/tinywasm/update"` (y quita el de installer si queda sin uso).

## Stage 3 — Tests / docs

- `TestVerifyChecksum_RejectsMismatch` ya fue eliminado (su contrato vive en
  `update.TestVerifyChecksum`). Los tests `TestTools_TinywasmFromAppRepo`,
  `TestInstallBinary_DownloadsRawGoreleaseAsset`, `TestInstallBinary_ResolvesLatestVersion`
  deben quedar **verdes sin cambios** (las URLs no cambian).
- `README.md`/docs de instalación: notar que la resolución de versión, el checksum y la descarga
  los provee `github.com/tinywasm/update`.

---

## Stages table

| Stage | Output | Done when |
|------|--------|-----------|
| 1 | `mode_binary.go` | llama a `update.ResolveLatestVersion`/`update.VerifyChecksum`; import añadido; compila |
| 2 | `cmd/tinywasm-installer/main.go` | usa `update.DefaultDownload` |
| 3 | tests/docs | `go test ./...` verde; docs notan la dependencia |

## Acceptance criteria

- `go test ./...` verde; `go vet ./...` limpio.
- No queda lógica de SHA256 / `releases/latest` / downloader propia en el installer — solo
  llamadas a `update.*`.
- URLs y comportamiento byte-idénticos (lo prueban los tests existentes).
