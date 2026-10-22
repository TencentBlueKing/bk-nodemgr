### Description

- API Version: v3.0.1+.
- Required Permission: plugin_operate (Operate Plugin).
- Function: Batch install plugins on specified hosts, with optional sub-config selection and custom render context.

### URL

POST /api/v3/plugin/install

### Input Parameters

| Parameter Name | Parameter Type | Required | Description                                                |
| -------------- | -------------- | -------- | ---------------------------------------------------------- |
| plugin         | object array   | Yes      | Plugin installation targets; at least one item is required |

#### plugin[n]

| Parameter Name        | Parameter Type | Required | Description                                                                                                        |
| --------------------- | -------------- | -------- | ------------------------------------------------------------------------------------------------------------------ |
| bk_host_id            | int64          | Yes      | Target host ID; must be a positive integer                                                                         |
| plugin_name           | string         | Yes      | Plugin name                                                                                                        |
| version               | string         | No       | Plugin version to install; when omitted or empty, the plugin's current default release version is used             |
| config_template_name  | string array   | No       | Sub-configuration file names to apply; values come from config template names in the target plugin release package |
| custom_config_context | object         | No       | Custom configuration render context; free-form object with reserved top-level keys forbidden                       |
| custom_spec           | object         | No       | Custom process resources and monitoring policy; replaces defaults and matched configuration policies in full       |

#### plugin[n].custom_spec

| Parameter Name | Parameter Type | Required | Description               |
| -------------- | -------------- | -------- | ------------------------- |
| resource       | object         | No       | Process resource limits   |
| monitor_policy | object         | No       | Process monitoring policy |

#### plugin[n].custom_spec.resource

| Parameter Name    | Parameter Type | Required | Description                    |
| ----------------- | -------------- | -------- | ------------------------------ |
| cpu_limit_percent | double         | No       | CPU usage limit, in percent    |
| mem_limit_percent | double         | No       | Memory usage limit, in percent |

#### plugin[n].custom_spec.monitor_policy

| Parameter Name          | Parameter Type | Required | Description                                                           |
| ----------------------- | -------------- | -------- | --------------------------------------------------------------------- |
| restart_type            | string         | No       | Restart mode: `auto` (automatic restart) or `manual` (manual restart) |
| start_check_seconds     | int64          | No       | Process survival check time after startup, in seconds                 |
| stop_check_seconds      | int64          | No       | Process status check time after stopping, in seconds                  |
| operate_timeout_seconds | int64          | No       | Process operation timeout, in seconds                                 |

**Parameter Notes**:

- `plugin`: At least one plugin item must be provided, otherwise request validation fails.
- `bk_host_id`: Provide a valid positive host ID. Omitting it or passing `0` or a negative value fails.
- `config_template_name`: This is not a fixed enum. Valid values depend on the configuration template names shipped in the target
  plugin release package. When omitted, the main configuration is used by default.
- `custom_config_context`: Custom variables for template rendering. Reserved top-level keys are forbidden, including
  `PluginInfo`, `NodeInfo`, `PreDefinitionConstants`, `CustomContext`, `plugin_path`, `nodeman`, `cmdb_instance`,
  `target`, and `control_info`.
- `custom_spec`: When omitted, defaults and matched configuration policies are used. When provided, it replaces both resource and monitoring policies in full, without field-by-field merging. Omitted objects or numeric fields become zero values, and an omitted `restart_type` becomes an empty string. Provide both groups in full when using this field.
- A successful response means the installation task has been submitted, not completed. Use `workflow_id` to query the plugin workflow and operation status. Template names and reserved keys are checked during workflow execution; inspect task status and logs for related errors.

### Request Example

Install `bk-monitor-agent` on two hosts, with one host specifying a sub-config, custom render variables, and process resource and monitoring policies.

```json
{
  "plugin": [
    {
      "bk_host_id": 1001,
      "plugin_name": "bk-monitor-agent",
      "version": "2.4.0",
      "config_template_name": ["bkmonitoragent.conf"],
      "custom_config_context": {
        "cluster_id": "prod-ap-guangzhou",
        "labels": {
          "env": "prod",
          "region": "ap-guangzhou"
        }
      },
      "custom_spec": {
        "resource": {
          "cpu_limit_percent": 10,
          "mem_limit_percent": 10
        },
        "monitor_policy": {
          "restart_type": "auto",
          "start_check_seconds": 5,
          "stop_check_seconds": 5,
          "operate_timeout_seconds": 5
        }
      }
    },
    {
      "bk_host_id": 1002,
      "plugin_name": "bk-monitor-agent",
      "version": "2.4.0"
    }
  ]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260324-000001",
  "error": null,
  "data": {
    "workflow_id": "plugin-install-workflow-123456"
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

| Parameter Name | Parameter Type | Description                                           |
| -------------- | -------------- | ----------------------------------------------------- |
| workflow_id    | string         | Workflow ID for tracking the plugin installation task |

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
