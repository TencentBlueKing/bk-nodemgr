### Description

- API Version: v3.0.1+.
- Required Permission: None.
- Function: Count the host distribution for each network area based on host filtering conditions.

### URL

POST /api/v3/topo/host/get_host_distribution_by_networkarea_id

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| exact_include_conditions | object | No | Exact match conditions |
| fuzzy_include_conditions | object | No | Fuzzy match conditions |

#### exact_include_conditions

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_host_id | array<int64> | No | Host ID list |
| bk_biz_id | array<int64> | No | Business ID list |
| bk_networkarea_id | array<int64> | No | Network area ID list |
| os_type | array<string> | No | Host operating system type list |
| node_role | array<string> | No | Node role list |
| node_status | array<string> | No | Node status list |
| node_version | array<string> | No | Node version list |
| bk_agent_id | array<string> | No | Agent ID list |
| bk_networkunit_id | array<int64> | No | Network unit ID list |
| node_generation | array<int64> | No | Node generation list |
| arch | array<string> | No | CPU architecture list |

#### fuzzy_include_conditions

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_host_name | array<string> | No | Host name list, matched with fuzzy conditions |
| dept_name | array<string> | No | Department name list, matched with fuzzy conditions |
| bk_host_innerip | array<string> | No | Host internal IPv4 address list, matched with fuzzy conditions |
| bk_host_innerip_v6 | array<string> | No | Host internal IPv6 address list, matched with fuzzy conditions |
| bk_host_outerip | array<string> | No | Host external IPv4 address list, matched with fuzzy conditions |
| bk_host_outerip_v6 | array<string> | No | Host external IPv6 address list, matched with fuzzy conditions |

### Request Example

```json
{
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "node_status": ["RUNNING"],
    "bk_networkarea_id": [1, 2]
  },
  "fuzzy_include_conditions": {
    "bk_host_innerip": ["10.0."]
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
    "1": 12,
    "2": 5
  }
}
```

### Response Parameters

| Parameter | Type | Description |
|---------|----------|------|
| code | int32 | Status code, `0` means success |
| message | string | Response message |
| request_id | string | Request ID |
| data | object | Host distribution result, where each key is a network area ID and each value is the matched host count |

#### data

| Parameter | Type | Description |
|---------|----------|------|
| {bk_networkarea_id} | int64 | Dynamic field keyed by network area ID |
| value of {bk_networkarea_id} | int64 | Number of matched hosts in that network area |

### Notes

- The contract of this API is defined by `proto/backend/api/v3/topo.proto`, `pkg/proto/backend/api/v3/host.go`, `internal/backend/router/api-v3/topo/host.go`, and `docs/api/swagger/backend/api/v3/topo.swagger.json`.
- `data` is defined as `map<int64, int64>` in proto. In actual JSON responses, object keys are serialized as string-form network area IDs.
