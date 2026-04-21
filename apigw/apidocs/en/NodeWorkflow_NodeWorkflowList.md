### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: agent_view (View Agent), proxy_view (View Proxy).
- Function: Query node workflow list, with pagination, exact filtering, and operation time range filtering support.

### URL

POST /api/v3/node/workflow/list

### Request Parameters

| Parameter                | Type   | Required | Description                                                   |
| ------------------------ | ------ | -------- | ------------------------------------------------------------- |
| page                     | object | No       | Pagination configuration                                      |
| only_count               | bool   | No       | Whether to return only the total count, without item details  |
| exact_include_conditions | object | No       | Exact match include conditions                                |
| fuzzy_include_conditions | object | No       | Fuzzy match include conditions, currently no available fields |
| operate_time_range       | object | No       | Operation time range, in seconds                              |

#### page

| Parameter | Type  | Required | Description                                  |
| --------- | ----- | -------- | -------------------------------------------- |
| offset    | int32 | No       | Pagination start position, starting from `0` |
| limit     | int32 | No       | Records per page                             |

#### exact_include_conditions

| Parameter          | Type         | Required | Description                                                                                                                                                                                                          |
| ------------------ | ------------ | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| workflow_id        | string array | No       | Workflow ID list                                                                                                                                                                                                     |
| type               | string array | No       | Workflow type list, available values: `install_agent`, `install_proxy`, `upgrade_agent`, `upgrade_proxy`, `reconfig_agent`, `reconfig_proxy`, `restart_agent`, `restart_proxy`, `uninstall_agent`, `uninstall_proxy` |
| bk_biz_id          | int64 array  | No       | Business ID list                                                                                                                                                                                                     |
| status             | string array | No       | Workflow status list, available values: `running`, `success`, `failed`, `partial_failed`                                                                                                                             |
| operator           | string array | No       | Operator list                                                                                                                                                                                                        |
| bk_host_innerip    | string array | No       | Host inner IPv4 list                                                                                                                                                                                                 |
| bk_host_innerip_v6 | string array | No       | Host inner IPv6 list                                                                                                                                                                                                 |
| node_role          | string array | No       | Node role list, available values: `blank`, `agent`, `proxy`                                                                                                                                                          |

#### fuzzy_include_conditions

The object is currently empty, pass `{}` or omit it.

#### operate_time_range

| Parameter           | Type  | Required | Description                |
| ------------------- | ----- | -------- | -------------------------- |
| start_timestamp_sec | int64 | No       | Start timestamp in seconds |
| end_timestamp_sec   | int64 | No       | End timestamp in seconds   |

**Permission Notes**:

- When `node_role` is not provided, backend checks both `agent_view` and `proxy_view` by default for safety.
- When `node_role` is provided, backend narrows required permission actions by role.

### Request Example

Query workflows under business `2` with `upgrade_agent` type and `running` status.

```json
{
  "page": {
    "offset": 0,
    "limit": 20
  },
  "only_count": false,
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "type": ["upgrade_agent"],
    "status": ["running"],
    "node_role": ["agent"]
  },
  "fuzzy_include_conditions": {},
  "operate_time_range": {
    "start_timestamp_sec": 1717171200,
    "end_timestamp_sec": 1719859599
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
    "items": [
      {
        "workflow_id": "wf-20240601-0001",
        "trigger_id": "trigger-9f2b",
        "type": "upgrade_agent",
        "bk_biz_id": [2],
        "bk_networkarea_id": [0],
        "bk_networkunit_id": [1],
        "operator": "admin",
        "operate_time": 1719820000000,
        "finish_time": 1719820300000,
        "status": "running",
        "bk_biz_name": ["BlueKing"],
        "node_role": ["agent"]
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

| Parameter | Type  | Description                                      |
| --------- | ----- | ------------------------------------------------ |
| total     | int64 | Total number of records matching current filters |
| items     | array | Node workflow list                               |

#### data.items[n]

| Parameter         | Type         | Description                                                              |
| ----------------- | ------------ | ------------------------------------------------------------------------ |
| workflow_id       | string       | Workflow ID                                                              |
| trigger_id        | string       | Trigger ID                                                               |
| type              | string       | Workflow type, same enum values as `exact_include_conditions.type`       |
| bk_biz_id         | int64 array  | Business ID list                                                         |
| bk_networkarea_id | int64 array  | Network area ID list                                                     |
| bk_networkunit_id | int64 array  | Network unit ID list                                                     |
| operator          | string       | Operator                                                                 |
| operate_time      | int64        | Operation time, Unix timestamp in milliseconds                           |
| finish_time       | int64        | Finish time, Unix timestamp in milliseconds                              |
| status            | string       | Workflow status, same enum values as `exact_include_conditions.status`   |
| bk_biz_name       | string array | Business name list                                                       |
| node_role         | string array | Node role list, same enum values as `exact_include_conditions.node_role` |
