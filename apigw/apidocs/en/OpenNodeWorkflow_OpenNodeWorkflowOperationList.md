### Description

- API Version: v3.0.1-alpha.114+.
- Required Permissions: system_access (System Access), agent_history_view (View Agent History), proxy_history_view (View Proxy History). History permissions are checked against the target workflow's stored businesses and node roles: `blank` and `agent` require Agent history permission; `proxy` requires Proxy history permission. Both categories, missing roles, or unknown roles require both permissions.
- Function: Query a node workflow's operations, node deployment information, and latest execution instance summaries, with pagination and exact filters.
- The API Gateway authenticates and forwards the user identity. Do not supply username or tenant parameters in the request body.

### URL

POST /api/v3/open/node/workflow/operation/list

### Request Parameters

| Parameter                | Type   | Required | Description                                                                  |
| ------------------------ | ------ | -------- | ---------------------------------------------------------------------------- |
| workflow_id              | string | Yes      | Non-empty node workflow ID                                                   |
| only_count               | bool   | No       | Return only the total count; defaults to `false`                             |
| page                     | object | Yes      | Pagination; provide valid pagination parameters even for count-only requests |
| exact_include_conditions | object | No       | Exact-match include conditions                                               |
| fuzzy_include_conditions | object | No       | No supported fields; omit or pass `{}`                                       |

#### page

| Parameter | Type  | Required | Description                                            |
| --------- | ----- | -------- | ------------------------------------------------------ |
| offset    | int32 | No       | Starting offset, defaults to `0`; must be non-negative |
| limit     | int32 | Yes      | Records per page, from `1` to `500`                    |

#### exact_include_conditions

| Parameter          | Type         | Required | Description                                                 |
| ------------------ | ------------ | -------- | ----------------------------------------------------------- |
| node_version       | string array | No       | Node versions                                               |
| bk_host_innerip    | string array | No       | Host inner IPv4 addresses                                   |
| bk_host_innerip_v6 | string array | No       | Host inner IPv6 addresses                                   |
| bk_biz_id          | int64 array  | No       | Business IDs                                                |
| bk_networkarea_id  | int64 array  | No       | Network area IDs                                            |
| bk_networkunit_id  | int64 array  | No       | Network unit IDs                                            |
| state              | string array | No       | Operation states, such as `running`, `success`, or `failed` |

### Request Example

Query running operations in business `2` under a specific workflow.

```json
{
  "workflow_id": "wf-20240601-0001",
  "only_count": false,
  "page": { "offset": 0, "limit": 20 },
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "state": ["running"]
  }
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
        "instance_ids": ["inst-001"],
        "operator": "admin",
        "create_time": 1719820000000,
        "node_deployment_info": {
          "bk_host_id": 1001,
          "bk_biz_id": 2,
          "bk_host_innerip_list": ["10.0.0.1"],
          "bk_host_innerip_v6_list": [],
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
            "name": "InstallAgent",
            "tags": []
          }
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

| Parameter  | Type  | Description                                                       |
| ---------- | ----- | ----------------------------------------------------------------- |
| total      | int64 | Total matching records after filtering, not the current page size |
| operations | array | Current page of operations; empty when `only_count=true`          |

#### data.operations[n]

| Parameter                   | Type         | Description                                                       |
| --------------------------- | ------------ | ----------------------------------------------------------------- |
| operation_id                | string       | Operation ID, used to query its execution instances               |
| instance_ids                | string array | Instance IDs associated with this operation                       |
| operator                    | string       | Operator                                                          |
| create_time                 | int64        | Operation creation time in Unix milliseconds                      |
| node_deployment_info        | object       | Node deployment information associated with the operation         |
| latest_oper_inst_brief_data | object       | Latest operation instance summary; may be absent if not available |

#### data.operations[n].node_deployment_info

| Parameter               | Type         | Description               |
| ----------------------- | ------------ | ------------------------- |
| bk_host_id              | int64        | Host ID                   |
| bk_biz_id               | int64        | Business ID               |
| bk_host_innerip_list    | string array | Host inner IPv4 addresses |
| bk_host_innerip_v6_list | string array | Host inner IPv6 addresses |
| bk_networkarea_id       | int64        | Network area ID           |
| bk_networkunit_id       | int64        | Network unit ID           |
| node_version            | string       | Node version              |

#### data.operations[n].latest_oper_inst_brief_data

| Parameter                     | Type   | Description                                   |
| ----------------------------- | ------ | --------------------------------------------- |
| life_cycle                    | object | Latest operation instance lifecycle           |
| latest_action_inst_brief_data | object | Latest action summary; may be an empty object |

#### data.operations[n].latest_oper_inst_brief_data.life_cycle

| Parameter   | Type   | Description                        |
| ----------- | ------ | ---------------------------------- |
| state       | string | Operation instance state           |
| create_time | int64  | Creation time in Unix milliseconds |
| start_time  | int64  | Start time in Unix milliseconds    |
| end_time    | int64  | End time in Unix milliseconds      |
| stop_time   | int64  | Stop time in Unix milliseconds     |

Lifecycle timestamps that have not been recorded are `0`.

#### data.operations[n].latest_oper_inst_brief_data.latest_action_inst_brief_data

| Parameter | Type         | Description |
| --------- | ------------ | ----------- |
| name      | string       | Action name |
| tags      | string array | Action tags |
