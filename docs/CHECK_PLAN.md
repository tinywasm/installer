# PLAN — instalación única de tinywasm/app vía el installer

Master orchestrator para que exista **una sola forma de instalar `tinywasm/app`**: el
`installer` descarga el binario publicado por `gorelease` en
`github.com/tinywasm/app/releases`. Se elimina el método manual suelto (curl ad-hoc) como
camino alternativo; queda documentado solo como fallback de bajo nivel que usa el **mismo**
asset que el installer.

> Documento HOW. El contrato de release (qué assets produce `gorelease`) vive en
> [devflow/docs/PLAN.md](../../devflow/docs/PLAN.md) y
> [devflow/docs/diagrams/GORELEASE_FLOW.md](../../devflow/docs/diagrams/GORELEASE_FLOW.md).

---

## Development Rules (constraints copiadas para este PLAN)

- **Documentation First**: este PLAN se actualiza ANTES de tocar código.
- **TDD / Red first**: en esta etapa SOLO se escriben tests que reproducen el flujo
  esperado. NO se toca código de producción. Los tests definen el contrato.
- **Dependency Injection**: toda I/O pasa por `Deps` (`Download`, `RunCmd`, `LookPath`…).
  Los tests inyectan `Deps` y NO tocan red ni disco real.
- **SRP**: descarga (`InstallBinary`), verificación (`Verify`) y catálogo (`Tools`) son
  unidades separadas. No mezclar.
- **Single source of truth**: el formato del asset lo define `gorelease` (devflow). El
  installer DEBE consumir ese formato, no inventar otro. Si cambia uno, cambia el contrato
  compartido y ambos lados se actualizan juntos.

---

## Problema

Tras separar fuente privada (`tinywasm/core`) de distribución pública (`tinywasm/app`):

1. **Contrato de asset incompatible** entre productor y consumidor:

   | | Productor `gorelease` | Consumidor `installer` (hoy) |
   |---|---|---|
   | Nombre | `tinywasm-{os}-{arch}` | `{name}_v{ver}_{os}_{arch}` |
   | Formato | binario crudo | `.tar.gz` / `.zip` |
   | Versión en nombre | no | sí |
   | Ruta release | `…/releases/download/v{ver}/` | `…/releases/download/v{ver}/` |

2. **`Tools` con placeholders**: `tinywasm-cli`/`tinywasm-server` → `tinywasm/tinywasm`
   eran ejemplos provisionales (no se sabía dónde quedaría el binario final). El binario
   real es `tinywasm` desde `tinywasm/app` (`cmd/tinywasm`).

3. **Dos formas de instalar**: el README de `tinywasm/app` documenta un `curl` manual
   suelto, en paralelo al installer. El usuario quiere **una sola**.

4. **Bootstrap depende de Go**: hoy `install.sh`/`install.ps1` instalan Go y luego hacen
   `go install tinywasm/installer`. El installer debe distribuirse **como binario**
   (compilado por `gorelease`), para correr sin Go previo.

## Solución

**El contrato canónico es el de `gorelease`: binario crudo `tinywasm-{os}-{arch}[.exe]`.**
El installer se adapta a ese formato (no al revés): es más simple (sin archivar/extraer) y
coincide con el `curl` de bajo nivel.

```
URL = {Source}/releases/download/v{Version}/tinywasm-{GOOS}-{GOARCH}[.exe]
       └─ Source = https://github.com/tinywasm/app
```

Descarga directa → `chmod +x` → mover a `$GOPATH/bin` (lógica de `getInstallPath` ya existe).

---

## Diseño de implementación (HOW — para el agente de ejecución)

### 0. El installer se distribuye como binario (bootstrap sin Go)
`gorelease` compila la carpeta `cmd/` → asset crudo `{cmdDir}-{os}-{arch}[.exe]`. Para que
el asset se llame **`tinywasm-installer-{os}-{arch}`** hay que **renombrar la carpeta**
`cmd/installer` → `cmd/tinywasm-installer` (gorelease deriva el nombre del directorio).
El asset se publica en `tinywasm/installer/releases` (`CGO_ENABLED=0`, estático).

El bootstrap deja de instalar Go y de hacer `go install`:

```
ANTES  curl|bash → instala Go → go install tinywasm/installer → ejecuta
AHORA  curl|bash → descarga tinywasm-installer-{os}-{arch} de releases → chmod +x → ejecuta
```

