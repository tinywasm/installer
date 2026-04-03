# Stage 9 — Tests (dependency injection + mocks)

## Goal

Tests that cover every path in the install flow diagram without downloading or
installing anything real. All external operations are injected as interfaces.

## Injected dependencies

```go
// deps.go — all external operations behind interfaces
type Deps struct {
    RunCmd      func(name string, args ...string) ([]byte, error) // exec.Command
    Download    func(url string) ([]byte, error)                  // net/http GET
    WriteFile   func(path string, data []byte, perm os.FileMode) error
    RemoveFile  func(path string) error
    LookPath    func(name string) (string, error)                 // exec.LookPath
    Checklist   func(tools []Tool) []int                          // interactive checklist
}
```

Each function in the codebase receives `Deps` instead of calling globals directly.
In production `cmd/installer/main.go` passes real implementations.
In tests, mocks are injected.

## Test cases mapped to diagram

### GoInstall mode

| Test | Diagram path | Mock behavior |
|------|-------------|---------------|
| `TestGoInstall_Success` | GoInstall → Verify → ✅ | `RunCmd("go","install",...)` returns nil, verify returns version |
| `TestGoInstall_Fail_Required` | GoInstall → Verify fail → ❌ FATAL | `RunCmd` returns error, tool is Required → expects fatal |
| `TestGoInstall_Fail_Optional` | GoInstall → Verify fail → ❌ warning | `RunCmd` returns error, tool is Optional → expects continue |

### Binary mode

| Test | Diagram path | Mock behavior |
|------|-------------|---------------|
| `TestBinary_Success` | Binary → Verify → ✅ | `Download` returns fake tarball, verify returns version |
| `TestBinary_DownloadFail` | Binary → ❌ | `Download` returns error |
| `TestBinary_VerifyFail` | Binary → Verify fail → ❌ | `Download` ok, verify returns wrong version |

### Checklist

| Test | Diagram path | Mock behavior |
|------|-------------|---------------|
| `TestChecklist_SelectSome` | Checklist → select 2 of 3 → install selected | `Checklist` returns `[0, 2]` → tools 0 and 2 install, tool 1 skipped |
| `TestChecklist_SelectNone` | Checklist → select none → skip all optional | `Checklist` returns `[]` → only required tools install |
| `TestChecklist_NoOptional` | No optional tools → skip checklist | No optional tools in registry → `Checklist` never called |

### Required vs Optional

| Test | Diagram path | Mock behavior |
|------|-------------|---------------|
| `TestRequired_StopsOnFail` | Required → fail → exit | First required tool fails → no more tools run |
| `TestRequired_AlwaysInstalled` | Required → no checklist | Required tools install without appearing in checklist |
| `TestOptional_ContinuesOnFail` | Optional selected → fail → next tool | Selected optional fails → next tool still runs |

### Dependencies

| Test | Diagram path | Mock behavior |
|------|-------------|---------------|
| `TestDep_SkipOnFail` | git fails → gh auto-skipped | `git` install fails → `gh` (DependsOn: "git") skipped with reason |
| `TestDep_SkipOnSkip` | git unchecked → gh auto-skipped | `git` not selected in checklist → `gh` also skipped |
| `TestDep_Success` | git ok → gh installs | `git` succeeds → `gh` installs normally |

### Uninstall

| Test | Diagram path | Mock behavior |
|------|-------------|---------------|
| `TestUninstall_Specific` | UNINSTALL=tinygo → remove one | `LookPath` returns path, `RemoveFile` called once |
| `TestUninstall_All` | UNINSTALL=all → remove all reverse | `RemoveFile` called for each tool in reverse order |
| `TestUninstall_NotFound` | tool not installed → warning | `LookPath` returns error → skip |

### Verify

| Test | Diagram path | Mock behavior |
|------|-------------|---------------|
| `TestVerify_VersionFlag` | tool version → ok | `RunCmd("tool","version")` returns matching version |
| `TestVerify_DashDashVersion` | tool --version → ok | `RunCmd("tool","version")` fails, `RunCmd("tool","--version")` returns match |
| `TestVerify_Mismatch` | version != expected → error | `RunCmd` returns different version string |

### Summary

| Test | Mock behavior |
|------|---------------|
| `TestSummary_AllSuccess` | 3 tools, all succeed → "3 installed, 0 failed, 0 skipped" |
| `TestSummary_Mixed` | 1 ok, 1 fail (optional), 1 skipped → "1 installed, 1 failed, 1 skipped" |
| `TestSummary_DepSkip` | git fails, gh dep-skipped → "0 installed, 1 failed, 1 skipped (requires git)" |

## Acceptance

`go test ./...` passes with zero network calls, zero disk writes, zero real installs.
