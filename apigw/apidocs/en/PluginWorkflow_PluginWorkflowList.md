### Description

- API Version: v3.0.1-alpha.1+.
- Version Changes: `v3.0.1-alpha.13+` added the permission response field; `v3.0.1-alpha.18+` added host IP and business ID filters; `v3.0.1-alpha.19+` added `plugin_history_view` authorization.
- Required Permission: plugin_history_view (View Plugin History).
- Function: Query plugin workflows with pagination, exact filters, and an operation time range.

### URL

POST /api/v3/plugin/workflow/list

### Request Parameters

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| page | object | No | Pagination configuration |
| only_count | bool | No | Whether to return only the count without workflow details |
| exact_include_conditions | object | No | Exact match include conditions |
| fuzzy_include_conditions | object | No | Fuzzy match include conditions; currently empty |
| operate_time_range | object | No | Operation time range in seconds |

#### page

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| offset | int32 | No | Pagination start position, starting from `0` |
| limit | int32 | No | Records per page; maximum value is `500` |

#### exact_include_conditions

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| workflow_id | string array | No | Workflow ID list |
| type | string array | No | Plugin workflow type list. Available values: `install_plugin`, `upgrade_plugin`, `uninstall_plugin`, `reconfig_plugin`, `apply_plugin_subconfig`, `restart_plugin`, `stop_plugin`, `stop_plugin_v2`, `ensure_plugin_v2`, `uninstall_plugin_v2`, `migrate_plugin_v2` |
| bk_host_id | int64 array | No | Host ID list |
| status | string array | No | Workflow status list. Available values: `running`, `success`, `failed`, `partial_failed` |
| operator | string array | No | Operator list |
| bk_host_innerip | string array | No | Host inner IPv4 list |
| bk_host_innerip_v6 | string array | No | Host inner IPv6 list |
| bk_biz_id | int64 array | No | Business ID list |

#### fuzzy_include_conditions

The object is currently empty. Pass `{}` or omit it.

#### operate_time_range

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| start_timestamp_sec | int64 | No | Start timestamp in seconds |
| end_timestamp_sec | int64 | No | End timestamp in seconds |

### Request Example

```json
{
  "page": {
    "offset": 0,
    "limit": 20
  },
  "only_count": false,
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "type": ["upgrade_plugin"],
    "status": ["running"]
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
        "workflow_id": "wf-plugin-0001",
        "trigger_id": "trigger-plugin-0001",
        "type": "upgrade_plugin",
        "bk_host_id": [1001],
        "operator": "admin",
        "operate_time": 1719820000000,
        "finish_time": 1719820300000,
        "status": "running",
        "bk_biz_id": [2]
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
| total | int64 | Total records matching the current filters |
| items | array | Plugin workflow list |

#### data.items[n]

| Parameter | Type | Description |
| --- | --- | --- |
| tenant_id | string | Tenant ID; reserved by the protocol and not populated by the current backend converter |
| workflow_id | string | Workflow ID |
| trigger_id | string | Trigger ID |
| type | string | Plugin workflow type; enum values are the same as `exact_include_conditions.type` |
| bk_host_id | int64 array | Host ID list |
| operator | string | Operator |
| operate_time | int64 | Operation time as a Unix timestamp in milliseconds |
| finish_time | int64 | Finish time as a Unix timestamp in milliseconds |
| status | string | Workflow status; enum values are the same as `exact_include_conditions.status` |
| bk_biz_id | int64 array | Business ID list |

### Processing Rules

- The backend narrows the queryable business scope according to `plugin_history_view` permission.
- When `only_count=true`, the response includes `data.total` and `data.items` is an empty array.
- `tenant_id` is declared in the Proto but is not set by the current response converter, so it is omitted from the success example.
