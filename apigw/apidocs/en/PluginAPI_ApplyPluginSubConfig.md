### Description

- API Version: v3.0.1+.
- Required Permission: .
- Function: Batch apply plugin sub-configurations to specified hosts, with optional custom render context.

### URL

POST /api/v3/plugin/apply_subconfig

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| plugin | object array | Yes | Plugin targets for sub-config application; at least one item is required |

#### plugin[n]

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_host_id | int64 | No | Target host ID; defaults to -1 when omitted, but must be a positive integer in practice |
| plugin_name | string | Yes | Plugin name |
| config_name | string array | No | Sub-configuration file names to apply; values come from config template names in the plugin release package |
| custom_config_context | object | No | Custom configuration render context; free-form object with reserved top-level keys forbidden |

**Parameter Notes**:
- `plugin`: At least one item must be provided, otherwise request validation fails.
- `bk_host_id`: Although optional in proto, omitting it causes the server to auto-fill `-1`, which then fails validation. In practice, provide a valid host ID.
- `config_name`: The current implementation does not force this field to be non-empty. When omitted, the system falls back to default behavior based on plugin configuration details.
- `custom_config_context`: Custom variables for template rendering. Reserved top-level keys are forbidden, including `PluginInfo`, `NodeInfo`, `PreDefinitionConstants`, `CustomContext`, `plugin_path`, `nodeman`, `cmdb_instance`, `target`, and `control_info`.

### Request Example

Apply sub-configurations for `bk-monitor-agent` on a target host with custom render variables.

```json
{
  "plugin": [
    {
      "bk_host_id": 1001,
      "plugin_name": "bk-monitor-agent",
      "config_name": [
        "bkmonitoragent.conf",
        "env.yaml"
      ],
      "custom_config_context": {
        "cluster_id": "prod-ap-guangzhou",
        "labels": {
          "env": "prod"
        }
      }
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
    "workflow_id": "plugin-apply-subconfig-workflow-123456"
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, 0 indicates success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, null on success |
| data | object | Response data |

#### data

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| workflow_id | string | Workflow ID for tracking the plugin sub-config apply task |

#### error

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| system | string | Error system identifier |
| message | string | Error message |
| details | array | Error detail list |

#### error.details[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | string | Error code |
| message | string | Error message |
