# NODE-WORKFLOW DATA ACCESS KNOWLEDGE BASE

## OVERVIEW

`pkg/dao/mongo/node-workflow` is the data-access layer for node workflow records in MongoDB.

It contains table schema binding, query options, and DAO handler implementation.  
Persistent workflow records map to shared domain types in `pkg/types/node_workflow.go`.

## WHERE TO LOOK

| Task | Location |
|------|----------|
| Mongo table name | `table.go` |
| Data schema and unique key definition (`Data`, `Table`) | `table.go` |
| Field keys used for filters/distinct | `constant.go` |
| Query option builders (`With*`, `Without*`, `WithOperateTimeRange`) | `options.go` |
| DAO interfaces and implementation methods | `handler.go` |
| Workflow enum/type validation and DTOs | `pkg/types/node_workflow.go` |
| Upstream business usage | `internal/backend/manager/node_manager.go` |

## TABLE MODEL

### Core schema

- `TableName(tenantID string)` returns `node_workflow_<tenantID>`.
- `Data` is the persisted document model (implements `base.IData`).
- `Data.UniqueFields()` and `Data.UniqueKey()` enforce uniqueness by `workflow_id`.
- `Table` is `base.TableBroker[*Data]`.
 
### Stored fields

- `tenant_id`, `workflow_id`, `trigger_id`, `type`
- `biz_ids`, `networkarea_ids`, `networkunit_ids`
- `operator`, `operate_time`, `finish_time`, `status`

### Filter field keys

- Defined in `constant.go`.
- Example keys:
  - `data.workflow_id`
  - `data.status`
  - `data.type`
  - `data.biz_ids`
  - `data.operator`
  - `data.operate_time`
  - `data.finish_time`

## QUERY AND HANDLER API

### Options (`options.go`)

- `WithWorkflowID` / `WithoutWorkflowID`
- `WithStatus` / `WithoutStatus`
- `WithType` / `WithoutType`
- `WithBizID` / `WithoutBizID`
- `WithOperator` / `WithoutOperator`
- `WithOperateTimeRange`

### Handler (`handler.go`)

- `New(client)` creates a tenant-aware handler.
- `IHandler` provides:
  - `Get`
  - `Count`
  - `List`
  - `Create`
  - `UpdateStatus`
  - `UpdateFinishTime`
  - `Distinct*` helpers for `type` / `status` / `biz-id` / `operator`

## SCHEMA CONVENTIONS

- Keep Mongo tags (`bson`) aligned with JSON tags and field names already used by storage callers.
- Treat this package as schema binding only; avoid business rules or validation logic here.
- Use shared `pkg/types` enums (`NodeWorkflowType`, `NodeWorkflowStatus`) for semantic checks in service layers.

## ANTI-PATTERNS

- Do not add business validation logic in this package (e.g. status transitions, permission checks).
- Do not introduce tenant-unsafe collection operations; keep tenant isolation via `TableName`.
- Do not redefine workflow type/status validation rules here; keep them in `pkg/types/node_workflow.go` and service/manager layers.
