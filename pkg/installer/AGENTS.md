# INSTALLER KNOWLEDGE BASE

## OVERVIEW

`pkg/installer` defines shared installer-facing contracts used by backend/application services and installer tooling.

This package should stay lightweight: command constants, file-name conventions, installer status enums, and plugin script templates.

## STRUCTURE

```
pkg/installer/
|- constant.go   # installer CLI commands, log/file naming, offline package layout conventions
|- state.go      # installer runtime state enum helpers
|- plugin.go     # plugin install/reload/restart script rendering and placeholders
`- README.md     # minimal package introduction
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add/adjust installer command tokens | `constant.go` | Keep values compatible with `tools/cmd/installer/**` CLI definitions |
| Add offline bundle naming/layout conventions | `constant.go` | Shared by `internal/application/router/.../workflow/offline.go` and `internal/backend/router/.../workflow/offline.go` |
| Add installer state mapping | `state.go` | Keep status semantics backward-compatible for workflow wait/report actions |
| Adjust plugin shell templates | `plugin.go` | Validate placeholder replacement and OS-specific script behavior |

## CONVENTIONS

- Keep constants as the single source of truth when names/paths are consumed by multiple services.
- Prefer adding new exported constants instead of scattering string literals in router/workflow code.
- Public constants and exported helpers must have clear English comments.
- When introducing a new cross-layer convention, add SYNC comments that point to upstream/downstream consumers.

## SYNC CONTRACTS

- `DataFileName` / `StatusFileName` must stay aligned with files written by `tools/internal/installer/node/*reporter`.
- Offline package constants (for example `OfflinePkgRelPathData`, `OfflinePkgInstallScriptName`) must match:
  - tar entry generation in application offline package download handler
  - `install.sh` generation logic in backend offline install info handler
  - related frontend/offline guide text where applicable
- Log format constants (`LogFieldSeparator`, `LogFieldCount`) must stay compatible with backend log parsers.

## ANTI-PATTERNS

- Do not place service-specific business logic in this package.
- Do not hardcode offline bundle path/file literals in multiple routers when a shared constant can be used.
- Do not add heavy dependencies here; keep this package import-safe for broad reuse.
- Do not silently change existing constant values that are part of tool/protocol compatibility.
