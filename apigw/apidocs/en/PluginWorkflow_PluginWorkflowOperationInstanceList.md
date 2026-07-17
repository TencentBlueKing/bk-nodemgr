### Description

- API Version: v3.0.1-alpha.1+.
- Version Changes: `v3.0.1-alpha.13+` added the permission response field; `v3.0.1-alpha.19+` added `plugin_operate` authorization.
- Required Permission: plugin_operate (Operate Plugin).
- Function: Query plugin workflow operation instances by operation ID list, with support for count-only mode.

### URL

POST /api/v3/plugin/workflow/operation/instance/list

### Request Parameters

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| only_count | bool | No | Whether to return only the count without operation instance details |
| operation_id | string array | Yes | Operation ID list; at least one element is required and the first element cannot be empty |

### Request Example

```json
{
  "only_count": false,
  "operation_id": ["op-plugin-0001"]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 1,
    "oper_inst_data": [
      {
        "operation_id": "op-plugin-0001",
        "oper_inst_id": "inst-plugin-0001",
        "oper_inst_status": "running",
        "operation_def_name": "upgrade_plugin",
        "parent_operation_id": "",
        "action_names": ["download_package", "upgrade_plugin"],
        "life_cycle": {
          "state": "running",
          "create_time": 1719820000000,
          "start_time": 1719820010000,
          "end_time": 0,
          "stop_time": 0
        }
      }
    ]
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
| total | int64 | Total operation instances matching the query |
| oper_inst_data | array | Operation instance brief data list |

#### data.oper_inst_data[n]

| Parameter | Type | Description |
| --- | --- | --- |
| operation_id | string | Operation ID |
| oper_inst_id | string | Operation instance ID |
| oper_inst_status | string | Operation instance status. Available values: `init`, `launched`, `running`, `success`, `failed`, `timeout`, `terminated` |
| operation_def_name | string | Operation definition name |
| parent_operation_id | string | Parent operation ID |
| action_names | string array | Action name list in the operation instance |
| life_cycle | object | Operation instance lifecycle information |
| latest_action_inst_brief_data | object | Latest action instance brief data; declared in the protocol but not populated by the current plugin workflow response converter |

#### data.oper_inst_data[n].life_cycle

| Parameter | Type | Description |
| --- | --- | --- |
| state | string | Lifecycle state; available values are the same as `oper_inst_status` |
| create_time | int64 | Creation time as a Unix timestamp in milliseconds |
| start_time | int64 | Start time as a Unix timestamp in milliseconds |
| end_time | int64 | End time as a Unix timestamp in milliseconds |
| stop_time | int64 | Stop time as a Unix timestamp in milliseconds |

#### data.oper_inst_data[n].latest_action_inst_brief_data

| Parameter | Type | Description |
| --- | --- | --- |
| name | string | Latest action name |
| tags | string array | Latest action tags |

### Processing Rules

- Request validation requires at least one `operation_id` element and checks that the first element is not empty.
- The backend uses the first operation ID to locate the owning workflow and check `plugin_operate` permission, then queries all operation IDs in the request.
- When `only_count=true`, the response includes `data.total` and `data.oper_inst_data` is an empty array.
- `latest_action_inst_brief_data` is declared in the shared Proto but is not set by the current plugin workflow response converter, so it is omitted from the success example.
