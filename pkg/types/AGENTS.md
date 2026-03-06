# TYPES KNOWLEDGE BASE

## OVERVIEW

`pkg/types` is the single source of truth for shared domain data structures across all services. It defines business-layer payloads (structs, enums, conditions, distinct selectors, manager params) that flow between API handlers, storage, workflow engine, and third-party adapters. Proto structs must be converted to/from these types at the boundary — see `pkg/proto/*/README.md`.

## WHERE TO LOOK

| File | Purpose |
|------|---------|
| `host.go` | `Host`, `HostStatic`, `HostDynamic` — core host model with addressing, status, role, proxy tags |
| `topo.go` | `Business`, `NetworkArea`, `NetworkUnit`, `AccessPoint`, `Links` — topology primitives |
| `scope.go` | `Scope`, `Target` — scope types (topo/service-template/set-template/instance/dynamic-group) for targeting hosts |
| `release.go` | `Release` and variants (`ReleaseAgent`, `ReleaseProxy`, `ReleaseCert`, `ReleasePlugin`, etc.) — package release metadata |
| `plugin.go` | `Plugin` struct and plugin group constants |
| `config_policy.go` | `ConfigPolicy`, `ConfigPolicyScope`, `ConfigPolicyTemplate` — configuration policy model |
| `deploy_policy.go` | `DeployPolicy`, `DeploySpec` — deployment policy with agent/proxy/plugin spec variants |
| `node_deployment.go` | `NodeDeployment`, `DeploymentInfo`, `NodeConf`, `TargetVersion` — per-node deployment state |
| `plugin_deployment.go` | `PluginDeployment`, `PluginDeploymentInfo`, `PluginConfigDetail` — per-plugin deployment state |
| `node_workflow.go` | `NodeWorkflow` — node operation workflow with type/status enums |
| `plugin_workflow.go` | `PluginWorkflow` — plugin operation workflow with type/status enums |
| `scheduled_workflow.go` | `ScheduledWorkflow` — scheduled/recurring workflow definitions |
| `node_agent.go` | Install/upgrade/restart/reconfig/uninstall param structs for agent operations |
| `node_proxy.go` | Install/upgrade/restart/reconfig/update/uninstall param structs for proxy operations |
| `manager.go` | High-level manager action params (`InstallNodeParam`, `UpgradePluginParam`, `ExecuteDeployPolicyParam`, etc.) |
| `gse.go` | GSE integration types: `AgentState`, `ScriptResult`, `TransferDetail`, `OperateAgent`, `GSEVersionFormatter` |
| `process.go` | `Process`, `ProcessSpec`, `ProcessController`, `ProcessConfig` — process management model |
| `network_policy.go` | `NetworkPolicy`, `NetworkPolicyEndpoint`, `NetworkPolicyService` — network access policy |
| `condition.go` | Query condition structs (`*ExactFields`, `*FuzzyFields`, `*Condition`) for every domain entity |
| `distinct.go` | Distinct/aggregation request and result types (`*DistinctRequest`, `*DistinctResult`) |
| `page.go` | `Page` struct with sort/limit/offset and helpers (`UnlimitedPage`, `SingleItemPage`) |
| `constant.go` | Release version/name constants (`ReleaseNameAgent`, `ReleaseNameProxy`, etc.) |
| `generation.go` | `Generation` enum (All/1/2) for agent generation selection |
| `cipher.go` | `Cipher` struct and `CipherKeyType` enum (AES256, RSA, SM2, SM4) |
| `upload.go` | `Upload` and origin package detail structs for file upload categorization |
| `tenant.go` | `Tenant` struct for multi-tenant context |
| `relay.go` | `RelayInfo` struct for relay service metadata |
| `host_event.go` | `HostEvent` — CMDB host change event model |
| `topo_event.go` | `TopoEvent` — topology change event model |
| `config_policy_event.go` | `ConfigPolicyEvent` — config policy change event model |
| `package_event.go` | `PackageEvent` — package lifecycle event model |
| `operinst_private_data.go` | Operation instance private data keys and `PDDetectInfo` struct |
| `graph.go` | `GraphNodeInfo` — workflow graph visualization node |
| `service_instance.go` | `ServiceInstance` — CMDB service instance reference |

## CONVENTIONS

### Enum Pattern

Domain enums are defined as named `string` (or `int64`) types with `const` blocks and always include:
1. A `Validate()` method that checks against valid values
2. `*ListToStringList` / `StringListTo*List` conversion helpers for API serialization

```go
type NodeStatus string
const (
    NodeStatusRunning NodeStatus = "running"
    // ...
)
func (s NodeStatus) Validate() error { ... }
func NodeStatusListToStringList(list []NodeStatus) []string { ... }
func StringListToNodeStatusList(list []string) []NodeStatus { ... }
```

### Scope Union Type

`Scope` uses a union pattern — exactly one inner scope variant is non-nil. Use `NewScopeWith*` constructors and `scope.Type()` / `scope.Get*Scope()` accessors; never set fields directly.

### Condition Structs

Every queryable entity has a triple of condition structs in `condition.go`:
- `*ExactFields` — exact-match filter fields
- `*FuzzyFields` — fuzzy/regex filter fields
- `*Condition` — combines exact, fuzzy, `TimeRange`, and `Page`

These map directly to DAO `OptFn` filter chains in `pkg/dao/mongo`.

### Distinct Selectors

`distinct.go` pairs each domain entity with `*DistinctRequest` (bool fields selecting which columns to aggregate) and `*DistinctResult` (aggregation output). Use `New*AllSet()` constructors for full-column aggregation.

### Manager Params

`manager.go` contains high-level action parameter structs consumed by `internal/*/manager`. These are the entry point for service-level operations (install/upgrade/restart/uninstall nodes and plugins, execute deploy policies).

### DeploySpec Union Type

`DeploySpec` in `deploy_policy.go` follows the same union pattern as `Scope` — use `NewDeploySpecWith*` constructors and `deploySpec.Type()` / `deploySpec.Get*Param()` accessors.

## ANTI-PATTERNS

- Do not add service-specific business logic to type files — keep validation limited to structural/value correctness.
- Do not use proto-generated structs as business-layer types — convert at the boundary via `pkg/proto/*` helpers.
- Do not expose raw third-party API structures (CMDB, GSE, etc.) — adapter types in `gse.go`, `cmdb.go` are the internal projection; raw API shapes live in `pkg/thirdparty/`.
- Do not duplicate condition or distinct structs — follow the existing `*ExactFields` / `*FuzzyFields` / `*Condition` triple pattern.
- Do not construct `Scope` or `DeploySpec` by setting fields directly — always use the `NewScopeWith*` / `NewDeploySpecWith*` constructors.
