### Description

- Available since: v3.0.1-alpha.107+.
- Required permissions: system_access (System Access), networkunit_view (View Network Unit).
- Function: Query network unit graph nodes, Agent/Proxy counts and health information through the user-state API. Results are narrowed to network units the user may view.
- User identity and tenant information are supplied by API Gateway. Do not provide a username or tenant ID in the request body.

### URL

POST /api/v3/open/topo/graph_node/get

### Input Parameters

| Parameter         | Type        | Required | Description                                                                                        |
| ----------------- | ----------- | -------- | -------------------------------------------------------------------------------------------------- |
| bk_networkunit_id | int64 array | No       | Network unit IDs; omission or an empty array queries network units within the user's viewing scope |

### Request Example

```json
{ "bk_networkunit_id": [10] }
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "graph_node_info": [
      {
        "bk_networkunit_id": 10,
        "running_proxy": 1,
        "total_proxy": 1,
        "running_agent": 10,
        "total_agent": 12,
        "is_healthy": true,
        "cycle_times": [
          {
            "bk_host_id": 1001,
            "bk_host_innerip_list": ["10.0.0.10"],
            "bk_host_innerip_v6_list": [],
            "bk_agent_id": "agent-1001",
            "time": "2ms"
          }
        ]
      }
    ]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                                                   |
| ---------- | ------ | ----------------------------------------------------------------------------- |
| code       | int32  | Status code; `0` indicates success                                            |
| message    | string | Request result message                                                        |
| request_id | string | Request ID                                                                    |
| error      | object | Error information; empty on success                                           |
| permission | object | Permission application information that may be returned when access is denied |
| data       | object | Response data                                                                 |

#### data

| Parameter       | Type         | Description                                                 |
| --------------- | ------------ | ----------------------------------------------------------- |
| graph_node_info | object array | Network unit graph nodes; order may differ from the request |

#### data.graph_node_info[n]

| Parameter         | Type         | Description                                                                                                                                                                                                                  |
| ----------------- | ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| bk_networkunit_id | int64        | Network unit ID                                                                                                                                                                                                              |
| running_proxy     | int64        | Running Proxy count                                                                                                                                                                                                          |
| total_proxy       | int64        | Total Proxy count                                                                                                                                                                                                            |
| running_agent     | int64        | Running Agent count                                                                                                                                                                                                          |
| total_agent       | int64        | Total Agent count                                                                                                                                                                                                            |
| is_healthy        | bool         | Current health criterion: at least one running Proxy, with the running Proxies' combined tags covering `dedicated_installer`, `cluster_tunnel`, `file_tunnel` and `data_tunnel`; not a guarantee that all Agents are healthy |
| cycle_times       | object array | Proxy connection round-trip time information                                                                                                                                                                                 |

#### data.graph_node_info[n].cycle_times[n]

| Parameter               | Type         | Description                                                                                                                           |
| ----------------------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------- |
| bk_host_id              | int64        | Proxy host ID                                                                                                                         |
| bk_host_innerip_list    | string array | Private IPv4 addresses                                                                                                                |
| bk_host_innerip_v6_list | string array | Private IPv6 addresses                                                                                                                |
| bk_agent_id             | string       | Proxy Agent ID                                                                                                                        |
| time                    | string       | Connection round-trip time reported by GSE; the original string format is preserved, rather than returning a number with a fixed unit |

### Notes

Returns an empty list when no network units match after authorization narrowing; authorization check failures return permission errors.
