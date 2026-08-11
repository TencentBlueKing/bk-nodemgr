### Description

- API Version: v3.0.1-alpha.20+.
- Required Permission: networkunit_use_for_proxy (Use Network Unit to Deploy Proxy), proxy_operate (Operate Proxy).
- Function: Batch-assign multiple groups of running, unassigned Proxy hosts to corresponding network units and launch an assignment workflow.

### URL

POST /api/v3/node/proxy/assign_unit_multi

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| items | object array | Yes | Assignment item list, cannot be empty |
| items[n].bk_host_id | int64 array | Yes | List of Proxy host IDs to assign, cannot be empty |
| items[n].bk_networkunit_id | int64 | Yes | Target network unit ID, must be greater than or equal to 0 |

**Parameter Notes**:

- Each host ID can appear in only one assignment item; duplicate host IDs return a parameter error
- Proxy hosts in each assignment item are assigned to the network unit specified by that item
- All found hosts must be running Proxy hosts and must not already be assigned to a network unit
- Each host's network area must match the corresponding target network unit's network area
- Non-existent host IDs are recorded as failure reasons; eligible hosts are still submitted to the assignment workflow

### Request Example

Assign different Proxy hosts to different network units.

```json
{
  "items": [
    {
      "bk_host_id": [10001, 10002],
      "bk_networkunit_id": 5
    },
    {
      "bk_host_id": [10003],
      "bk_networkunit_id": 6
    }
  ]
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
    "failed_reasons": ["host-id(10003) not found"],
    "workflow_id": "wf-20240101-001"
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
| success_count | int64 | Number of Proxy hosts successfully submitted to the assignment workflow |
| failed_count | int64 | Number of Proxy hosts that failed to be assigned, including non-existent hosts and hosts affected by workflow launch failure |
| failed_reasons | string array | List of failure reasons, each describing a specific host failure or workflow launch failure |
| workflow_id | string | Assignment workflow ID; empty when workflow launch fails |

**Common failure reasons**:

- `host-id(xxx) not found`: The specified host ID does not exist
- `node role is not proxy. host-id(xxx), node-role(xxx)`: The host is not a Proxy node
- `host-id(xxx) is not online, status(xxx)`: The Proxy host is not running
- `host-id(xxx) already assigned to networkunit-id(yyy)`: The Proxy host is already assigned to another network unit
- `host network area (xxx) does not match target networkunit area (yyy)`: The host network area does not match the target network unit's network area
- `failed to launch workflow: xxx`: Failed to launch the assignment workflow

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
