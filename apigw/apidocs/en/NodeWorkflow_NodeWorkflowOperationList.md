### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: agent_operate (Operate Agent), proxy_operate (Operate Proxy).
  - Permission Note: Dynamically narrowed based on node roles in the workflow. If the workflow only involves Agent nodes, only agent_operate is required; if only Proxy nodes, only proxy_operate is required; if both or unknown roles, both are required.
- Function: Query operation list under a specific node workflow, with pagination, exact filters on deployment dimensions, and operation state filters.

### URL

POST /api/v3/node/workflow/operation/list

### Request Parameters

| Parameter                | Type   | Required | Description                                                   |
| ------------------------ | ------ | -------- | ------------------------------------------------------------- |
| only_count               | bool   | No       | Whether to return only total count, without detail records    |
| page                     | object | No       | Pagination configuration                                      |
| workflow_id              | string | Yes      | Node workflow ID                                              |
| exact_include_conditions | object | No       | Exact match include conditions                                |
| fuzzy_include_conditions | object | No       | Fuzzy match include conditions, currently no available fields |

#### page

| Parameter | Type  | Required | Description                                  |
| --------- | ----- | -------- | -------------------------------------------- |
| offset    | int32 | No       | Pagination start position, starting from `0` |
| limit     | int32 | No       | Records per page                             |

#### exact_include_conditions

| Parameter          | Type         | Required | Description                                                                                                         |
| ------------------ | ------------ | -------- | ------------------------------------------------------------------------------------------------------------------- |
| node_version       | string array | No       | Node version list                                                                                                   |
| bk_host_innerip    | string array | No       | Host inner IPv4 list                                                                                                |
| bk_host_innerip_v6 | string array | No       | Host inner IPv6 list                                                                                                |
| bk_biz_id          | int64 array  | No       | Business ID list                                                                                                    |
| bk_networkarea_id  | int64 array  | No       | Network area ID list                                                                                                |
| bk_networkunit_id  | int64 array  | No       | Network unit ID list                                                                                                |
| state              | string array | No       | Operation state list, available values: `init`, `launched`, `running`, `success`, `failed`, `timeout`, `terminated` |

#### fuzzy_include_conditions

The object is currently empty, pass `{}` or omit it.

### Request Example

Query operations under workflow `wf-20240601-0001` with state `running` and business `2`.

```json
{
  "only_count": false,
  "page": {
    "offset": 0,
    "limit": 20
  },
  "workflow_id": "wf-20240601-0001",
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "state": ["running"],
    "node_version": ["2.1.0"]
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
    "total": 1,
    "operations": [
      {
        "operation_id": "op-8a3d",
        "instance_ids": ["inst-001", "inst-002"],
        "operator": "admin",
        "create_time": 1719820000000,
        "node_deployment_info": {
          "bk_host_id": 1001,
          "bk_biz_id": 2,
          "bk_host_innerip_list": ["10.0.0.1"],
          "bk_host_innerip_v6_list": ["fe80::1"],
          "bk_networkarea_id": 0,
          "bk_networkunit_id": 1,
          "node_version": "2.1.0"
        },
        "latest_oper_inst_brief_data": {
          "life_cycle": {
            "state": "running",
            "create_time": 1719820000000,
            "start_time": 1719820010000,
            "end_time": 0,
            "stop_time": 0
          },
          "latest_action_inst_brief_data": {
            "name": "install_agent",
            "tags": ["main"]
          }
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

#### error

| Parameter | Type   | Description             |
| --------- | ------ | ----------------------- |
| system    | string | Error system identifier |
| message   | string | Error message           |
| details   | array  | Error detail list       |

#### error.details[n]

| Parameter | Type   | Description   |
| --------- | ------ | ------------- |
| code      | string | Error code    |
| message   | string | Error message |

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

#### data

| Parameter  | Type  | Description                            |
| ---------- | ----- | -------------------------------------- |
| total      | int64 | Total records matching current filters |
| operations | array | Operation list                         |

#### data.operations[n]

| Parameter                   | Type         | Description                                           |
| --------------------------- | ------------ | ----------------------------------------------------- |
| operation_id                | string       | Operation ID                                          |
| instance_ids                | string array | Operation instance ID list                            |
| operator                    | string       | Operator                                              |
| create_time                 | int64        | Operation create time, Unix timestamp in milliseconds |
| node_deployment_info        | object       | Node deployment information                           |
| latest_oper_inst_brief_data | object       | Latest operation instance brief data                  |

#### data.operations[n].node_deployment_info

| Parameter               | Type         | Description          |
| ----------------------- | ------------ | -------------------- |
| bk_host_id              | int64        | Host ID              |
| bk_biz_id               | int64        | Business ID          |
| bk_host_innerip_list    | string array | Host inner IPv4 list |
| bk_host_innerip_v6_list | string array | Host inner IPv6 list |
| bk_networkarea_id       | int64        | Network area ID      |
| bk_networkunit_id       | int64        | Network unit ID      |
| node_version            | string       | Node version         |

#### data.operations[n].latest_oper_inst_brief_data

| Parameter                     | Type   | Description                       |
| ----------------------------- | ------ | --------------------------------- |
| life_cycle                    | object | Lifecycle brief data              |
| latest_action_inst_brief_data | object | Latest action instance brief data |

#### data.operations[n].latest_oper_inst_brief_data.life_cycle

| Parameter   | Type   | Description                                 |
| ----------- | ------ | ------------------------------------------- |
| state       | string | State value                                 |
| create_time | int64  | Create time, Unix timestamp in milliseconds |
| start_time  | int64  | Start time, Unix timestamp in milliseconds  |
| end_time    | int64  | End time, Unix timestamp in milliseconds    |
| stop_time   | int64  | Stop time, Unix timestamp in milliseconds   |

#### data.operations[n].latest_oper_inst_brief_data.latest_action_inst_brief_data

| Parameter | Type         | Description        |
| --------- | ------------ | ------------------ |
| name      | string       | Latest action name |
| tags      | string array | Latest action tags |

### Processing Rules

- `workflow_id` is required and cannot be empty.
- The query first filters operations by the workflow-linked `trigger_id`, then joins deployment records by operation token.
- When `only_count=true`, only `data.total` is returned, and `data.operations` is an empty array.
