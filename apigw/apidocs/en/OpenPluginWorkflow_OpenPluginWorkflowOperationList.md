### Description

- API Version: v3.0.1-alpha.114+.
- Required Permission: system_access (System Access), plugin_history_view (View Plugin History).
- Function: Query operations under a specified plugin workflow with pagination, exact-match conditions, and count-only mode.

### URL

POST /api/v3/open/plugin/workflow/operation/list

### Request Parameters

| Parameter                | Type   | Required | Description                                                                 |
| ------------------------ | ------ | -------- | --------------------------------------------------------------------------- |
| only_count               | bool   | No       | Return only the total without operation details; defaults to `false`        |
| page                     | object | Yes      | Pagination settings; supply valid pagination for count-only queries as well |
| workflow_id              | string | Yes      | Plugin workflow ID; cannot be empty                                         |
| exact_include_conditions | object | No       | Exact-match inclusion conditions                                            |
| fuzzy_include_conditions | object | No       | Fuzzy-match inclusion conditions; currently an empty object                 |

#### page

| Parameter | Type  | Required | Description                                    |
| --------- | ----- | -------- | ---------------------------------------------- |
| offset    | int32 | No       | Nonnegative pagination offset; defaults to `0` |
| limit     | int32 | Yes      | Number of items per page, in `(0, 500]`        |

#### exact_include_conditions

| Parameter      | Type         | Required | Description                                                                                     |
| -------------- | ------------ | -------- | ----------------------------------------------------------------------------------------------- |
| bk_host_id     | int64 array  | No       | Host ID list                                                                                    |
| plugin_name    | string array | No       | Plugin name list                                                                                |
| plugin_version | string array | No       | Plugin version list                                                                             |
| state          | string array | No       | Operation states: `init`, `launched`, `running`, `success`, `failed`, `timeout`, `terminated`   |
| bk_biz_id      | int64 array  | No       | Business ID list; currently has no filtering effect and must not be relied on to narrow results |

#### fuzzy_include_conditions

Currently an empty object. Pass `{}` or omit it.

### Request Example

```json
{
  "only_count": false,
  "page": {
    "offset": 0,
    "limit": 20
  },
  "workflow_id": "wf-plugin-0001",
  "exact_include_conditions": {
    "bk_host_id": [1001],
    "plugin_name": ["bkmonitorbeat"],
    "plugin_version": ["3.6.0"],
    "state": ["running"]
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
        "operation_id": "op-plugin-0001",
        "instance_ids": ["inst-plugin-0001"],
        "operator": "admin",
        "create_time": 1719820000000,
        "plugin_deployment_info": {
          "bk_host_id": 1001,
          "bk_biz_id": 2,
          "bk_networkarea_id": 0,
          "bk_networkunit_id": 1,
          "bk_host_innerip_list": ["10.0.0.1"],
          "bk_host_innerip_v6_list": ["fe80::1"],
          "plugin_name": "bkmonitorbeat",
          "plugin_version": "3.6.0"
        },
        "latest_oper_inst_brief_data": {
          "life_cycle": {
            "state": "running",
            "create_time": 1719820000000,
            "start_time": 1719820010000,
            "end_time": 0
          },
          "latest_action_inst_brief_data": {
            "name": "upgrade_plugin",
            "tags": []
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
| code       | int32  | Status code; `0` means success              |
| message    | string | Request message                             |
| request_id | string | Request ID                                  |
| error      | object | Error information, usually empty on success |
| permission | object | Permission information                      |
| data       | object | Response data                               |

#### data

| Parameter  | Type  | Description                                |
| ---------- | ----- | ------------------------------------------ |
| operations | array | Operation list                             |
| total      | int64 | Total records matching the current filters |

#### data.operations[n]

| Parameter                   | Type         | Description                                                                  |
| --------------------------- | ------------ | ---------------------------------------------------------------------------- |
| operation_id                | string       | Operation ID                                                                 |
| instance_ids                | string array | Operation instance ID list                                                   |
| operator                    | string       | Operator                                                                     |
| create_time                 | int64        | Operation creation time as a Unix timestamp in milliseconds                  |
| plugin_deployment_info      | object       | Plugin deployment information                                                |
| latest_oper_inst_brief_data | object       | Latest operation instance summary; not provided when no summary is available |

#### data.operations[n].plugin_deployment_info

| Parameter               | Type         | Description                    |
| ----------------------- | ------------ | ------------------------------ |
| bk_host_id              | int64        | Host ID                        |
| bk_biz_id               | int64        | Business ID                    |
| bk_networkarea_id       | int64        | Network area ID                |
| bk_networkunit_id       | int64        | Network unit ID                |
| bk_host_innerip_list    | string array | Host private IPv4 address list |
| bk_host_innerip_v6_list | string array | Host private IPv6 address list |
| plugin_name             | string       | Plugin name                    |
| plugin_version          | string       | Plugin version                 |

#### data.operations[n].latest_oper_inst_brief_data

| Parameter                     | Type   | Description                    |
| ----------------------------- | ------ | ------------------------------ |
| life_cycle                    | object | Lifecycle summary              |
| latest_action_inst_brief_data | object | Latest action instance summary |

#### data.operations[n].latest_oper_inst_brief_data.life_cycle

| Parameter   | Type   | Description                                                                                  |
| ----------- | ------ | -------------------------------------------------------------------------------------------- |
| state       | string | Lifecycle state: `init`, `launched`, `running`, `success`, `failed`, `timeout`, `terminated` |
| create_time | int64  | Creation time as a Unix timestamp in milliseconds                                            |
| start_time  | int64  | Start time as a Unix timestamp in milliseconds; `0` when not recorded                        |
| end_time    | int64  | End time as a Unix timestamp in milliseconds; `0` when not recorded                          |
| stop_time   | int64  | Stop time as a Unix timestamp in milliseconds; currently not provided in this summary        |

#### data.operations[n].latest_oper_inst_brief_data.latest_action_inst_brief_data

| Parameter | Type         | Description                               |
| --------- | ------------ | ----------------------------------------- |
| name      | string       | Latest action name                        |
| tags      | string array | Latest action tags; may be an empty array |

### Processing Rules

- The current user and tenant come from the gateway authentication context. Request parameters cannot select or switch them.
- The workflow is located by `workflow_id`, and `plugin_history_view` permission is checked for the businesses associated with that workflow.
- `bk_biz_id` currently has no filtering effect. The fuzzy-condition object currently has no supported filter fields.
- `page.limit` must be in `(0, 500]`. Supply valid pagination even when `only_count=true`.
- When `only_count=true`, the response includes `data.total` and `data.operations` is an empty array.
- Custom sorting is not supported. Do not assume results are sorted by operation creation time.
