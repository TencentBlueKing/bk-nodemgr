### Description

- API Version: v3.0.1-alpha.56+.
- Required Permission: agent_operate (Operate Agent), proxy_operate (Operate Proxy).
- Function: Get manual handling information for a specified operation, including network policies and manual commands.

**Permission Notes**: Backend narrows `agent_operate` or `proxy_operate` permissions by the node roles of the workflow identified by `workflow_id`.

### URL

POST /api/v3/node/workflow/operation/manual/info/get

### Request Parameters

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| workflow_id | string | Yes | Workflow ID |
| operation_id | string | Yes | Operation ID |

### Request Example

```json
{
  "workflow_id": "wf-20240601-0001",
  "operation_id": "op-20240601-0001"
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "network_policies": [],
    "commands": [
      {
        "type": "bash",
        "command": "bash install.sh"
      },
      {
        "type": "bat",
        "command": "install.bat"
      }
    ]
  }
}
```

### Response Parameters

| Parameter | Type | Description |
| --- | --- | --- |
| code | int32 | Status code, `0` means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, usually empty on success |
| permission | object | Permission information |
| data | object | Response data |

#### data

| Parameter | Type | Description |
| --- | --- | --- |
| network_policies | object array | Network policy list |
| commands | object array | Manual command list |

#### data.network_policies[]

| Parameter | Type | Description |
| --- | --- | --- |
| name | string | Network policy name |
| description_en | string | English description |
| description_zh | string | Chinese description |
| source | object | Source endpoint |
| target | object | Target endpoint |
| service | object | Service configuration |

#### source / target

| Parameter | Type | Description |
| --- | --- | --- |
| name | string | Endpoint name |
| type | string | Endpoint type |
| values | string array | Endpoint values |
| description_en | string | English description |
| description_zh | string | Chinese description |

#### service

| Parameter | Type | Description |
| --- | --- | --- |
| protocol | string | Protocol |
| ports | string array | Port list |
| description_en | string | English description |
| description_zh | string | Chinese description |

#### data.commands[]

| Parameter | Type | Description |
| --- | --- | --- |
| type | string | Command type, available values: `bash`, `bat` |
| command | string | Command content |
