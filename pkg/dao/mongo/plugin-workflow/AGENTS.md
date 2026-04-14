# Plugin Workflow DAO

## Scope
DAO layer for `plugin_workflow_{tenantID}` collections - plugin workflow lifecycle persistence and query interface with multi-tenant isolation.

## Responsibility
- Plugin workflow CRUD operations (create, read, list, count, distinct)
- Query filtering by workflow_id, trigger_id, host_ids, type, operator, status, time range
- Distinct value queries for filter dropdowns (host_id, operator, status, type)
- Multi-tenant table management (separate collection per tenant)
- Index management for query performance

## Data Model

### Core Structure
```
PluginWorkflow {
  tenant_id: string
  workflow_id: string (unique key)
  trigger_id: string (links to workflow trigger)
  type: string (workflow type: install, uninstall, upgrade, etc.)
  host_ids: []int64 (target hosts for this workflow)
  operator: string (user who initiated the workflow)
  operate_time: time (workflow start time)
  finish_time: time (workflow completion time)
  status: string (workflow status: pending, running, success, failed, etc.)
}
```

### Key Fields
- `workflow_id`: Unique identifier for the workflow
- `trigger_id`: Links workflow to trigger (used for operation→workflow queries)
- `host_ids`: Direct host association - enables 2-step IP filtering (IP→deployment→workflow)
- `type`: Workflow type for filtering by operation category
- `status`: Current workflow state for filtering by execution status

## Query Patterns

### Supported Filters
- `WithWorkflowID(ids...)` / `WithoutWorkflowID(ids...)`: Query by workflow ID(s)
- `WithTriggerID(triggerIDs...)`: Query workflows by trigger ID (operation→workflow lookup)
- `WithHostIDs(hostIDs...)` / `WithoutHostIDs(hostIDs...)`: Query by target host(s)
- `WithType(types...)` / `WithoutType(types...)`: Filter by workflow type
- `WithOperator(operators...)` / `WithoutOperator(operators...)`: Filter by initiating user
- `WithStatus(statuses...)` / `WithoutStatus(statuses...)`: Filter by workflow status
- `WithOperateTimeRange(timeRange)`: Filter by time range

### Distinct Queries
- `DistinctPluginWorkflowBkHostID()`: Get unique host IDs for filter dropdown
- `DistinctPluginWorkflowOperator()`: Get unique operators for filter dropdown
- `DistinctPluginWorkflowStatus()`: Get unique statuses for filter dropdown
- `DistinctPluginWorkflowType()`: Get unique types for filter dropdown

### Index Strategy
Current indexes:
- `workflow_id` (unique)
- `trigger_id` (operation association)
- `host_ids` (host-based filtering)
- `type`, `operator`, `status` (common filter fields)
- `operate_time` (time-range queries and sorting)

## Design Constraints

### Multi-Tenant Isolation
- Each tenant has a separate collection: `plugin_workflow_{tenantID}`
- Table name determined by `TableName(tenantID)` function
- No cross-tenant queries at DAO level
- Tenant ID must be provided for all operations

### host_ids Design
- Stores direct host associations (unlike node_workflow which only has trigger_id)
- Enables efficient IP-based filtering: IP→node_deployment(host_id)→plugin_workflow(host_ids)
- Array field allows one workflow to target multiple hosts
- Indexed for performance on host-based queries

### Query Scope
- Provides atomic query operations only
- Complex multi-table joins (e.g., IP→deployment→workflow) belong in storage layer
- No business logic or orchestration in DAO

### Migration Compatibility
- Adding new fields requires careful migration planning across all tenant tables
- Index additions should be tested for performance impact
- Tenant table creation is automatic on first write

## Integration Points

### Upstream (Callers)
- `internal/backend/storage/plugin/dao_plugin_workflow.go`: Orchestrates multi-step queries (e.g., IP filtering)
- `internal/backend/router/api-v3/plugin/workflow/`: API handlers for workflow list/distinct/statistics
- Workflow engine: Creates and updates workflows during execution

### Downstream (Dependencies)
- MongoDB `plugin_workflow_{tenantID}` collections (one per tenant)
- `pkg/dao/mongo/base`: Shared query builder and filter utilities
- `pkg/types`: PluginWorkflow, PluginWorkflowStatus, PluginWorkflowType definitions

## Comparison with NodeWorkflow

| Aspect | PluginWorkflow | NodeWorkflow |
|--------|----------------|--------------|
| Host association | ✅ Direct `host_ids` field | ❌ Only `trigger_id` (indirect via operation) |
| IP filtering | 2-step: IP→deployment→workflow | 3-step: IP→deployment→operation→workflow |
| Multi-tenant | ✅ Separate tables per tenant | ✅ Separate tables per tenant |
| Distinct queries | ✅ host_id, operator, status, type | ✅ Similar fields |

## Evolution Path

### Current State
- Direct host association enables efficient IP-based filtering
- Multi-tenant isolation is well-established
- Distinct queries support frontend filter dropdowns

### Future Considerations
1. **Performance optimization**: Evaluate compound indexes for common query combinations
2. **Tenant management**: Add utilities for tenant table migration and cleanup
3. **Richer filtering**: Add support for complex conditions (e.g., status transitions, duration ranges)

### Anti-Patterns
- ❌ Do not add business logic or validation rules in DAO layer
- ❌ Do not query across tenant boundaries at DAO level
- ❌ Do not bypass DAO layer and query MongoDB directly from upper layers
- ❌ Do not assume host_ids is always non-empty (some workflows may not target specific hosts)

## Reference
- README.md: Package overview and design intent
- constant.go: Field key constants for query building
- options.go: Filter function implementations
- table.go: Data structure definitions and multi-tenant table naming
