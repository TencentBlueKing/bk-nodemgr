### Description

- API Version: v3.0.1-alpha.77+.
- Required Permission: plugin_operate (Operate Plugin).
- Function: Batch remove plugin sub-configuration files from specified hosts; the plugin is reloaded after removal.

### URL

POST /api/v3/plugin/remove_subconfig

### Input Parameters

| Parameter Name | Parameter Type | Required | Description                                                          |
| -------------- | -------------- | -------- | -------------------------------------------------------------------- |
| plugin         | object array   | Yes      | Plugin targets for sub-config removal; at least one item is required |

#### plugin[n]

| Parameter Name       | Parameter Type | Required | Description                                                                                  |
| -------------------- | -------------- | -------- | -------------------------------------------------------------------------------------------- |
| bk_host_id           | int64          | Yes      | Target host ID                                                                               |
| plugin_name          | string         | Yes      | Plugin name                                                                                  |
| config_template_name | string array   | No       | Template names of sub-configurations to remove; mutually exclusive with `config_file_name`   |
| config_file_name     | string array   | No       | Exact sub-configuration file names to remove; mutually exclusive with `config_template_name` |

**Parameter Notes**:

- `plugin`: At least one item must be provided, otherwise request validation fails.
- `bk_host_id`: Must be a positive integer.
- `plugin_name`: Must not be empty.
- `config_template_name`: Matches target sub-configuration files by config template name. The server resolves the actual config file names. If a template name does not exist, request validation fails.
- `config_file_name`: Specifies target sub-configuration files by exact file name.
- `config_template_name` and `config_file_name` are mutually exclusive and one of them is required for each `plugin[n]` item.
- Duplicate values in `config_template_name` and `config_file_name` are deduplicated by the server. Empty string items are not allowed.

### Request Example

Remove sub-configurations for `bk-monitor-agent` on a target host by config template name.

```json
{
  "plugin": [
    {
      "bk_host_id": 1001,
      "plugin_name": "bk-monitor-agent",
      "config_template_name": ["env.yaml"]
    }
  ]
}
```

Remove sub-configurations from a target host by exact config file name.

```json
{
  "plugin": [
    {
      "bk_host_id": 1002,
      "plugin_name": "bk-log-collector",
      "config_file_name": ["custom-log-path.conf"]
    }
  ]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260828-000001",
  "error": null,
  "data": {
    "workflow_id": "plugin-remove-subconfig-workflow-123456"
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

| Parameter Name | Parameter Type | Description                                                 |
| -------------- | -------------- | ----------------------------------------------------------- |
| workflow_id    | string         | Workflow ID for tracking the plugin sub-config removal task |

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
