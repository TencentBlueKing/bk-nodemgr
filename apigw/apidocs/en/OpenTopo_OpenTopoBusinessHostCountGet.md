### Description

- Available since: v3.0.1-alpha.107+.
- Required permissions: system_access (System Access), plus agent_view (View Agent) or proxy_view (View Proxy), depending on `node_role`. `blank` and `agent` require Agent viewing permission; `proxy` requires Proxy viewing permission. When roles are omitted or both permission categories are requested, the authorized business scopes of all required actions are intersected.
- Function: Count hosts by business through the user-state API, limited to the requested node roles and the user's authorized business scope.
- User identity and tenant information are supplied by API Gateway. Do not provide a username or tenant ID in the request body.

### URL

POST /api/v3/open/topo/business/host_count/get

### Input Parameters

| Parameter | Type         | Required | Description                                                                                          |
| --------- | ------------ | -------- | ---------------------------------------------------------------------------------------------------- |
| bk_biz_id | int64 array  | Yes      | Nonempty list of business IDs to count; intersected with the user's authorized business scope        |
| node_role | string array | No       | Node roles: `blank` (node not installed), `agent`, or `proxy`; omitted or empty means no role filter |

### Request Example

```json
{
  "bk_biz_id": [2, 3],
  "node_role": ["agent"]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "items": [
      { "bk_biz_id": 2, "host_count": 100 },
      { "bk_biz_id": 3, "host_count": 20 }
    ]
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

| Parameter | Type         | Description                                                                                             |
| --------- | ------------ | ------------------------------------------------------------------------------------------------------- |
| items     | object array | Counts grouped by business; zero-count entries are not guaranteed for businesses without matching hosts |

#### data.items[n]

| Parameter  | Type  | Description                                   |
| ---------- | ----- | --------------------------------------------- |
| bk_biz_id  | int64 | Business ID                                   |
| host_count | int64 | Number of hosts matching the node role filter |

### Notes

- If the requested business IDs do not intersect with the authorized business scope, `items` is empty. Failed permission checks return a permission error.
- The result order is not guaranteed to follow the requested business ID order.
