### Description

- API Version: v3.0.1-alpha.122+.
- Required Permission: `plugin_operate` (Operate Plugin) for every plugin in the request.
- Function: Distribute applicable dependency files to the plugin configuration directories on the specified hosts and restart the plugin processes.

The workflow matches dependency files against each target plugin process's operating system, CPU architecture, and current plugin package version. Every non-empty scope must match. If applicable files share a destination filename, ignoring case, distribution fails for that target.

For each target, the workflow records the distributed dependencies after all applicable files have been transferred successfully, then restarts the process. If no files match, the workflow skips file transfer and still restarts the process.

### URL

POST /api/v3/plugin/apply_requirements

### Input Parameters

| Parameter Name | Parameter Type | Required | Description                                             |
| -------------- | -------------- | -------- | ------------------------------------------------------- |
| plugin         | object array   | Yes      | Target plugin processes. At least one item is required. |

#### plugin[n]

| Parameter Name | Parameter Type | Required | Description                                                                                     |
| -------------- | -------------- | -------- | ----------------------------------------------------------------------------------------------- |
| bk_host_id     | int64          | Yes      | Target host ID. Must be a positive integer.                                                     |
| plugin_name    | string         | Yes      | Name of an existing plugin process on the target host. Must contain a non-whitespace character. |

The workflow processes repeated host ID and plugin name pairs only once per request.

### Request Example

```json
{
  "plugin": [
    {
      "bk_host_id": 1001,
      "plugin_name": "bk-monitor-agent"
    }
  ]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20261127-000001",
  "error": null,
  "data": {
    "workflow_id": "wf-plugin-0001"
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description                          |
| -------------- | -------------- | ------------------------------------ |
| code           | int32          | `0` means the workflow has started   |
| message        | string         | Request message                      |
| request_id     | string         | Request ID                           |
| error          | object         | Error information, `null` on success |
| permission     | object         | Permission information               |
| data           | object         | Response data                        |

#### data

| Parameter Name | Parameter Type | Description                                                            |
| -------------- | -------------- | ---------------------------------------------------------------------- |
| workflow_id    | string         | Workflow ID for tracking file distribution and restart for each target |

Pass `workflow_id` to `POST /api/v3/plugin/workflow/operation/list` to check each target's execution status.

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

| Parameter Name | Parameter Type | Description            |
| -------------- | -------------- | ---------------------- |
| system         | string         | Permission system ID   |
| system_name    | string         | Permission system name |
| apply_url      | string         | Permission request URL |
| actions        | array          | Related actions        |

#### permission.actions[n]

| Parameter Name         | Parameter Type | Description            |
| ---------------------- | -------------- | ---------------------- |
| id                     | string         | Action ID              |
| name                   | string         | Action name            |
| related_resource_types | array          | Related resource types |
