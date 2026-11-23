### Description

- Available since: v3.0.1-alpha.107+.
- Required permissions: system_access (System Access), plus agent_view (View Agent) or proxy_view (View Proxy), depending on `node_role`. `blank` and `agent` require Agent viewing permission; `proxy` requires Proxy viewing permission. When roles are omitted or both permission categories are requested, the authorized business scopes of all required actions are intersected.
- Function: Query the instance topology of a business through the user-state API, including host counts matching the node role filter at each node.
- User identity and tenant information are supplied by API Gateway. Do not provide a username or tenant ID in the request body.

### URL

POST /api/v3/open/topo/business/inst_topo/get

### Input Parameters

| Parameter | Type         | Required | Description                                                                                          |
| --------- | ------------ | -------- | ---------------------------------------------------------------------------------------------------- |
| bk_biz_id | int64        | Yes      | Business ID; must not be `0`. Viewing permission for this business is required                       |
| node_role | string array | No       | Node roles: `blank` (node not installed), `agent`, or `proxy`; omitted or empty means no role filter |

### Request Example

```json
{ "bk_biz_id": 2, "node_role": ["agent"] }
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "items": {
      "topo_obj_id": "biz",
      "topo_inst_id": 2,
      "topo_inst_name": "Example Business",
      "host_count": 10,
      "children": [
        {
          "topo_obj_id": "set",
          "topo_inst_id": 20,
          "topo_inst_name": "Example Set",
          "host_count": 10,
          "children": []
        }
      ]
    }
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                                                                                  |
| ---------- | ------ | ------------------------------------------------------------------------------------------------------------ |
| code       | int32  | Status code; `0` indicates success                                                                           |
| message    | string | Request result message                                                                                       |
| request_id | string | Request ID                                                                                                   |
| error      | object | Error information, including `system`, `message`, and `details` (error codes and messages); empty on success |
| permission | object | Permission application information that may be returned when access is denied                                |
| data       | object | Response data                                                                                                |

#### data

| Parameter | Type   | Description                                                                                 |
| --------- | ------ | ------------------------------------------------------------------------------------------- |
| items     | object | A single business topology root node, **not an array**; empty when CMDB returns no topology |

#### data.items / children[n]

Child nodes recursively use the same structure.

| Parameter      | Type         | Description                                                                                                    |
| -------------- | ------------ | -------------------------------------------------------------------------------------------------------------- |
| topo_obj_id    | string       | Topology object type, such as `biz`, `set`, or `module`; CMDB custom levels may also occur                     |
| topo_inst_id   | int64        | Topology instance ID                                                                                           |
| topo_inst_name | string       | Topology instance name                                                                                         |
| host_count     | int64        | Hosts in this node's scope matching the role filter; custom levels count deduplicated hosts in descendant sets |
| children       | object array | Child nodes                                                                                                    |

### Notes

- Missing the required viewing permissions for the business returns a permission error, not an empty topology.
- Topology comes from CMDB while host counts use Node Manager host records. Synchronization delays may cause differences.
