### Description

- API Version: v3.0.1-alpha.114+.
- Required Permission: system_access (System Access), plugin_history_view (View Plugin History).
- Function: Query plugin workflows with pagination, exact-match conditions, and an operation time range.

### URL

POST /api/v3/open/plugin/workflow/list

### Request Parameters

| Parameter                | Type   | Required | Description                                                                          |
| ------------------------ | ------ | -------- | ------------------------------------------------------------------------------------ |
| page                     | object | No       | Pagination settings; valid pagination is required when `only_count=false` or omitted |
| only_count               | bool   | No       | Return only the total without workflow details; defaults to `false`                  |
| exact_include_conditions | object | Yes      | Exact-match inclusion conditions; pass `{}` when no filters are needed               |
| fuzzy_include_conditions | object | No       | Fuzzy-match inclusion conditions; currently an empty object                          |
| operate_time_range       | object | No       | Operation time range as Unix timestamps in seconds                                   |

#### page

| Parameter | Type  | Required | Description                                    |
| --------- | ----- | -------- | ---------------------------------------------- |
| offset    | int32 | No       | Nonnegative pagination offset; defaults to `0` |
| limit     | int32 | Yes      | Number of items per page, in `(0, 500]`        |

#### exact_include_conditions

| Parameter          | Type         | Required | Description                                                                                                                                                                                                                                                                                                                                                                                                                   |
| ------------------ | ------------ | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| workflow_id        | string array | No       | Workflow ID list                                                                                                                                                                                                                                                                                                                                                                                                              |
| type               | string array | No       | Plugin workflow types: `install_plugin`, `install_plugin_no_start`, `upgrade_plugin`, `uninstall_plugin`, `reconfig_plugin`, `apply_plugin_config`, `apply_plugin_subconfig`, `compose_plugin_subconfig`, `compose_projected_plugin_subconfigs`, `remove_plugin_subconfig`, `start_plugin`, `restart_plugin`, `stop_plugin`, `debug_plugin`, `stop_plugin_v2`, `ensure_plugin_v2`, `uninstall_plugin_v2`, `migrate_plugin_v2` |
| bk_host_id         | int64 array  | No       | Host ID list                                                                                                                                                                                                                                                                                                                                                                                                                  |
| status             | string array | No       | Workflow statuses: `running`, `success`, `failed`, `partial_failed`                                                                                                                                                                                                                                                                                                                                                           |
| operator           | string array | No       | Operators recorded in workflow history; does not change the current request identity                                                                                                                                                                                                                                                                                                                                          |
| bk_host_innerip    | string array | No       | Host private IPv4 address list                                                                                                                                                                                                                                                                                                                                                                                                |
| bk_host_innerip_v6 | string array | No       | Host private IPv6 address list                                                                                                                                                                                                                                                                                                                                                                                                |
| bk_biz_id          | int64 array  | No       | Business ID list                                                                                                                                                                                                                                                                                                                                                                                                              |
| deploy_policy_id   | int64 array  | No       | Deploy policy ID list                                                                                                                                                                                                                                                                                                                                                                                                         |

#### fuzzy_include_conditions

Currently an empty object. Pass `{}` or omit it.

#### operate_time_range

| Parameter           | Type  | Required | Description                                                                                            |
| ------------------- | ----- | -------- | ------------------------------------------------------------------------------------------------------ |
| start_timestamp_sec | int64 | No       | Nonnegative start timestamp in seconds; defaults to `0`                                                |
| end_timestamp_sec   | int64 | No       | Nonnegative end timestamp in seconds; defaults to `0` and must not be earlier than the start timestamp |

### Request Example

```json
{
  "page": {
    "offset": 0,
    "limit": 20
  },
  "only_count": false,
  "exact_include_conditions": {
    "workflow_id": ["wf-plugin-0001"]
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
        "workflow_id": "wf-plugin-0001",
        "trigger_id": "trigger-plugin-0001",
        "type": "upgrade_plugin",
        "bk_host_id": [1001],
        "operator": "admin",
        "operate_time": 1719820000000,
        "finish_time": 0,
        "status": "running",
        "bk_biz_id": [2],
        "deploy_policy_ids": [10001]
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

| Parameter | Type  | Description                                |
| --------- | ----- | ------------------------------------------ |
| total     | int64 | Total records matching the current filters |
| items     | array | Plugin workflow list                       |

#### data.items[n]

| Parameter         | Type        | Description                                                                                        |
| ----------------- | ----------- | -------------------------------------------------------------------------------------------------- |
| tenant_id         | string      | Tenant ID; currently not provided in the response and must not be relied on to identify the tenant |
| workflow_id       | string      | Workflow ID                                                                                        |
| trigger_id        | string      | Trigger ID                                                                                         |
| type              | string      | Plugin workflow type; values are the same as `exact_include_conditions.type`                       |
| bk_host_id        | int64 array | Host ID list                                                                                       |
| operator          | string      | Operator                                                                                           |
| operate_time      | int64       | Operation time as a Unix timestamp in milliseconds                                                 |
| finish_time       | int64       | Finish time as a Unix timestamp in milliseconds; `0` when not recorded                             |
| status            | string      | Workflow status; values are the same as `exact_include_conditions.status`                          |
| bk_biz_id         | int64 array | Business ID list                                                                                   |
| deploy_policy_ids | int64 array | Deploy policy ID list                                                                              |

### Processing Rules

- The current user and tenant come from the gateway authentication context. Request parameters cannot select or switch them; `operator` only filters historical records.
- Results are restricted to businesses the current user can query with `plugin_history_view` permission.
- Pass `exact_include_conditions: {}` even when no exact-match filters are needed.
- When `only_count=false` or omitted, `page.limit` must be in `(0, 500]`. When `only_count=true`, `page` is not required, `data.total` is returned, and `data.items` is an empty array.
- Workflows are returned in descending operation time order. Custom sorting is not supported, and the fuzzy-condition object currently has no supported filter fields.
- Request time ranges use seconds; response timestamps use milliseconds.
