### Description

- API Version: v3.0.1-alpha.38+.
- Required Permission: `agent_view (View Agent)`, `proxy_view (View Proxy)`.
- Function: Query the business instance topology by business ID, with aggregated host counts for topology nodes.

### URL

POST /api/v3/topo/business/inst_topo/get

### Request Parameters

| Parameter | Type  | Required | Description |
| --------- | ----- | -------- | ----------- |
| bk_biz_id | int64 | Yes      | Business ID |

### Request Example

Query the business instance topology of business `2`.

```json
{
  "bk_biz_id": 2
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "items": {
      "topo_inst_id": 2,
      "topo_inst_name": "prod-payment",
      "topo_obj_id": "biz",
      "host_count": 155,
      "children": [
        {
          "topo_inst_id": 10,
          "topo_inst_name": "default set",
          "topo_obj_id": "set",
          "host_count": 120,
          "children": [
            {
              "topo_inst_id": 100,
              "topo_inst_name": "module-a",
              "topo_obj_id": "module",
              "host_count": 120,
              "children": []
            }
          ]
        },
        {
          "topo_inst_id": 11,
          "topo_inst_name": "custom layer",
          "topo_obj_id": "custom_level",
          "host_count": 35,
          "children": [
            {
              "topo_inst_id": 101,
              "topo_inst_name": "module-b",
              "topo_obj_id": "module",
              "host_count": 35,
              "children": []
            }
          ]
        }
      ]
    }
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                          |
| ---------- | ------ | ---------------------------------------------------- |
| code       | int32  | Status code, `0` means success                       |
| message    | string | Response message                                     |
| request_id | string | Request ID                                           |
| error      | object | Error information, empty on success                  |
| permission | object | Permission information, typically empty for this API |
| data       | object | Response data                                        |

#### data

| Parameter | Type   | Description                                                                               |
| --------- | ------ | ----------------------------------------------------------------------------------------- |
| items     | object | Root node of the business instance topology. It may be empty if no topology root is found |

#### data.items / data.items.children[n]

| Parameter      | Type   | Description                                                                          |
| -------------- | ------ | ------------------------------------------------------------------------------------ |
| topo_inst_id   | int64  | Topology instance ID                                                                 |
| topo_inst_name | string | Topology instance name                                                               |
| topo_obj_id    | string | Topology object ID, such as `biz`, `set`, `module`, or a CMDB custom level object ID |
| host_count     | int64  | Aggregated host count for the current topology node                                  |
| children       | array  | Child topology node list. Each child uses the same node structure                    |

### Notes

- The request and response contract is defined by `proto/backend/api/v3/topo.proto`, `pkg/proto/backend/api/v3/business.go`, `internal/backend/router/api-v3/topo/business.go`, and `docs/api/swagger/backend/api/v3/topo.swagger.json`.
- If `bk_biz_id` is `0`, protocol-layer validation returns a parameter error.
- The backend queries the business instance topology from CMDB and counts hosts directly for `biz`, `set`, and `module` nodes. For other custom topology levels, `host_count` is aggregated from child nodes.
- `data.items` is the root node of a single business topology tree. Every `children` element uses the same node structure.
