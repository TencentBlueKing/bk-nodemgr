### Description

- API Version: v3.0.1+.
- Required Permission: networkunit_use_for_agent (Use Network Unit for Agent), agent_operate (Operate Agent).
- Function: Batch-assign unassigned hosts to a specified network unit (metadata-only, no remote operations).

### URL

POST /api/v3/node/agent/assign_unit

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_host_id | int64 array | Yes | List of host IDs to assign, cannot be empty |
| bk_networkunit_id | int64 | Yes | Target network unit ID, must be greater than or equal to 0 |

**Parameter Notes**:
- All hosts must belong to the same network area, and the network area must match the target network unit's network area
- Hosts that are already assigned to a network unit will be skipped and recorded as failed
- Duplicate host IDs are automatically deduplicated

### Request Example

Assign 3 hosts to network unit ID 5.

```json
{
  "bk_host_id": [10001, 10002, 10003],
  "bk_networkunit_id": 5
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {
    "success_count": 2,
    "failed_count": 1,
    "failed_reasons": [
      "host-id(10003) already assigned to networkunit-id(3)"
    ]
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, 0 indicates success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, null on success |
| data | object | Response data |

#### data

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| success_count | int64 | Number of hosts successfully assigned |
| failed_count | int64 | Number of hosts that failed to be assigned |
| failed_reasons | string array | List of failure reasons, each describing the specific host and reason |

**Common failure reasons**:
- `host-id(xxx) not found`: The specified host ID does not exist
- `host-id(xxx) already assigned to networkunit-id(yyy)`: The host is already assigned to another network unit

#### error

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| system | string | Error system identifier |
| message | string | Error message |
| details | array | Error details list |

#### error.details[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | string | Error code |
| message | string | Error message |
