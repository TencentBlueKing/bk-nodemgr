### Description

- API Version: v3.0.1-alpha.114+.
- Required Permission: system_access (System Access), plugin_history_view (View Plugin History).
- Function: Query plugin workflow operation instances by operation IDs, with count-only mode. A single operation ID per request is recommended to locate instances.

### URL

POST /api/v3/open/plugin/workflow/operation/instance/list

### Request Parameters

| Parameter    | Type         | Required | Description                                                                                                                      |
| ------------ | ------------ | -------- | -------------------------------------------------------------------------------------------------------------------------------- |
| only_count   | bool         | No       | Return only the total without operation instance details; defaults to `false`                                                    |
| operation_id | string array | Yes      | Operation ID list with at least one element; the first element cannot be empty. A single operation ID per request is recommended |

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
        "action_names": [
          "fetch_plugin_process",
          "verify_plugin_availability",
          "check_plugin_process_alive",
          "stop_process",
          "inject_plugin_custom_deploy_config",
          "inject_main_config_detail",
          "render_plugin_deployment",
          "fetch_process_sub_config_into_deployment",
          "ensure_and_update_plugin_config_details",
          "render_plugin_config",
          "overwrite_plugin_config_for_compatibility",
          "transfer_plugin_pkg_to_node",
          "upgrade_plugin",
          "wait_plugin_installer_complete",
          "push_plugin_config",
          "start_process",
          "update_process"
        ],
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

| Parameter  | Type   | Description                                 |
| ---------- | ------ | ------------------------------------------- |
| code       | int32  | Status code; `0` means success              |
| message    | string | Request message                             |
| request_id | string | Request ID                                  |
| error      | object | Error information, usually empty on success |
| permission | object | Permission information                      |
| data       | object | Response data                               |

#### data

| Parameter      | Type  | Description                                  |
| -------------- | ----- | -------------------------------------------- |
| total          | int64 | Total operation instances matching the query |
| oper_inst_data | array | Operation instance brief data list           |

#### data.oper_inst_data[n]

| Parameter                     | Type         | Description                                                                                           |
| ----------------------------- | ------------ | ----------------------------------------------------------------------------------------------------- |
| operation_id                  | string       | Operation ID                                                                                          |
| oper_inst_id                  | string       | Operation instance ID                                                                                 |
| oper_inst_status              | string       | Operation instance state: `init`, `launched`, `running`, `success`, `failed`, `timeout`, `terminated` |
| operation_def_name            | string       | Operation definition name                                                                             |
| parent_operation_id           | string       | Parent operation ID; empty when there is no parent operation                                          |
| action_names                  | string array | Action names in the order specified by the operation definition                                       |
| life_cycle                    | object       | Operation instance lifecycle information                                                              |
| latest_action_inst_brief_data | object       | Latest action instance summary; currently not provided in this response                               |

#### data.oper_inst_data[n].life_cycle

| Parameter   | Type   | Description                                                           |
| ----------- | ------ | --------------------------------------------------------------------- |
| state       | string | Lifecycle state; values are the same as `oper_inst_status`            |
| create_time | int64  | Creation time as a Unix timestamp in milliseconds                     |
| start_time  | int64  | Start time as a Unix timestamp in milliseconds; `0` when not recorded |
| end_time    | int64  | End time as a Unix timestamp in milliseconds; `0` when not recorded   |
| stop_time   | int64  | Stop time as a Unix timestamp in milliseconds; `0` when not recorded  |

#### data.oper_inst_data[n].latest_action_inst_brief_data

| Parameter | Type         | Description        |
| --------- | ------------ | ------------------ |
| name      | string       | Latest action name |
| tags      | string array | Latest action tags |

### Processing Rules

- The current user and tenant come from the gateway authentication context. Request parameters cannot select or switch them.
- `operation_id` must contain at least one element, and the first element cannot be empty. Pass one `operation_id` per troubleshooting request. Cross-workflow batch troubleshooting is not supported.
- When `only_count=true`, the response includes `data.total` and `data.oper_inst_data` is an empty array.
- This API has no pagination, condition-filtering, or sorting parameters. The order of returned operation instances is not guaranteed.
- Use `action_names` to order action logs returned by the log query API, rather than relying on the order of object keys in the log response.
