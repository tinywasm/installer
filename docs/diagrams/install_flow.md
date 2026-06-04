# Install Flow

```mermaid
flowchart TD
    subgraph PHASE1["Phase 1 — bash / ps1"]
        A[install.sh / install.ps1] --> B[Detect OS/Arch]
        B --> C[Download raw binary\ntinywasm-installer-{os}-{arch}]
        C --> D[chmod +x]
        D --> E[Run: installer]
    end

    subgraph PHASE2["Phase 2 — Go binary"]
        E --> F[Load embedded\ngo_version.conf]
        F --> G{UNINSTALL\nenv set?}

        G -->|yes| UN[Uninstall flow]
        UN --> UN1{UNINSTALL=all?}
        UN1 -->|yes| UN_ALL[Remove each tool\nin reverse order]
        UN1 -->|no| UN_ONE[Remove specific tool]
        UN_ALL --> DONE
        UN_ONE --> DONE

        G -->|no| CHK[Show checklist\nof optional tools]
        CHK --> SEL["User selects with\n↑/↓ Space Enter"]
        SEL --> H[Install all\nrequired + selected\ngoroutines + spinner]

        H --> LOOP[For each tool]

        LOOP --> DEP{DependsOn\nfailed/skipped?}
        DEP -->|yes| DEP_SKIP["⏭ tool — skipped\n(dependency failed)"]
        DEP_SKIP --> NEXT
        DEP -->|no| MODE{mode?}
        MODE -->|GoInstall| I[go install\nsource@version]
        MODE -->|Binary| BIN[Download release\nfrom GitHub]

        I --> V[Verify: tool version]
        BIN --> V

        V -->|ok| OK["✅ tool — version"]
        V -->|fail + required| FAIL_REQ["❌ tool — FATAL"]
        V -->|fail + optional| FAIL_OPT["❌ tool — warning, continue"]

        OK --> NEXT{More tools?}
        FAIL_OPT --> NEXT
        FAIL_REQ --> ERR2[Exit 1]
        NEXT -->|yes| LOOP
        NEXT -->|no| DONE[Summary]
    end
```

## Terminal output

```
Installing Go 1.25.2...
Go 1.25.2 installed successfully.

Select optional tools to install:
  [x] git — 2.47.0
  [x]   └─ gh — 2.65.0  (requires git)
  [ ] lazygit — 0.44.1
  [x] wasmtime — 25.0.0

  ↑/↓: move  Space: toggle  Enter: confirm

⠋ Installing tinygo...
✅ tinygo — 0.40.1
⠋ Installing git...
✅ git — 2.47.0
⠋ Installing gh...
✅ gh — 2.65.0
⠋ Installing wasmtime...
✅ wasmtime — 25.0.0
⏭ lazygit — skipped

Done. 4 installed, 0 failed, 1 skipped.
```

## Two phases

| Phase | Runtime | Responsibility |
|-------|---------|---------------|
| 1 | bash / ps1 | Bootstrap: detect OS/Arch, download `tinywasm-installer` binary, execute it. |
| 2 | Go binary | Everything: install Go (if needed), checklist, spinner, GoInstall, Binary, checksum verify, uninstall |

## Required vs Optional

| Type | On failure | Selection |
|------|-----------|-----------|
| Required | Fatal — stop installer | Always installed, no prompt |
| Optional | Warning — continue with next tool | Shown in checklist, user toggles before install starts |

## Checklist interaction

The checklist uses `golang.org/x/term` for raw terminal input:

| Key | Action |
|-----|--------|
| `↑` / `↓` | Move cursor between optional tools |
| `Space` | Toggle `[x]` / `[ ]` |
| `Enter` | Confirm selection, start installation |

All optional tools default to `[ ]` (unchecked). Required tools are not shown
in the checklist — they always install.

## Tool dependencies

Tools can declare `DependsOn: "other_tool"`. Before installing a tool, the
installer checks if its dependency succeeded. If the dependency failed or was
skipped, the dependent tool is auto-skipped.

```
⠋ Installing git...
❌ git — install failed (warning, continue)
⏭ gh — skipped (requires git)
```

Dependency rules:
- A tool can depend on at most one other tool (single `DependsOn` string).
- Dependencies must appear before their dependents in the `tools` slice.
- If a dependency was skipped (unchecked in checklist), dependents are also skipped.
- If a dependency failed, dependents are auto-skipped with a reason message.
