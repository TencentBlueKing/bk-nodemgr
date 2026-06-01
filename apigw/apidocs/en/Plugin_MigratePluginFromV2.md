### Description

- API Version: v3.0.1-alpha.32+.
- Required Permission: plugin_operate (Operate Plugin).
- Function: Batch migrate plugin processes from v2 to the current plugin workflow on specified hosts.

### URL

POST /api/v3/plugin/migrate_from_v2

### Input Parameters

| Parameter Name | Parameter Type | Required | Description                                             |
| -------------- | -------------- | -------- | ------------------------------------------------------- |
| plugin         | object array   | Yes      | Plugin migration targets; at least one item is required |

#### plugin[n]

| Parameter Name | Parameter Type | Required | Description    |
| -------------- | -------------- | -------- | -------------- |
| bk_host_id     | int64          | Yes      | Target host ID |
| plugin_name    | string         | Yes      | Plugin name    |

**Parameter Notes**:

- `plugin`: At least one item must be provided, otherwise request validation fails.
- `bk_host_id`: Host ID must be a positive integer.
- `plugin_name`: Must not be empty.

### Request Example

```json
{
  "plugin": [
    {
      "bk_host_id": 1001,
      "plugin_name": "bk-monitor-agent"
    },
    {
      "bk_host_id": 1002,
      "plugin_name": "bk-log-collector"
    }
  ]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260601-000001",
  "error": null,
  "data": {
    "workflow_id": "plugin-migrate-v2-workflow-123456"
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description                        |
| -------------- | -------------- | ---------------------------------- |
| code           | int32          | Status code, 0 indicates success   |
| message        | string         | Request message                    |
| request_id     | string         | Request ID                         |
| error          | object         | Error information, null on success |
| permission     | object         | Permission information             |
| data           | object         | Response data                      |

#### data

| Parameter Name | Parameter Type | Description                                        |
| -------------- | -------------- | -------------------------------------------------- |
| workflow_id    | string         | Workflow ID for tracking the plugin migration task |

#### error

| Parameter Name | Parameter Type | Description             |
| -------------- | -------------- | ----------------------- |
| system         | string         | Error system identifier |
| message        | string         | Error message           |
| details        | array          | Error detail list       |

#### error.details[n]

| Parameter Name | Parameter Type | Description   |
| -------------- | -------------- | ------------- |
| code           | string         | Error code    |
| message        | string         | Error message |

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
