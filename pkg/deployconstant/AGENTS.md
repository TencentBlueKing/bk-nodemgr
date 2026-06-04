# DEPLOY CONSTANT KNOWLEDGE BASE

## OVERVIEW

`pkg/deployconstant` defines shared deployment-facing constants, default values, and path derivation helpers used by node/plugin deployment flows.

This package should stay lightweight: generation + OS keyed deploy config registration, stable directory or IPC naming, and reusable default path rendering.

## STRUCTURE

```
pkg/deployconstant/
|- deploy_constnat.go        # base DeployConf, validation, generation+OS registry, deploy/work dir generation
|- node_deploy_constant.go   # node deploy defaults, log/config dir defaults, IPC path/port helpers
|- plugin_deploy_constant.go # plugin deploy defaults, plugin dir helpers, hostid path, common constants accessors
`- README.md                 # minimal package introduction
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add or adjust base deployment config fields | `deploy_constnat.go` | Keep `DeployConf.Validate()` and generation/OS registry semantics stable |
| Add node deployment default paths or IPC defaults | `node_deploy_constant.go` | Reuse existing `Generate*` helpers instead of scattering path joins |
| Add plugin deployment default paths or constants | `plugin_deploy_constant.go` | Keep plugin home/config/run/data dir derivation consistent |
| Add shared deploy literals reused by multiple flows | `node_deploy_constant.go`, `plugin_deploy_constant.go` | Prefer package-level constants over repeated string literals in callers |

## CONVENTIONS

- Keep this package limited to constants, defaults, and path derivation; deployment orchestration stays in `internal/*` layers.
- Prefer extending existing `DeployConf` / `NodeDeployConf` / `PluginDeployConf` structs and `Generate*` helpers instead of adding parallel helper packages.
- Validate config before registration and keep the “set once per `Generation` + `OSType`” pattern intact.
- Public constants and exported helpers must have clear English comments.
- Derive OS-aware paths through `pkg/format/tool` and `pkg/system.GetEnv()` instead of duplicating path-format logic in callers.

## SYNC CONTRACTS

- `DeployConf`, `NodeDeployConf`, and `PluginDeployConf` lookup semantics must stay compatible with callers that read deploy defaults by `types.Generation` and `criteria.OSType`.
- Default path formats (for example log dir, extra config dir, hostid path, plugin home/config/run/data dirs) must stay aligned with the actual installer/runtime file layout expected by deployment workflows.
- Windows IPC port defaults and Unix IPC path naming must remain stable for components that connect through generated IPC endpoints.
- Shared deploy literals should be updated here first when the same value is consumed by multiple deployment flows.

## ANTI-PATTERNS

- Do not place service-specific deployment business logic in this package.
- Do not put user custom rendering or request-derived configuration assembly in this package.
- Do not duplicate deploy path literals in callers when an existing `Generate*` helper already defines the convention.
- Do not add DAO/API/proto/workflow dependencies here; keep this package import-safe for broad reuse.
- Do not silently change existing default values or path formats that act as cross-module compatibility contracts.
