### Description

- API Version: v3.0.1-alpha.114+.
- Required Permissions: system_access (System Access), agent_history_view (View Agent History), proxy_history_view (View Proxy History). History permissions are checked against the target workflow's stored businesses and node roles: `blank` and `agent` require Agent history permission; `proxy` requires Proxy history permission. Both categories, missing roles, or unknown roles require both permissions.
- Function: Query execution instances and historical execution states by node operation ID, optionally filtered by instance IDs, with a count-only option.
- The API Gateway authenticates and forwards the user identity. Do not supply username or tenant parameters in the request body.

### URL

POST /api/v3/open/node/workflow/operation/instance/list

### Request Parameters

| Parameter    | Type         | Required | Description                                                                                   |
| ------------ | ------------ | -------- | --------------------------------------------------------------------------------------------- |
| only_count   | bool         | No       | Return only the total count; defaults to `false`                                              |
| operation_id | string array | Yes      | Non-empty operation ID list; use one operation ID per troubleshooting request                 |
| oper_inst_id | string array | No       | Further filter instances within the supplied operation scope; does not replace `operation_id` |

### Usage Limits

- Use one `operation_id` per troubleshooting request: obtain an operation ID from the operation list, then query instances for that operation. The `operation_id` parameter remains an array.
- Cross-workflow batch troubleshooting is not supported. Do not mix operations from different workflows in one request.
- This API has no pagination parameters. `only_count` controls whether instance details are returned; it does not provide pagination or sorting controls.

### Request Example

Query execution instances for a single operation.

```json
{
  "only_count": false,
  "operation_id": ["op-8a3d"]
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
        "operation_id": "op-8a3d",
        "oper_inst_id": "inst-001",
        "oper_inst_status": "running",
        "operation_def_name": "install_agent",
        "parent_operation_id": "",
        "action_names": ["DownloadPackage", "InstallAgent"],
        "life_cycle": {
          "state": "running",
          "create_time": 1719820000000,
          "start_time": 1719820010000,
          "end_time": 0,
          "stop_time": 0
        },
        "latest_action_inst_brief_data": {
          "name": "InstallAgent",
          "tags": []
        }
      }
    ]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                        |
| ---------- | ------ | ---------------------------------- |
| code       | int32  | Status code; `0` indicates success |
| message    | string | Request message                    |
| request_id | string | Request ID                         |
| error      | object | Error information                  |
| permission | object | Permission information             |
| data       | object | Response data                      |

#### data

| Parameter      | Type  | Description                                                |
| -------------- | ----- | ---------------------------------------------------------- |
| total          | int64 | Total matching operation instances                         |
| oper_inst_data | array | Operation instance summaries; empty when `only_count=true` |

#### data.oper_inst_data[n]

| Parameter                     | Type         | Description                                                         |
| ----------------------------- | ------------ | ------------------------------------------------------------------- |
| operation_id                  | string       | Parent operation ID                                                 |
| oper_inst_id                  | string       | Operation instance ID, used to query step logs                      |
| oper_inst_status              | string       | Operation instance state, such as `running`, `success`, or `failed` |
| operation_def_name            | string       | Operation definition name                                           |
| parent_operation_id           | string       | Parent operation ID in the operation hierarchy                      |
| action_names                  | string array | Names of actions in the instance                                    |
| life_cycle                    | object       | Operation instance lifecycle                                        |
| latest_action_inst_brief_data | object       | Latest action summary; may be an empty object                       |

#### data.oper_inst_data[n].life_cycle

| Parameter   | Type   | Description                        |
| ----------- | ------ | ---------------------------------- |
| state       | string | Operation instance lifecycle state |
| create_time | int64  | Creation time in Unix milliseconds |
| start_time  | int64  | Start time in Unix milliseconds    |
| end_time    | int64  | End time in Unix milliseconds      |
| stop_time   | int64  | Stop time in Unix milliseconds     |

Lifecycle timestamps that have not been recorded are `0`.

#### data.oper_inst_data[n].latest_action_inst_brief_data

| Parameter | Type         | Description |
| --------- | ------------ | ----------- |
| name      | string       | Action name |
| tags      | string array | Action tags |