`install.sh` / `install.ps1` solo detectan `{os}-{arch}`, descargan el binario
(`…/releases/latest/download/tinywasm-installer-{os}-{arch}[.exe]`) y lo ejecutan. El binario
installer (Phase 2) sigue instalando Go + TinyGo + `tinywasm` + tools para el entorno de
desarrollo del usuario. Resultado: **no se requiere Go para arrancar el installer**.

Validación: integración/manual (shell), no unit test Go. Ver [TEST_WINDOWS.md](TEST_WINDOWS.md).

### 1. Catálogo `Tools` — consolidar a un binario
Reemplazar las entradas `tinywasm-cli`/`tinywasm-server` por una sola, **sin versión quemada**:
```go
{Mode: Binary, Name: "tinywasm", Source: "https://github.com/tinywasm/app", Version: "", Required: true},
```
`Version: ""` ⇒ resolver siempre la última release en runtime (ver 1b). `Required: true`:
el installer existe para dejar `tinywasm` operativo; es el tool central. (Fácil de invertir.)

### 1b. Resolución de versión — NO hardcodear
**Problema:** con `Version` quemado, publicar una versión nueva en `tinywasm/app` obligaría a
actualizar y republicar el installer. Sin sentido. El installer debe **consultar** la última
release publicada.

**Diseño (cada mecanismo en su contexto):**

