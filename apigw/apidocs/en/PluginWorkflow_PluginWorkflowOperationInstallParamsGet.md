### Description

- API Version: v3.0.1-alpha.120+.
- Required Permission: system_access (System Access), plugin_history_view (View Plugin History), plugin_view (View Plugin). Both plugin permissions are checked for all businesses associated with the historical workflow.
- Function: Read recoverable parameters from selected historical plugin installation operations to prefill a new installation form. This API does not install, retry, or modify historical tasks.

**Usage Boundaries**:

- Only `install_plugin` workflows are supported, regardless of operation execution status. Use the [operation list API](PluginWorkflow_PluginWorkflowOperationList.md) to obtain operation IDs.
- Parameters come from historical deployment records, not a complete snapshot of the original request. Operations whose records expired or are missing appear in `unavailable_operation_ids` instead of `items`.
- `missing_fields` identifies inputs that could not be recovered. Custom configuration is returned as recorded, without system-generated configuration context, rendered configuration, or deployment tokens. Handle any sensitive business information in custom configuration appropriately.
- Historical parameters do not guarantee that hosts, plugin versions, or configuration remain valid. Confirm and complete the parameters before creating a new task through the normal [plugin installation API](Plugin_InstallPlugin.md).

### URL

POST /api/v3/plugin/workflow/operation/install_params/get

### Request Parameters

| Parameter     | Type         | Required | Description                                                                                                                                                                         |
| ------------- | ------------ | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| workflow_id   | string       | Yes      | Historical installation workflow ID in the current tenant; must not be empty or whitespace-only                                                                                     |
| operation_ids | string array | Yes      | Operation IDs in the selected workflow; nonempty list with no empty, whitespace-only, or duplicate IDs. A nonexistent operation or operation outside the workflow fails the request |

### Request Example

```json
{
  "workflow_id": "wf-plugin-0001",
  "operation_ids": ["op-plugin-0001", "op-plugin-0002"]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "type": "install_plugin",
    "items": [
      {
        "operation_id": "op-plugin-0001",
        "bk_host_id": 1001,
        "plugin_name": "bkmonitorbeat",
        "version": "3.1.0",
        "custom_config_context": { "interval": 60 },
        "custom_spec": null,
        "missing_fields": []
      }
    ],
    "unavailable_operation_ids": ["op-plugin-0002"]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                      |
| ---------- | ------ | -------------------------------- |
| code       | int32  | Status code; 0 indicates success |
| message    | string | Request result message           |
| request_id | string | Request ID                       |
| data       | object | Recovery result                  |

#### data

| Parameter                 | Type         | Description                                        |
| ------------------------- | ------------ | -------------------------------------------------- |
| type                      | string       | Workflow type; currently `install_plugin`          |
| items                     | object array | Recoverable parameters in request order            |
| unavailable_operation_ids | string array | Operation IDs without available deployment records |

#### items[]

| Parameter             | Type         | Description                                                                                  |
| --------------------- | ------------ | -------------------------------------------------------------------------------------------- |
| operation_id          | string       | Historical operation ID                                                                      |
| bk_host_id            | int64        | Historical installation target host ID                                                       |
| plugin_name           | string       | Plugin name                                                                                  |
| version               | string       | Version stored in the historical deployment record                                           |
| custom_config_context | object       | Historical custom configuration parameters; may be empty                                     |
| custom_spec           | object       | Historical custom deployment specification; null if unspecified                              |
| missing_fields        | string array | Unrecoverable field names; do not treat empty values for these fields as previous selections |

#### custom_spec

| Parameter      | Type   | Description                      |
| -------------- | ------ | -------------------------------- |
| resource       | object | Custom resource limits           |
| monitor_policy | object | Custom process monitoring policy |

#### resource

| Parameter         | Type   | Description                        |
| ----------------- | ------ | ---------------------------------- |
| cpu_limit_percent | double | CPU usage limit as a percentage    |
| mem_limit_percent | double | Memory usage limit as a percentage |

#### monitor_policy

| Parameter               | Type   | Description                       |
| ----------------------- | ------ | --------------------------------- |
| restart_type            | string | Restart mode: `auto` or `manual`  |
| start_check_seconds     | int64  | Startup check duration in seconds |
| stop_check_seconds      | int64  | Stop check duration in seconds    |
| operate_timeout_seconds | int64  | Operation timeout in seconds      |
