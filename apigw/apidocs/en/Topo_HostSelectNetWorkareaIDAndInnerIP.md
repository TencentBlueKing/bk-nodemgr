### Description

- API Version: v3.0.1-alpha.60+.
- Required Permission: None.
- Function: Select host network area IDs and inner IPv4 addresses in batches by host conditions.

### URL

POST /api/v3/topo/host/scenario/select_networkarea_id_and_inner_ip

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| exact_include_conditions | object | No | Exact include filter conditions |
| fuzzy_include_conditions | object | No | Fuzzy include filter conditions |
| exact_exclude_conditions | object | No | exact exclude conditions |

#### exact_include_conditions

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_host_id | int64 array | No | Host ID |
| bk_biz_id | int64 array | No | Business ID |
| bk_networkarea_id | int64 array | No | Network area ID |
| os_type | string array | No | Operating system type |
| node_role | string array | No | Node role |
| node_status | string array | No | node status |
| node_version | string array | No | Node version |
| bk_agent_id | string array | No | bk agent id |
| bk_networkunit_id | int64 array | No | Network unit ID |
| node_generation | int64 array | No | node generation |
| arch | string array | No | arch |
| proxy_tags | string array | No | proxy tags |
| bk_set_id | int64 array | No | bk set id |
| bk_module_id | int64 array | No | bk module id |

#### fuzzy_include_conditions

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_host_name | string array | No | bk host name |
| dept_name | string array | No | dept name |
| bk_host_innerip | string array | No | Host inner IPv4 address |
| bk_host_innerip_v6 | string array | No | Host inner IPv6 address |
| bk_host_outerip | string array | No | bk host outerip |
| bk_host_outerip_v6 | string array | No | bk host outerip v6 |

#### exact_exclude_conditions

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_host_id | int64 array | No | Host ID |
| bk_biz_id | int64 array | No | Business ID |
| bk_networkarea_id | int64 array | No | Network area ID |
| os_type | string array | No | Operating system type |
| node_role | string array | No | Node role |
| node_status | string array | No | node status |
| node_version | string array | No | Node version |
| bk_agent_id | string array | No | bk agent id |
| bk_networkunit_id | int64 array | No | Network unit ID |
| node_generation | int64 array | No | node generation |
| arch | string array | No | arch |
| proxy_tags | string array | No | proxy tags |
| bk_set_id | int64 array | No | bk set id |
| bk_module_id | int64 array | No | bk module id |

### Request Example

```json
{
  "exact_include_conditions": {
    "bk_host_id": [
      1
    ],
    "bk_biz_id": [
      1
    ],
    "bk_networkarea_id": [
      1
    ],
    "os_type": [
      "linux"
    ],
    "node_role": [
      "agent"
    ],
    "node_status": [
      "running"
    ],
    "node_version": [
      "3.2.1"
    ],
    "bk_agent_id": [
      "id-001"
    ]
  },
  "fuzzy_include_conditions": {
    "bk_host_name": [
      "default"
    ],
    "dept_name": [
      "default"
    ],
    "bk_host_innerip": [
      "10.0.0.1"
    ],
    "bk_host_innerip_v6": [
      "2001:db8::1"
    ],
    "bk_host_outerip": [
      "10.0.0.1"
    ],
    "bk_host_outerip_v6": [
      "2001:db8::1"
    ]
  },
  "exact_exclude_conditions": {
    "bk_host_id": [
      1
    ],
    "bk_biz_id": [
      1
    ],
    "bk_networkarea_id": [
      1
    ],
    "os_type": [
      "linux"
    ],
    "node_role": [
      "agent"
    ],
    "node_status": [
      "running"
    ],
    "node_version": [
      "3.2.1"
    ],
    "bk_agent_id": [
      "id-001"
    ]
  }
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "error": null,
  "permission": null,
  "data": {
    "items": [
      "string"
    ]
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, 0 means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, empty on success |
| permission | object | Permission information |
| data | object | Response data |

#### data

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| items | string array | Data list |
