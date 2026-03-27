# NODE-DEPLOYMENT DATA ACCESS KNOWLEDGE BASE

## OVERVIEW

`pkg/dao/mongo/node-deployment` (`package nodedeployment`) is the MongoDB DAO for **node deployment** artifacts: persisted install/upgrade/reconfig context keyed by **token**.

Documents use the standard `base.TableBroker[*Data]` envelope (`basic` + `data`). The public API converts to/from `pkg/types` (`NodeDeployment`, `DeploymentInfo`, `NodeConf`).

Design boundaries and TTL intent: `README.md`.

## WHERE TO LOOK

| Task | Location |
|------|----------|
| Collection name, ORM wiring, TTL index | `nodedeploy.go` |
| BSON/document model (`Data`, `Info`, `NodeConf`, unique key) | `table.go` |
| Filter field paths for queries | `consts.go` |
| Query option builders (`With*`) | `options.go` |
| `IHandler` / `Handler` CRUD + partial field reads/updates | `handler.go` |
| Shared DTOs for callers | `pkg/types` (search `NodeDeployment`, `DeploymentInfo`, `NodeConf`) |
| Storage orchestration | `internal/backend/storage/node` (`dao_node_deloyment.go`, `storage.go`) |
| Mongo ORM / envelope rules | `pkg/dao/mongo/base` (`AGENTS.md`) |

## TABLE MODEL

### Collection

- **Name**: fixed `node_deployment` (not tenant-suffixed; unlike `node_workflow_<tenant>`).
- **Index**: TTL on `data.expire_at` with `ExpireTimeSec` — see `GetIndexes()` in `nodedeploy.go`. Align persisted documents with that path if TTL must apply.

### Core schema

- `Table` is `base.TableBroker[*Data]`.
- `Data` implements `base.IData`; **unique key** is `Token` (`UniqueFields`: `data.token`).
- Inner payload: `Token`, `Info`, `NodeConf` (see `table.go` for nested fields).

### Filter field keys (`consts.go`)

- `data.token`, `data.info`, `data.node_conf`
- `data.info.inner_ip_list`, `data.info.inner_ip_list_v6`, `data.info.biz_id`, `data.info.node_version`
- `data.info.networkarea_id`, `data.info.networkunit_id`

## QUERY AND HANDLER API

### Options (`options.go`)

- `WithToken`
- `WithInfoInnerIP` / `WithInfoInnerIPV6`
- `WithInfoBizID` / `WithInfoNetworkAreaID` / `WithInfoNetworkUnitID`
- `WithInfoNodeVersion`

### Handler (`handler.go`)

- `New(client *mongo.Database)` — ensures indexes (warns on failure).
- `CreateNodeDeployment`, `ListNodeDeployment` (with `types.Page` + opts)
- `GetNodeDeploymentInfo` (projection on `FieldKeyInfo`)
- `GetNodeDeploymentNodeConf` / `SetNodeDeploymentNodeConf` (partial `node_conf`)
- `UpdateNodeDeploymentInfo` (partial `info`)

## SCHEMA CONVENTIONS

- Keep `json`/`bson` tags aligned with `handler` conversions and storage callers.
- This package owns persistence shape and field paths; **avoid** workflow/business rules here.
- Handlers use `contextx.IContext` and `base` ORM patterns (soft-delete via `AliveFilter()`).

## ANTI-PATTERNS

- Do not add install/workflow policy or permission checks in this package.
- Do not bypass `TableBroker` / `base.IOrm` conventions (see `pkg/dao/mongo/base/AGENTS.md`).
- Do not expose raw DAO models outside the package; callers should use `pkg/types` via `IHandler`.
