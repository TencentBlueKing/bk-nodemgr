### Description

- API Version: v3.0.1+。
  resolution).
- Required Permission: .
- Function: Batch install plugins on specified hosts, with optional sub-config selection and custom render context.

### URL

POST /api/v3/plugin/install

### Input Parameters

| Parameter Name | Parameter Type | Required | Description                                                |
|----------------|----------------|----------|------------------------------------------------------------|
| plugin         | object array   | Yes      | Plugin installation targets; at least one item is required |

#### plugin[n]

| Parameter Name        | Parameter Type | Required | Description                                                                                                        |
|-----------------------|----------------|----------|--------------------------------------------------------------------------------------------------------------------|
| bk_host_id            | int64          | No       | Target host ID; defaults to -1 when omitted                                                                        |
| plugin_name           | string         | Yes      | Plugin name                                                                                                        |
| version               | string         | Yes      | Plugin version to install                                                                                          |
| config_name           | string array   | No       | Sub-configuration file names to apply; values come from config template names in the target plugin release package |
| custom_config_context | object         | No       | Custom configuration render context; free-form object with reserved top-level keys forbidden                       |

**Parameter Notes**:

- `plugin`: At least one plugin item must be provided, otherwise request validation fails.
- `bk_host_id`: When omitted, the server auto-fills `-1`; negative values fail validation.
- `config_name`: This is not a fixed enum. Valid values depend on the configuration template names shipped in the target
  plugin release package. When omitted, the main configuration is used by default.
- `custom_config_context`: Custom variables for template rendering. Reserved top-level keys are forbidden, including
  `PluginInfo`, `NodeInfo`, `PreDefinitionConstants`, `CustomContext`, `plugin_path`, `nodeman`, `cmdb_instance`,
  `target`, and `control_info`.

### Request Example

Install `bk-monitor-agent` on two hosts, with one host specifying a sub-config and custom render variables.

```json
{
  "plugin": [
    {
      "bk_host_id": 1001,
      "plugin_name": "bk-monitor-agent",
      "version": "2.4.0",
      "config_name": [
        "bkmonitoragent.conf"
      ],
      "custom_config_context": {
        "cluster_id": "prod-ap-guangzhou",
        "labels": {
          "env": "prod",
          "region": "ap-guangzhou"
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
|----------------|----------------|------------------------------------|
| code           | int32          | Status code, 0 indicates success   |
| message        | string         | Request message                    |
| request_id     | string         | Request ID                         |
| error          | object         | Error information, null on success |
| data           | object         | Response data                      |

#### data

| Parameter Name | Parameter Type | Description                                           |
|----------------|----------------|-------------------------------------------------------|
| workflow_id    | string         | Workflow ID for tracking the plugin installation task |

#### error

| Parameter Name | Parameter Type | Description             |
|----------------|----------------|-------------------------|
| system         | string         | Error system identifier |
| message        | string         | Error message           |
| details        | array          | Error detail list       |

#### error.details[n]

| Parameter Name | Parameter Type | Description   |
|----------------|----------------|---------------|
| code           | string         | Error code    |
| message        | string         | Error message |
