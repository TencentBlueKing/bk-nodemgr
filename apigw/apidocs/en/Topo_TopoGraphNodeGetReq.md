### Description

- API Version: v3.0.1+.
- Required Permission: networkunit_view (View Network Unit).
- Function: Retrieve topology graph node information, returning Agent and Proxy statistics and health status for specified network units.

### URL

POST /api/v3/topo/graph_node/get

### Input Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_networkunit_id | array(int64) | No | List of network unit IDs. Returns information for all authorized network units when empty |

### Request Example

```json
{
  "bk_networkunit_id": [1, 2, 3]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "abc123",
  "data": {
    "graph_node_info": [
      {
        "bk_networkunit_id": 1,
        "running_proxy": 2,
        "total_proxy": 2,
        "running_agent": 150,
        "total_agent": 200,
        "is_healthy": true,
        "cycle_times": ["2024-01-01T10:00:00Z", "2024-01-01T10:05:00Z"]
      },
      {
        "bk_networkunit_id": 2,
        "running_proxy": 0,
        "total_proxy": 1,
        "running_agent": 80,
        "total_agent": 100,
        "is_healthy": false,
        "cycle_times": []
      }
    ]
  }
}
```

### Response Parameters

| Parameter | Type | Description |
|---------|----------|------|
| code | int32 | Status code, 0 indicates success |
| message | string | Request message |
| request_id | string | Request ID |
| data | object | Response data |

#### data

| Parameter | Type | Description |
|---------|----------|------|
| graph_node_info | array | List of graph node information |

#### graph_node_info[n]

| Parameter | Type | Description |
|---------|----------|------|
| bk_networkunit_id | int64 | Network unit ID |
| running_proxy | int64 | Number of running Proxies |
| total_proxy | int64 | Total number of Proxies |
| running_agent | int64 | Number of running Agents |
| total_agent | int64 | Total number of Agents |
| is_healthy | bool | Node health status. False when running_proxy is 0 or required Proxy tags are not satisfied |
| cycle_times | array(string) | List of Agent heartbeat cycle times in ISO 8601 format |
