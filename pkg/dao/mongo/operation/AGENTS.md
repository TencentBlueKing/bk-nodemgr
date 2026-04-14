# Operation DAO

## Scope
DAO layer for `operation` collection - workflow execution unit persistence and query interface.

## Responsibility
- Operation CRUD operations (create, read, list, count)
- Query filtering by operation_id, trigger_id, parent_operation_id, instantiation state, instance lifecycle state
- Query filtering by init_content fields (currently: token)
- Index management for query performance
- Data structure definition and migration compatibility

## Data Model

### Core Structure
```
Operation {
  operation_id: string (unique key)
  trigger_id: string (links to workflow trigger)
  oper_inst_ids: []string (operation instance IDs)
  def_snapshot: DefSnapshot (operation definition snapshot)
  parameters: Parameters {
    parent_operation_id: string
    timeout: duration
    init_content: map[string]any (flexible initialization data)
    retry_start_point: map[string]bool
  }
  retry_flags: []RetryFlag
  instantiated: bool
  latest_inst_brief_data: InstBriefData (latest instance state)
  create_time: time
}
```

### Key Fields
- `operation_id`: Unique identifier for the operation
- `trigger_id`: Links operation to workflow trigger (used for workflow→operation queries)
- `init_content`: Flexible map storing initialization parameters (e.g., token, operator, and potentially IP in future)
- `latest_inst_brief_data.life_cycle.state`: Current execution state of the operation

## Query Patterns

### Supported Filters
- `WithOperationID(ids...)`: Query by operation ID(s)
- `WithTriggerID(triggerID...)`: Query operations by trigger ID (workflow→operation lookup)
- `WithParentOperationID(parentID...)`: Query child operations
- `WithInstantiated(bool...)`: Filter by instantiation state
- `WithLatestInstState(state...)`: Filter by current execution state
- `WithInitContentToken(token...)`: Query by init_content.token field

### Index Strategy
Current indexes:
- `operation_id` (unique)
- `trigger_id` (workflow association)
- `init_content.token` (deployment association)
- `latest_inst_brief_data.life_cycle.state` (state filtering)

Future considerations:
- `init_content.inner_ip` / `init_content.inner_ipv6` (IP-based workflow queries)

## Design Constraints

### init_content Flexibility
- `init_content` is `map[string]any` - no schema enforcement at DAO level
- Current usage: `{token: string, operator: string}`
- Extensible for future fields (e.g., IP, host_id) without schema migration
- Callers must handle missing/unexpected fields gracefully

### Query Scope
- Provides atomic query operations only
- Complex multi-table joins (e.g., operation→deployment→workflow) belong in storage layer
- No business logic or orchestration in DAO

### Migration Compatibility
- Adding new init_content fields is backward-compatible (map structure)
- Adding new top-level fields requires careful migration planning
- Index additions should be tested for performance impact on large collections

## Integration Points

### Upstream (Callers)
- `internal/backend/storage/node/dao_node_workflow.go`: Queries operations by trigger_id for workflow IP filtering
- `internal/backend/storage/plugin/dao_plugin_workflow.go`: Similar pattern for plugin workflows
- Workflow engine: Creates and updates operations during execution

### Downstream (Dependencies)
- MongoDB `operation` collection
- `pkg/dao/mongo/base`: Shared query builder and filter utilities

## Evolution Path

### Planned Enhancements
1. **IP-based filtering**: Add `init_content.inner_ip` / `init_content.inner_ipv6` fields and indexes
   - Enables 2-step workflow queries: IP→operation→workflow (vs current 3-step via deployment)
   - Requires updating operation creation logic to populate IP fields
2. **Richer query options**: Add time-range filtering, batch operations
3. **Performance optimization**: Evaluate compound indexes for common query patterns

### Anti-Patterns
- ❌ Do not add business logic or validation rules in DAO layer
- ❌ Do not assume init_content fields always exist - handle missing fields
- ❌ Do not bypass DAO layer and query MongoDB directly from upper layers
- ❌ Do not store large blobs in init_content (keep it lightweight)

## Reference
- README.md: Package overview and design intent
- constant.go: Field key constants for query building
- options.go: Filter function implementations
- table.go: Data structure definitions
