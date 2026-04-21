### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: agent_operate (Operate Agent), proxy_operate (Operate Proxy).
  - Permission Note: Dynamically narrowed based on node roles in the workflow. If the workflow only involves Agent nodes, only agent_operate is required; if only Proxy nodes, only proxy_operate is required; if both or unknown roles, both are required.
- Function: Query node workflow operation instances by operation IDs or operation instance IDs, with support for count-only mode.

### URL

POST /api/v3/node/workflow/operation/instance/list

### Request Parameters

| Parameter    | Type         | Required | Description                                                             |
| ------------ | ------------ | -------- | ----------------------------------------------------------------------- |
| only_count   | bool         | No       | Whether to return count only, when `true` only `data.total` is returned |
| operation_id | string array | No       | Operation ID list, filters instances by operation dimension             |
| oper_inst_id | string array | No       | Operation instance ID list, filters instances by instance dimension     |

#### Query Parameter Notes

Callers can combine the following filters:

| Parameter    | Type         | Description                                |
| ------------ | ------------ | ------------------------------------------ |
| operation_id | string array | Filter instances under specific operations |
| oper_inst_id | string array | Filter specific operation instances        |
| only_count   | bool         | Return count only without detail list      |

### Request Example

Query operation instances under a specific operation and return details.

```json
{
  "only_count": false,
  "operation_id": ["op-20260421-0001"],
  "oper_inst_id": []
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260421-001",
  "data": {
    "total": 1,
    "oper_inst_data": [
      {
        "operation_id": "op-20260421-0001",
        "oper_inst_id": "inst-20260421-0001",
        "oper_inst_status": "running",
        "operation_def_name": "install_agent",
        "parent_operation_id": "",
        "action_names": ["DownloadPackage", "InstallAgent"],
        "life_cycle": {
          "state": "running",
          "create_time": 1718670000000,
          "start_time": 1718670010000,
          "end_time": 0,
          "stop_time": 0
        },
        "latest_action_inst_brief_data": {
          "name": "InstallAgent",
          "tags": ["agent", "linux"]
        }
      }
    ]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                 |
| ---------- | ------ | ------------------------------------------- |
| code       | int32  | Status code, `0` means success              |
| message    | string | Request message                             |
| request_id | string | Request ID                                  |
| error      | object | Error information, usually empty on success |
| permission | object | Permission information                      |
| data       | object | Response data                               |

#### data

| Parameter      | Type  | Description                                      |
| -------------- | ----- | ------------------------------------------------ |
| total          | int64 | Total number of records matching current filters |
| oper_inst_data | array | Operation instance brief data list               |

#### data.oper_inst_data[n]

| Parameter                     | Type         | Description                              |
| ----------------------------- | ------------ | ---------------------------------------- |
| operation_id                  | string       | Operation ID                             |
| oper_inst_id                  | string       | Operation instance ID                    |
| oper_inst_status              | string       | Operation instance status                |
| operation_def_name            | string       | Operation definition name                |
| parent_operation_id           | string       | Parent operation ID                      |
| action_names                  | string array | Action name list in this instance        |
| life_cycle                    | object       | Instance lifecycle information           |
| latest_action_inst_brief_data | object       | Brief info of the latest action instance |

#### data.oper_inst_data[n].life_cycle

| Parameter   | Type   | Description                                   |
| ----------- | ------ | --------------------------------------------- |
| state       | string | Lifecycle state                               |
| create_time | int64  | Creation time, Unix timestamp in milliseconds |
| start_time  | int64  | Start time, Unix timestamp in milliseconds    |
| end_time    | int64  | End time, Unix timestamp in milliseconds      |
| stop_time   | int64  | Stop time, Unix timestamp in milliseconds     |

#### data.oper_inst_data[n].latest_action_inst_brief_data

| Parameter | Type         | Description            |
| --------- | ------------ | ---------------------- |
| name      | string       | Latest action name     |
| tags      | string array | Latest action tag list |

### Notes

- When `only_count=true`, the response returns `data.total`, and `data.oper_inst_data` is empty or omitted.
