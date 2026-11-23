### Description

- Available since: v3.0.1-alpha.107+.
- Required permissions: system_access (System Access), networkunit_view (View Network Unit).
- Function: Query a network unit, its access points, links and deployment configuration by ID through the user-state API. Viewing permission is checked for the target network unit.
- User identity and tenant information are supplied by API Gateway. Do not provide a username or tenant ID in the request body.

### URL

POST /api/v3/open/topo/networkunit/get

### Input Parameters

| Parameter         | Type  | Required | Description                                                   |
| ----------------- | ----- | -------- | ------------------------------------------------------------- |
| bk_networkunit_id | int64 | No       | Network unit ID; must be nonnegative. Omission queries ID `0` |

### Request Example

```json
{ "bk_networkunit_id": 10 }
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "tenant_id": "default",
    "bk_networkunit_id": 10,
    "bk_networkunit_name": "unit-prod",
    "bk_networkarea_id": 1,
    "accesspoints": [
      {
        "tenant_id": "default",
        "accesspoint_id": 1001,
        "accesspoint_name": "ap-prod",
        "bk_networkarea_id": 1,
        "endpoints": {
          "cluster": ["10.0.0.10:48668"],
          "file": ["10.0.0.10:48669"],
          "data": ["10.0.0.10:48670"]
        },
        "cloud_security_group_pool_id": ""
      }
    ],
    "links": {},
    "is_direct": false,
    "direct_endpoints": null,
    "generation": 2,
    "custom_deploy_config": {}
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
| data       | object | Network unit details                                                          |

#### data

| Parameter            | Type         | Description                                                                     |
| -------------------- | ------------ | ------------------------------------------------------------------------------- |
| tenant_id            | string       | Tenant ID                                                                       |
| bk_networkunit_id    | int64        | Network unit ID                                                                 |
| bk_networkunit_name  | string       | Network unit name                                                               |
| bk_networkarea_id    | int64        | Owning network area ID                                                          |
| accesspoints         | object array | Full access point information, not an ID array                                  |
| links                | object       | Upstream links by `cluster`, `file` and `data` channel                          |
| is_direct            | bool         | Whether direct endpoints are used                                               |
| direct_endpoints     | object       | Direct endpoints; may be empty when not configured                              |
| generation           | int64        | Node generation                                                                 |
| custom_deploy_config | object       | Custom deployment configuration keyed by operating system type, such as `linux` |

#### data.accesspoints[n]

| Parameter                    | Type   | Description                                                                                 |
| ---------------------------- | ------ | ------------------------------------------------------------------------------------------- |
| tenant_id                    | string | Tenant ID                                                                                   |
| accesspoint_id               | int64  | Access point ID                                                                             |
| accesspoint_name             | string | Access point name                                                                           |
| bk_networkarea_id            | int64  | Owning network area ID                                                                      |
| endpoints                    | object | Endpoints, using the structure below                                                        |
| cloud_security_group_pool_id | string | Bound global security group pool ID; empty when cloud security group management is disabled |

#### data.direct_endpoints / data.accesspoints[n].endpoints

| Parameter | Type         | Description               |
| --------- | ------------ | ------------------------- |
| cluster   | string array | Cluster channel addresses |
| file      | string array | File channel addresses    |
| data      | string array | Data channel addresses    |

#### data.links

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

#### data.custom_deploy_config.{os_type}

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
