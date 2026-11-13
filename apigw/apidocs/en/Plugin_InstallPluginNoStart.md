### Description

- API Version: v3.0.1+.
- Required Permission: plugin_operate (Operate Plugin).
- Function: Batch deliver the V3 plugin packages and render their configs on specified hosts **without** starting
  the plugin processes. If the plugin process on a target host is running, the workflow stops it first. This API does
  not take part in V2/V3 compatibility handling: it neither stops V2 plugins nor writes compatibility configs.

### URL

POST /api/v3/plugin/install_no_start

### Input Parameters

| Parameter Name | Parameter Type | Required | Description                                                |
| -------------- | -------------- | -------- | ---------------------------------------------------------- |
| plugin         | object array   | Yes      | Plugin installation targets; at least one item is required |

#### plugin[n]

| Parameter Name        | Parameter Type | Required | Description                                                                        |
| --------------------- | -------------- | -------- | ---------------------------------------------------------------------------------- |
| bk_host_id            | int64          | Yes      | Target host ID; must be a positive integer                                         |
| plugin_name           | string         | Yes      | Plugin name                                                                        |
| version               | string         | No       | Plugin version to install; when omitted or empty, the plugin's current default release version is used |
| custom_config_context | object         | No       | Custom configuration render context; free-form object with reserved top-level keys forbidden |
| custom_spec           | object         | No       | Custom process resources and monitoring policy; replaces defaults and matched configuration policies in full |

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

- `plugin`: At least one item must be provided, otherwise request validation fails.
- `bk_host_id`: Provide a valid positive host ID. Omitting it or passing `0` or a negative value fails.
- `version`: When omitted or empty, the current default release version of the plugin is used.
- `custom_config_context`: Custom variables used to render config templates; all main config files of the same plugin share this context.
  Reserved top-level keys are forbidden: `PluginInfo`, `NodeInfo`, `PreDefinitionConstants`, `CustomContext`, `plugin_path`,
  `nodeman`, `cmdb_instance`, `target`, `control_info`.
- `custom_spec`: When omitted, the default config and the matched config policy are used. When provided, it replaces the resource
  and monitor policy as a whole (no field-level merge); omitted objects or numeric fields are treated as zero values, and an omitted
  `restart_type` is treated as an empty string. Both sections should be provided completely.
- A successful response only means the delivery task has been submitted, not that the plugin package has been delivered.
  Use `workflow_id` to query the workflow and operation status. Reserved key validation happens during workflow execution.
- After the workflow succeeds, the plugin process record of the host is marked as `stopped` and the process is not started.
  Call `POST /api/v3/plugin/start` later to start the plugin process on the same host.

### Request Example

Deliver the `bk-monitor-agent` package (without starting it) on two hosts, with custom rendering variables and
custom process resource/monitor policy for one of them.

```json
{
  "plugin": [
    {
      "bk_host_id": 1001,
      "plugin_name": "bk-monitor-agent",
      "version": "2.4.0",
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
  "request_id": "req-20260324-000002",
  "error": null,
  "data": {
    "workflow_id": "plugin-install-no-start-workflow-123456"
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

| Parameter Name | Parameter Type | Description                                   |
| -------------- | -------------- | --------------------------------------------- |
| workflow_id    | string         | Workflow ID for tracking the plugin delivery (without start) task |

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