| Contexto | Mecanismo | Por qué |
|----------|-----------|---------|
| Installer (Go) | API `GET https://api.github.com/repos/{owner}/{repo}/releases/latest` → `tag_name` | obtiene la versión para descargar **y** para `Verify` (coherente con la versión embebida, devflow #2) |
| Bootstrap (sh/ps1) | redirect `…/releases/latest/download/<asset>` | shell simple, sin parsear JSON (sin `jq`) |

Flujo en `InstallBinary`:
1. `tag := resolveLatestVersion(t.Source, d)` — deriva `owner/repo` de `Source`, descarga la
   API `releases/latest` (vía `d.Download`), parsea `tag_name`.
2. Construir la URL del asset con ese `tag` (binario crudo, sección 2).
3. `Verify` compara la versión auto-reportada del binario contra `tag` (requiere devflow #2).

Siempre la última release. No se hardcodea ninguna versión.

Test (live): `TestInstallBinary_ResolvesLatestVersion` 🔴 — la URL del asset debe usar el
`tag_name` resuelto por la API, no un valor quemado.

### 2. `InstallBinary` — modo binario crudo + verificación de integridad
Cambiar la construcción de URL, **verificar checksum** y eliminar la extracción de archivo:
```go
asset := fmt.Sprintf("tinywasm-%s-%s", runtime.GOOS, runtime.GOARCH)
if runtime.GOOS == "windows" { asset += ".exe" }
base := fmt.Sprintf("%s/releases/download/v%s", t.Source, t.Version)
data, err := d.Download(base + "/" + asset)
// #1 SEGURIDAD: descargar checksums.txt, buscar la línea de `asset`,
// comparar sha256(data) ANTES de escribir. Si no coincide → error, NO instalar.
sums, err := d.Download(base + "/checksums.txt")
if err := verifyChecksum(asset, data, sums); err != nil { return err }
// escribir data directo a getInstallPath(t.Name) con 0755 (sin tar/zip)
```
`extractTarGz`/`extractZip` quedan obsoletos para este flujo (eliminar o aislar si algún
otro tool los necesitara; hoy no hay otro Binary tool).

**#1 — Verificación de checksum (SEGURIDAD).** Se descarga un ejecutable y se hace
`chmod +x`; sin verificar, un release comprometido o un MITM = ejecución de código
arbitrario. `gorelease` publica `checksums.txt` (ver devflow PLAN #1); el installer compara
el SHA256 **antes** de instalar. Helper puro testeable:
```go
// test de referencia (se activa al crear verifyChecksum):
func TestVerifyChecksum_RejectsMismatch(t *testing.T) {
    data := []byte("binary-bytes")
    sums := "deadbeef...  tinywasm-linux-amd64\n" // sha incorrecto
    if err := verifyChecksum("tinywasm-linux-amd64", data, []byte(sums)); err == nil {
        t.Fatal("expected error on checksum mismatch")
    }
}
```

**#5 — Timeout de descarga (FIABILIDAD).** [mode_binary.go:155] usa `http.Get` sin timeout:
una conexión colgada bloquea el installer sin feedback. Cambiar a
`(&http.Client{Timeout: 60*time.Second}).Get(url)`. Validación: revisión de código (no
unit test trivial sin servidor de prueba).

### 3. Cobertura de plataformas — resuelto
`gorelease` publica (mejora C, ya en el PLAN de devflow): linux/{amd64,arm64},
darwin/{arm64,amd64}, windows/amd64. Cubre las plataformas que el installer detecta
(`runtime.GOOS/GOARCH`), así que no hay 404. Si en el futuro corre en una plataforma fuera
de la matriz, `InstallBinary` debe mostrar "plataforma no soportada" en vez de un 404 crudo.

### 4. README de `tinywasm/app` — una sola vía
Sustituir el bloque `curl` manual por: "Instala con el tinywasm installer" apuntando a
`tinywasm/installer`. El `curl` directo se conserva solo como nota de fallback avanzado,
descargando el **mismo** asset `tinywasm-{os}-{arch}`.

---

## Estrategia de tests (lo que se escribe AHORA)

Archivo: `app_binary_test.go` (package `installer`, interno — accede a `Tools` y métodos).
Toda I/O se mockea con `Deps`.

| Test | Escenario | Aserción | Estado HOY |
|------|-----------|----------|------------|
| `TestTools_TinywasmFromAppRepo` | catálogo | existe tool `tinywasm`, `Mode==Binary`, `Source=="https://github.com/tinywasm/app"` | 🔴 RED |
| `TestInstallBinary_DownloadsRawGoreleaseAsset` | formato del asset | `Download` recibe `…/v{tag}/tinywasm-{os}-{arch}[.exe]` (crudo, sin `.tar.gz`) | 🔴 RED |
| `TestInstallBinary_ResolvesLatestVersion` | última release | la URL usa el `tag_name` resuelto por la API `releases/latest` (sin versión quemada) | 🔴 RED |
| `TestVerifyChecksum_RejectsMismatch` (referencia) | integridad | `verifyChecksum` rechaza sha que no coincide (#1) | ⏸ requiere helper |
| `TestBinary_DownloadFail` (existente) | error de red | `InstallBinary` propaga error | 🟢 GREEN (no cambia) |

Los 🔴 RED pasan a GREEN tras implementar el diseño. No deben requerirse cambios en los
tests existentes de `InstallBinary`/`Verify`/`ResolveTools`.

---

## Checklist de ejecución

- [x] Crear este `PLAN.md`
- [x] Escribir tests RED en `app_binary_test.go`
- [ ] Renombrar `cmd/installer` → `cmd/tinywasm-installer` (asset `tinywasm-installer-{os}-{arch}`)
- [ ] **Bootstrap**: `install.sh`/`install.ps1` descargan `tinywasm-installer-{os}-{arch}` (sin Go, sin `go install`)
- [ ] Publicar el installer vía `gorelease` (release en `tinywasm/installer`)
- [ ] Consolidar `Tools` a un único `tinywasm` desde `tinywasm/app` (`Required: true`, `Version: ""`)
- [ ] **Resolución de versión**: `resolveLatestVersion` (API `releases/latest`) — siempre la última, sin hardcodear
- [ ] `InstallBinary`: modo binario crudo (URL `gorelease` + sin extracción)
- [ ] **#1** Verificar `checksums.txt` (SHA256) antes de `chmod +x`; activar `TestVerifyChecksum_RejectsMismatch`
- [ ] **#5** `DefaultDownload`: `http.Client{Timeout}` en vez de `http.Get`
- [x] Cobertura darwin/amd64: resuelto — `gorelease` añade el target (devflow mejora C)
- [ ] Actualizar README de `tinywasm/app`: instalación única vía installer
- [ ] Actualizar `docs/diagrams/install_flow.md` (bootstrap binario + fuente tinywasm/app)
- [ ] Actualizar README del installer (Quick start, "How it works", tabla Tools)
- [ ] `gotest` verde

## Related

- [README](../README.md) — uso del installer
- [docs/diagrams/install_flow.md](diagrams/install_flow.md) — flujo de instalación
- [devflow/docs/PLAN.md](../../devflow/docs/PLAN.md) — contrato de release (gorelease)
