### Description

- API Version: v3.0.1+.
- Required Permission: None.
- Function: Query distinct field values from topology events under filter conditions, and return candidate sets for event type, network area ID, network unit ID, access point ID, and operator.

### URL

POST /api/v3/topo/event/distinct

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| exact_include_conditions | object | No | Exact match include conditions |
| fuzzy_include_conditions | object | No | Fuzzy match include conditions |
| operate_time_range | object | No | Operation time range in Unix timestamp seconds |

#### exact_include_conditions

Exact match include conditions. Matching any provided condition can place an event into the distinct candidate set.

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_networkarea_id | int64 array | No | Network area ID list |
| bk_networkunit_id | int64 array | No | Network unit ID list |
| accesspoint_id | int64 array | No | Access point ID list |
| type | string array | No | Event type list. Valid values: `networkarea-create`, `networkarea-update`, `networkarea-delete`, `networkunit-create`, `networkunit-update`, `networkunit-delete`, `accesspoint-create`, `accesspoint-update`, `accesspoint-delete` |
| operator | string array | No | Operator username list |

#### fuzzy_include_conditions

Fuzzy match include conditions. Matching any provided condition can place an event into the distinct candidate set.

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_networkarea_name | string array | No | Network area name list, matched fuzzily by name |
| bk_networkunit_name | string array | No | Network unit name list, matched fuzzily by name |

#### operate_time_range

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| start_timestamp_sec | int64 | No | Start of the operation time range, Unix timestamp in seconds |
| end_timestamp_sec | int64 | No | End of the operation time range, Unix timestamp in seconds |

### Request Example

Query topology events on 2026-01-01 whose names contain `prod`, and return distinct candidate values.

```json
{
  "exact_include_conditions": {
    "type": [
      "networkunit-create",
      "accesspoint-update"
    ],
    "operator": [
      "admin"
    ]
  },
  "fuzzy_include_conditions": {
    "bk_networkarea_name": [
      "prod"
    ]
  },
  "operate_time_range": {
    "start_timestamp_sec": 1767225600,
    "end_timestamp_sec": 1767311999
  }
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "bk_networkarea_id": [
      1,
      2
    ],
    "bk_networkunit_id": [
      1001,
      1002
    ],
    "accesspoint_id": [
      10,
      11
    ],
    "type": [
      "networkunit-create",
      "accesspoint-update"
    ],
    "operator": [
      "admin",
      "ops"
    ]
  }
}
```

### Response Parameters

| Parameter | Type | Description |
|---------|----------|------|
| code | int32 | Status code, `0` means success |
| message | string | Response message |
| request_id | string | Request ID |
| error | object | Error information, empty on success |
| permission | object | Permission information, typically empty for this API |
| data | object | Distinct result |

#### data

| Parameter | Type | Description |
|---------|----------|------|
| bk_networkarea_id | int64 array | Distinct network area IDs from topology events matching the conditions |
| bk_networkunit_id | int64 array | Distinct network unit IDs from topology events matching the conditions |
| accesspoint_id | int64 array | Distinct access point IDs from topology events matching the conditions |
| type | string array | Distinct topology event types from topology events matching the conditions |
| operator | string array | Distinct operators from topology events matching the conditions |

### Notes

- The contract of this API is defined by `proto/backend/api/v3/topo.proto`, `pkg/proto/backend/api/v3/event.go`, and `docs/api/swagger/backend/api/v3/topo.swagger.json`.
- The current backend handler `internal/backend/router/api-v3/topo/event.go` converts request conditions into `types.TopoEventCondition`, then calls `DistinctTopoEvent` to return a fixed set of five distinct field groups. The request does not support selecting return columns dynamically.
