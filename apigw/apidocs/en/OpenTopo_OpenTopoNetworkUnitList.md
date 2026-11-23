### Description

- Available since: v3.0.1-alpha.107+.
- Required permissions: system_access (System Access), networkunit_view (View Network Unit).
- Function: Query network units through the user-state API, narrowed to the network units the user may view.
- User identity and tenant information are supplied by API Gateway. Do not provide a username or tenant ID in the request body.

### URL

POST /api/v3/open/topo/networkunit/list

### Input Parameters

| Parameter                | Type   | Required | Description                                                                          |
| ------------------------ | ------ | -------- | ------------------------------------------------------------------------------------ |
| page                     | object | Yes      | Pagination; a valid `limit` is required even when `only_count=true`                  |
| only_count               | bool   | No       | Not handled separately by this endpoint; it still returns a paginated list and total |
| exact_include_conditions | object | No       | Exact filters                                                                        |

#### page

| Parameter | Type  | Required | Description                                                  |
| --------- | ----- | -------- | ------------------------------------------------------------ |
| offset    | int32 | No       | Starting record offset, defaults to `0`; must be nonnegative |
| limit     | int32 | Yes      | Maximum records to return, in `(0, 1000]`                    |

#### exact_include_conditions

| Parameter         | Type        | Required | Description                                                    |
| ----------------- | ----------- | -------- | -------------------------------------------------------------- |
| bk_networkunit_id | int64 array | No       | Network unit IDs, intersected with the user's authorized scope |
| bk_networkarea_id | int64 array | No       | Network area IDs                                               |
| is_direct         | bool array  | No       | Whether direct endpoints are used                              |
| generation        | int64 array | No       | Node generations                                               |

### Request Example

```json
{
  "page": { "offset": 0, "limit": 10 },
  "exact_include_conditions": { "bk_networkarea_id": [1] }
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
    "items": [
      {
        "tenant_id": "default",
        "bk_networkunit_id": 10,
        "bk_networkunit_name": "unit-prod",
        "bk_networkarea_id": 1,
        "accesspoints": [1001],
        "links": {},
        "is_direct": false,
        "direct_endpoints": null,
        "generation": 2,
        "custom_deploy_config": {}
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

| Parameter | Type         | Description                                        |
| --------- | ------------ | -------------------------------------------------- |
| total     | int64        | Total matching records within the authorized scope |
| items     | object array | Network units on the current page                  |

#### data.items[n]

| Parameter            | Type        | Description                                                                     |
| -------------------- | ----------- | ------------------------------------------------------------------------------- |
| tenant_id            | string      | Tenant ID                                                                       |
| bk_networkunit_id    | int64       | Network unit ID                                                                 |
| bk_networkunit_name  | string      | Network unit name                                                               |
| bk_networkarea_id    | int64       | Owning network area ID                                                          |
| accesspoints         | int64 array | Access point IDs, not full access point objects                                 |
| links                | object      | Upstream link configuration                                                     |
| is_direct            | bool        | Whether direct endpoints are used                                               |
| direct_endpoints     | object      | Direct endpoints; may be empty when not configured                              |
| generation           | int64       | Node generation                                                                 |
| custom_deploy_config | object      | Custom deployment configuration keyed by operating system type, such as `linux` |

#### data.items[n].links

| Parameter | Type   | Description                   |
| --------- | ------ | ----------------------------- |
| cluster   | object | Upstream cluster channel link |
| file      | object | Upstream file channel link    |
| data      | object | Upstream data channel link    |

Each channel link uses the following structure; unconfigured channels may be empty.

| Parameter         | Type  | Description              |
| ----------------- | ----- | ------------------------ |
| bk_networkarea_id | int64 | Upstream network area ID |
| bk_networkunit_id | int64 | Upstream network unit ID |
| accesspoint_id    | int64 | Upstream access point ID |

#### data.items[n].direct_endpoints

| Parameter | Type         | Description               |
| --------- | ------------ | ------------------------- |
| cluster   | string array | Cluster channel addresses |
| file      | string array | File channel addresses    |
| data      | string array | Data channel addresses    |

#### data.items[n].custom_deploy_config.{os_type}

| Parameter         | Type   | Description                     |
| ----------------- | ------ | ------------------------------- |
| installer_runtime | object | Installer runtime configuration |
| node_runtime      | object | Node runtime configuration      |
| plugin_runtime    | object | Plugin runtime configuration    |

#### installer_runtime

| Parameter           | Type         | Description                                                                                                  |
| ------------------- | ------------ | ------------------------------------------------------------------------------------------------------------ |
| base_work_dir       | string       | Installer working root directory                                                                             |
| firewall            | object       | Firewall configuration; omission inherits deployment defaults, while an empty object disables firewall setup |
| firewall.remote_ips | string array | Remote IP configuration                                                                                      |
| firewall.features   | string array | Firewall feature configuration                                                                               |

#### node_runtime

| Parameter       | Type   | Description                    |
| --------------- | ------ | ------------------------------ |
| base_deploy_dir | string | Node deployment root directory |
| data_ipc        | string | Data IPC configuration         |
| plugin_ipc      | string | Plugin IPC configuration       |
| log_dir         | string | Node log directory             |
| zone_id         | string | Zone identifier                |
| city_id         | string | City identifier                |

#### plugin_runtime

| Parameter       | Type   | Description                      |
| --------------- | ------ | -------------------------------- |
| base_deploy_dir | string | Plugin deployment root directory |
| log_dir         | string | Plugin log directory             |

### Notes

Returns an empty list when the authorized scope and query conditions do not intersect; authorization check failures return permission errors. Use `OpenTopo_OpenTopoNetworkUnitGet` for full access point information.
