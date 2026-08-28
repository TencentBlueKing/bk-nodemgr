### Description

- API Version: v3.0.1-alpha.77+.
- Required Permission: plugin_view (View Plugin).
- Function: Query the plugin configuration files on a target host by host ID and plugin name.

### URL

POST /api/v3/plugin/list_config_files

### Input Parameters

| Parameter Name | Parameter Type | Required | Description                                                          |
| -------------- | -------------- | -------- | -------------------------------------------------------------------- |
| bk_host_id     | int64          | Yes      | Host ID                                                              |
| plugin_name    | string         | Yes      | Plugin name; used to match the process name in configuration records |

**Parameter Notes**:

- `bk_host_id`: A non-zero value is required, otherwise request validation fails.
- `plugin_name`: A non-empty string is required, otherwise request validation fails.
- Permission is checked against the business that the target host belongs to. The caller must have plugin view permission for that business.

### Request Example

Query the configuration files of `bk-monitor-agent` on host `1001`.

```json
{
  "bk_host_id": 1001,
  "plugin_name": "bk-monitor-agent"
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
    "items": [
      {
        "name": "bkmonitoragent.conf",
        "template_name": "bkmonitoragent.conf",
        "process_name": "bk-monitor-agent",
        "bk_host_id": 1001,
        "set": "default",
        "is_main_config": true,
        "content": "dataid: 100001\nperiod: 60\n",
        "md5": "d9d21a34e52e49dff8e6f5aaeb5c3d51",
        "file_path": "etc/bkmonitoragent.conf",
        "custom_config_context": {
          "cluster_id": "prod-ap-guangzhou",
          "labels": {
            "env": "prod"
          }
        }
      }
    ]
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

| Parameter Name | Parameter Type | Description                    |
| -------------- | -------------- | ------------------------------ |
| items          | object array   | Plugin configuration file list |

#### data.items[n]

| Parameter Name        | Parameter Type | Description                                                                  |
| --------------------- | -------------- | ---------------------------------------------------------------------------- |
| name                  | string         | Configuration file name                                                      |
| template_name         | string         | Original plugin configuration template name                                  |
| process_name          | string         | Plugin process name that the configuration belongs to                        |
| bk_host_id            | int64          | Host ID                                                                      |
| set                   | string         | The `set` field in the configuration record                                  |
| is_main_config        | bool           | Whether it is the main configuration file                                    |
| content               | string         | Configuration file content                                                   |
| md5                   | string         | MD5 of the configuration file content                                        |
| file_path             | string         | Configuration file path relative to the plugin deployment directory          |
| custom_config_context | object         | Custom configuration render context for the plugin process; free-form object |

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
