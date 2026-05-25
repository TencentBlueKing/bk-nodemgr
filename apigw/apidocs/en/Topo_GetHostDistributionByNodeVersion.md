### Description

- API Version: v3.0.1+.
- Required Permission: networkarea_view (View Network Area).
- Function: Retrieve host count distribution grouped by node version based on host filtering conditions.

### URL

POST /api/v3/topo/host/get_host_distribution_by_node_version

### Request Parameters

| Parameter                | Type   | Required | Description                      |
| ------------------------ | ------ | -------- | -------------------------------- |
| exact_include_conditions | object | No       | Exact match inclusion conditions |
| fuzzy_include_conditions | object | No       | Fuzzy match inclusion conditions |

#### exact_include_conditions

Exact match inclusion conditions. Each field is an array, and the server filters hosts based on these conditions.

| Parameter         | Type         | Required | Description                |
| ----------------- | ------------ | -------- | -------------------------- |
| bk_host_id        | int64 array  | No       | Host ID list               |
| bk_biz_id         | int64 array  | No       | Business ID list           |
| bk_networkarea_id | int64 array  | No       | Network area ID list       |
| os_type           | string array | No       | Operating system type list |
| node_role         | string array | No       | Node role list             |
| node_status       | string array | No       | Node status list           |
| node_version      | string array | No       | Node version list          |
| bk_agent_id       | string array | No       | Agent ID list              |
| bk_networkunit_id | int64 array  | No       | Network unit ID list       |
| node_generation   | int64 array  | No       | Node generation list       |
| arch              | string array | No       | CPU architecture list      |
| proxy_tags        | string array | No       | Proxy tag list. Valid values: `dedicated_installer`, `cluster_tunnel`, `file_tunnel`, `data_tunnel` |

#### fuzzy_include_conditions

Fuzzy match inclusion conditions. Each field is an array, and the server performs fuzzy filtering on host string fields.

| Parameter          | Type         | Required | Description                                          |
| ------------------ | ------------ | -------- | ---------------------------------------------------- |
| bk_host_name       | string array | No       | Host name list, fuzzy match by host name             |
| dept_name          | string array | No       | Department name list, fuzzy match by department name |
| bk_host_innerip    | string array | No       | Host inner IPv4 list, fuzzy match by inner IPv4      |
| bk_host_innerip_v6 | string array | No       | Host inner IPv6 list, fuzzy match by inner IPv6      |
| bk_host_outerip    | string array | No       | Host outer IPv4 list, fuzzy match by outer IPv4      |
| bk_host_outerip_v6 | string array | No       | Host outer IPv6 list, fuzzy match by outer IPv6      |

### Request Example

Get host count distribution by node version for business `2`.

```json
{
  "exact_include_conditions": {
    "bk_biz_id": [2]
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
    "2.1.5": 85,
    "2.1.6": 43,
    "2.2.0": 12
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                        |
| ---------- | ------ | -------------------------------------------------- |
| code       | int32  | Status code, 0 indicates success                   |
| message    | string | Request message                                    |
| request_id | string | Request ID                                         |
| error      | object | Error information, usually empty on success        |
| permission | object | Permission information, usually empty for this API |
| data       | object | Mapping from node version to host count            |

#### data

| Parameter      | Type  | Description                                                                                     |
| -------------- | ----- | ----------------------------------------------------------------------------------------------- |
| {node_version} | int64 | Host count for a specific node version, key is the node version string (e.g., "2.1.5", "2.1.6") |

### Notes

- The request contract and response structure are defined in `proto/backend/api/v3/topo.proto`, `pkg/proto/backend/api/v3/host.go`, and `docs/api/swagger/backend/api/v3/topo.swagger.json`.
- The backend router in `internal/backend/router/api-v3/topo/host.go` performs `BindJSON` on the request body, converts the filtering conditions to `types.HostCondition`, and calls the storage layer for statistics. The return value is a mapping from node version string to count.
