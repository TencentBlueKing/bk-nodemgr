### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: agent_view (View Agent), proxy_view (View Proxy).
- Function: Return operation instance status statistics grouped by workflow ID.

### URL

POST /api/v3/node/workflow/statistics

### Request Parameters

| Parameter   | Type         | Required | Description      |
| ----------- | ------------ | -------- | ---------------- |
| workflow_id | string array | No       | Workflow ID list |

**Permission Notes**:

- The request does not include a `node_role` filter, backend checks both `agent_view` and `proxy_view` by default for safety.

### Request Example

Get status statistics for two workflows.

```json
{
  "workflow_id": ["wf-20240601-0001", "wf-20240601-0002"]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "items": [
      {
        "workflow_id": "wf-20240601-0001",
        "total_count": 10,
        "init_count": 1,
        "launched_count": 1,
        "running_count": 2,
        "success_count": 4,
        "failed_count": 1,
        "timeout_count": 1,
        "terminated_count": 0
      },
      {
        "workflow_id": "wf-20240601-0002",
        "total_count": 0,
        "init_count": 0,
        "launched_count": 0,
        "running_count": 0,
        "success_count": 0,
        "failed_count": 0,
        "timeout_count": 0,
        "terminated_count": 0
      }
    ]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                 |
| ---------- | ------ | ------------------------------------------- |
| code       | int32  | Status code, `0` means success              |
| message    | string | Request message                             |
| request_id | string | Request ID                                  |
| error      | object | Error information, usually empty on success |
| permission | object | Permission information                      |
| data       | object | Response data                               |

#### error

| Parameter | Type   | Description             |
| --------- | ------ | ----------------------- |
| system    | string | Error system identifier |
| message   | string | Error message           |
| details   | array  | Error detail list       |

#### error.details[n]

| Parameter | Type   | Description   |
| --------- | ------ | ------------- |
| code      | string | Error code    |
| message   | string | Error message |

#### permission

| Parameter   | Type   | Description            |
| ----------- | ------ | ---------------------- |
| system      | string | Permission system ID   |
| system_name | string | Permission system name |
| apply_url   | string | Permission apply URL   |
| actions     | array  | Related action list    |

#### permission.actions[n]

| Parameter              | Type   | Description            |
| ---------------------- | ------ | ---------------------- |
| id                     | string | Action ID              |
| name                   | string | Action name            |
| related_resource_types | array  | Related resource types |

#### data

| Parameter | Type  | Description              |
| --------- | ----- | ------------------------ |
| items     | array | Workflow statistics list |

#### data.items[n]

| Parameter        | Type   | Description                                                                            |
| ---------------- | ------ | -------------------------------------------------------------------------------------- |
| workflow_id      | string | Workflow ID                                                                            |
| total_count      | int64  | Total instance count, equals the sum of all state counts and includes not-inited count |
| init_count       | int64  | Count of instances in `init` state                                                     |
| launched_count   | int64  | Count of instances in `launched` state                                                 |
| running_count    | int64  | Count of instances in `running` state                                                  |
| success_count    | int64  | Count of instances in `success` state                                                  |
| failed_count     | int64  | Count of instances in `failed` state                                                   |
| timeout_count    | int64  | Count of instances in `timeout` state                                                  |
| terminated_count | int64  | Count of instances in `terminated` state                                               |
