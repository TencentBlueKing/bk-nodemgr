### Description

- API Version: v3.0.1+.
- Required Permission: networkarea_view (View Network Area).
- Function: Count the host distribution for each node role based on host filter conditions.

### URL

POST /api/v3/topo/host/get_host_distribution_by_node_role

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| exact_include_conditions | object | No | Exact match include conditions |
| fuzzy_include_conditions | object | No | Fuzzy match include conditions |

#### exact_include_conditions

Exact match include conditions. Each field is an array and the server filters hosts by these conditions.

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_host_id | int64 array | No | Host ID list |
| bk_biz_id | int64 array | No | Business ID list |
| bk_networkarea_id | int64 array | No | Network area ID list |
| os_type | string array | No | Operating system type list |
| node_role | string array | No | Node role list |
| node_status | string array | No | Node status list |
| node_version | string array | No | Node version list |
| bk_agent_id | string array | No | Agent ID list |
| bk_networkunit_id | int64 array | No | Network unit ID list |
| node_generation | int64 array | No | Node generation list |
| arch | string array | No | CPU architecture list |

#### fuzzy_include_conditions

Fuzzy match include conditions. Each field is an array and the server performs fuzzy filtering on host string fields.

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_host_name | string array | No | Host name list, matched fuzzily by host name |
| dept_name | string array | No | Department name list, matched fuzzily by department name |
| bk_host_innerip | string array | No | Host internal IPv4 list, matched fuzzily by internal IPv4 |
| bk_host_innerip_v6 | string array | No | Host internal IPv6 list, matched fuzzily by internal IPv6 |
| bk_host_outerip | string array | No | Host external IPv4 list, matched fuzzily by external IPv4 |
| bk_host_outerip_v6 | string array | No | Host external IPv6 list, matched fuzzily by external IPv6 |

### Request Example

Count the host distribution of each node role under business `2`.

```json
{
  "exact_include_conditions": {
    "bk_biz_id": [
      2
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
  "data": {
    "agent": 120,
    "proxy": 8,
    "blank": 15
  }
}
```

### Response Parameters

| Parameter | Type | Description |
|---------|----------|------|
| code | int32 | Status code, `0` means success |
| message | string | Response message |
| request_id | string | Request ID |
| error | object | Error information, usually empty on success |
| permission | object | Permission information, typically empty for this API |
| data | object | Mapping from node role to host count |

#### data

| Parameter | Type | Description |
|---------|----------|------|
| {node_role} | int64 | Host count for a node role. The key is the node role string, and current domain role values include `blank`, `agent`, and `proxy` |

### Notes

- The request and response contract is defined by `proto/backend/api/v3/topo.proto`, `pkg/proto/backend/api/v3/host.go`, and `docs/api/swagger/backend/api/v3/topo.swagger.json`.
- In backend router `internal/backend/router/api-v3/topo/host.go`, the handler binds the request body, converts the filters into `types.HostCondition`, calls storage for aggregation, and returns a mapping from node role strings to counts.
