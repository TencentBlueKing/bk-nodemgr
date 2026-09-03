### Description

- API Version: v3.0.1-alpha.1+.
- Version Changes: `v3.0.1-alpha.13+` added the permission response field; `v3.0.1-alpha.18+` added host IP and business ID filters.
- Required Permission: None.
- Function: Get distinct plugin workflow type, host ID, operator, and status candidates under the specified filters.

### URL

POST /api/v3/plugin/workflow/distinct

### Request Parameters

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| exact_include_conditions | object | No | Exact match include conditions |
| fuzzy_include_conditions | object | No | Fuzzy match include conditions; currently empty |

#### exact_include_conditions

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| workflow_id | string array | No | Workflow ID list |
| type | string array | No | Plugin workflow type list. Available values: `install_plugin`, `upgrade_plugin`, `uninstall_plugin`, `reconfig_plugin`, `apply_plugin_subconfig`, `remove_plugin_subconfig`, `start_plugin`, `restart_plugin`, `stop_plugin`, `debug_plugin`, `stop_plugin_v2`, `migrate_plugin_v2` |
| bk_host_id | int64 array | No | Host ID list |
| status | string array | No | Workflow status list. Available values: `running`, `success`, `failed`, `partial_failed` |
| operator | string array | No | Operator list |
| bk_host_innerip | string array | No | Host inner IPv4 list |
| bk_host_innerip_v6 | string array | No | Host inner IPv6 list |
| bk_biz_id | int64 array | No | Business ID list |
| deploy_policy_id | int64 array | No | Deploy policy ID list |

#### fuzzy_include_conditions

The object is currently empty. Pass `{}` or omit it.

### Request Example

```json
{
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "type": ["upgrade_plugin"],
    "status": ["running"]
  },
  "fuzzy_include_conditions": {}
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "type": ["install_plugin", "upgrade_plugin"],
    "bk_host_id": [1001, 1002],
    "operator": ["admin", "ops"],
    "status": ["running", "success"]
  }
}
```

### Response Parameters

| Parameter | Type | Description |
| --- | --- | --- |
| code | int32 | Status code; `0` means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, usually empty on success |
| permission | object | Permission information |
| data | object | Response data |

#### data

| Parameter | Type | Description |
| --- | --- | --- |
| type | string array | Distinct plugin workflow types; enum values are the same as `exact_include_conditions.type` |
| bk_host_id | int64 array | Distinct host IDs |
| operator | string array | Distinct operators |
| status | string array | Distinct workflow statuses; enum values are the same as `exact_include_conditions.status` |

### Notes

- The current implementation always returns these four distinct fields and does not support selecting columns in the request.
- The current handler does not perform an IAM permission check.
