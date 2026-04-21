### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: agent_operate (Operate Agent), proxy_operate (Operate Proxy).
  - Permission Note: Dynamically narrowed based on node roles in the workflow. If the workflow only involves Agent nodes, only agent_operate is required; if only Proxy nodes, only proxy_operate is required; if both or unknown roles, both are required.
- Function: Retry operations in a specified node workflow, supporting full retry or partial retry by operation IDs.

### URL

POST /api/v3/node/workflow/operation/retry

### Request Parameters

| Parameter     | Type         | Required | Description                                                                                                    |
| ------------- | ------------ | -------- | -------------------------------------------------------------------------------------------------------------- |
| workflow_id   | string       | Yes      | Workflow ID                                                                                                    |
| retry_mod     | string       | Yes      | Retry mode, available values: `ALL` (retry all failed operations), `PARTIAL` (retry only specified operations) |
| operation_ids | string array | Yes      | Operation ID list, used to specify target operations when `retry_mod=PARTIAL`                                  |

### Request Example

Retry only part of failed operations.

```json
{
  "workflow_id": "wf-20240601-0001",
  "retry_mod": "PARTIAL",
  "operation_ids": ["op-001", "op-003"]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {}
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
| data       | object | Response data, empty object on success      |

#### error

| Parameter | Type   | Description         |
| --------- | ------ | ------------------- |
| system    | string | Error source system |
| message   | string | Error message       |
| details   | array  | Error detail list   |

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

This object is empty, returns `{}`.
