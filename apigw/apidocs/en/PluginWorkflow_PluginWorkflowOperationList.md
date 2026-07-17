### Description

- API Version: v3.0.1-alpha.1+.
- Version Changes: `v3.0.1-alpha.6+` renamed the count field to `total`; `v3.0.1-alpha.13+` added the permission response field; `v3.0.1-alpha.18+` added the business ID condition field; `v3.0.1-alpha.19+` added `plugin_operate` authorization.
- Required Permission: plugin_operate (Operate Plugin).
- Function: Query operations under a specified plugin workflow, with pagination, exact filters, and count-only mode.

### URL

POST /api/v3/plugin/workflow/operation/list

### Request Parameters

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| only_count | bool | No | Whether to return only the count without operation details |
| page | object | No | Pagination configuration |
| workflow_id | string | Yes | Plugin workflow ID; cannot be empty |
| exact_include_conditions | object | No | Exact match include conditions |
| fuzzy_include_conditions | object | No | Fuzzy match include conditions; currently empty |

#### page

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| offset | int32 | No | Pagination start position, starting from `0` |
| limit | int32 | No | Records per page; maximum value is `500` |

#### exact_include_conditions

| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| bk_host_id | int64 array | No | Host ID list |
| plugin_name | string array | No | Plugin name list |
| plugin_version | string array | No | Plugin version list |
| state | string array | No | Operation state list. Available values: `init`, `launched`, `running`, `success`, `failed`, `timeout`, `terminated` |
| bk_biz_id | int64 array | No | Business ID list; declared in the protocol but not used for filtering by the current backend converter |

#### fuzzy_include_conditions

The object is currently empty. Pass `{}` or omit it.

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
            "tags": ["main"]
          }
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
| operations | array | Operation list |
| total | int64 | Total records matching the current filters |

#### data.operations[n]

| Parameter | Type | Description |
| --- | --- | --- |
| operation_id | string | Operation ID |
| instance_ids | string array | Operation instance ID list |
| operator | string | Operator |
| create_time | int64 | Operation creation time as a Unix timestamp in milliseconds |
| plugin_deployment_info | object | Plugin deployment information |
| latest_oper_inst_brief_data | object | Latest operation instance brief data |

#### data.operations[n].plugin_deployment_info

| Parameter | Type | Description |
| --- | --- | --- |
| bk_host_id | int64 | Host ID |
| bk_biz_id | int64 | Business ID |
| bk_networkarea_id | int64 | Network area ID |
| bk_networkunit_id | int64 | Network unit ID |
| bk_host_innerip_list | string array | Host inner IPv4 list |
| bk_host_innerip_v6_list | string array | Host inner IPv6 list |
| plugin_name | string | Plugin name |
| plugin_version | string | Plugin version |

#### data.operations[n].latest_oper_inst_brief_data

| Parameter | Type | Description |
| --- | --- | --- |
| life_cycle | object | Lifecycle brief data |
| latest_action_inst_brief_data | object | Latest action instance brief data |

#### data.operations[n].latest_oper_inst_brief_data.life_cycle

| Parameter | Type | Description |
| --- | --- | --- |
| state | string | Lifecycle state. Available values: `init`, `launched`, `running`, `success`, `failed`, `timeout`, `terminated` |
| create_time | int64 | Creation time as a Unix timestamp in milliseconds |
| start_time | int64 | Start time as a Unix timestamp in milliseconds |
| end_time | int64 | End time as a Unix timestamp in milliseconds |
| stop_time | int64 | Stop time as a Unix timestamp in milliseconds; declared in the protocol but not populated by the current backend converter |

#### data.operations[n].latest_oper_inst_brief_data.latest_action_inst_brief_data

| Parameter | Type | Description |
| --- | --- | --- |
| name | string | Latest action name |
| tags | string array | Latest action tags |

### Processing Rules

- The backend queries operations using the workflow-linked `trigger_id`, then joins plugin deployment information by operation token.
- `bk_biz_id` is declared in the request Proto, but the current converter does not pass it to operation or plugin deployment query conditions, so it currently has no filtering effect.
- When `only_count=true`, the response includes `data.total` and `data.operations` is an empty array.
- `life_cycle.stop_time` for the latest operation instance is declared in the shared Proto but is not set by the current response converter, so it is omitted from the success example.
