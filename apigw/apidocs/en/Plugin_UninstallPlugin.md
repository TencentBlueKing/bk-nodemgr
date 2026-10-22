### Description

- API Version: v3.0.1+.
- Required Permission: plugin_operate (Operate Plugin).
- Function: Batch uninstall plugins from specified hosts.

### URL

POST /api/v3/plugin/uninstall

### Input Parameters

| Parameter Name | Parameter Type | Required | Description                                             |
| -------------- | -------------- | -------- | ------------------------------------------------------- |
| plugin         | object array   | Yes      | Plugin uninstall targets; at least one item is required |

#### plugin[n]

| Parameter Name | Parameter Type | Required | Description                                |
| -------------- | -------------- | -------- | ------------------------------------------ |
| bk_host_id     | int64          | Yes      | Target host ID; must be a positive integer |
| plugin_name    | string         | Yes      | Plugin name                                |

**Parameter Notes**:

- `plugin`: At least one item must be provided, otherwise request validation fails.
- `bk_host_id`: Provide a valid positive host ID. Omitting it or passing `0` or a negative value fails.
- `plugin_name`: Must not be empty.
- A successful response means the uninstallation task has been submitted, not completed. Use `workflow_id` to query the plugin workflow and operation status.

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
  "request_id": "req-20260324-000005",
  "error": null,
  "data": {
    "workflow_id": "plugin-uninstall-workflow-123456"
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
| data           | object         | Response data                      |

#### data

| Parameter Name | Parameter Type | Description                                        |
| -------------- | -------------- | -------------------------------------------------- |
| workflow_id    | string         | Workflow ID for tracking the plugin uninstall task |

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
