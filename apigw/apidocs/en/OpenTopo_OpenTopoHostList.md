### Description

- Available since: v3.0.1-alpha.107+.
- Required permissions: system_access (System Access), plus agent_view (View Agent) or proxy_view (View Proxy) according to `node_role`. `blank` and `agent` use Agent permission; `proxy` uses Proxy permission. If roles are omitted or both categories are requested, the authorized business scopes for the required actions are intersected.
- Function: Query hosts through the user-state API with pagination and property filters. Results are narrowed to authorized businesses and ordered by most recent update first.
- User identity and tenant information are supplied by API Gateway. Do not provide a username or tenant ID in the request body.

### URL

POST /api/v3/open/topo/host/list

### Input Parameters

| Parameter                | Type   | Required    | Description                                                                 |
| ------------------------ | ------ | ----------- | --------------------------------------------------------------------------- |
| page                     | object | Conditional | A valid `limit` is required unless `only_count=true`                        |
| only_count               | bool   | No          | Defaults to `false`; when `true`, returns only the total with empty `items` |
| exact_include_conditions | object | No          | Exact filters                                                               |
| fuzzy_include_conditions | object | No          | Fuzzy filters                                                               |

#### page

| Parameter | Type  | Required    | Description                                                                                                |
| --------- | ----- | ----------- | ---------------------------------------------------------------------------------------------------------- |
| offset    | int32 | No          | Starting record offset, defaults to `0`; must be nonnegative                                               |
| limit     | int32 | Conditional | In `(0, 1000]` for list queries; may be omitted for count-only queries, but cannot be negative if provided |

#### exact_include_conditions

Each field provides candidate values as an array. `node_role` also determines the required viewing permissions.

| Parameter          | Type         | Required | Description                                                                                               |
| ------------------ | ------------ | -------- | --------------------------------------------------------------------------------------------------------- |
| bk_host_id         | int64 array  | No       | Host IDs                                                                                                  |
| bk_biz_id          | int64 array  | No       | Business IDs, intersected with the authorized business scope                                              |
| bk_networkarea_id  | int64 array  | No       | Network area IDs                                                                                          |
| bk_set_id          | int64 array  | No       | CMDB set IDs                                                                                              |
| bk_module_id       | int64 array  | No       | CMDB module IDs                                                                                           |
| bk_host_innerip    | string array | No       | Exact private IPv4 addresses                                                                              |
| bk_host_innerip_v6 | string array | No       | Exact private IPv6 addresses                                                                              |
| os_type            | string array | No       | Operating system types                                                                                    |
| node_role          | string array | No       | Node roles: `blank`, `agent`, `proxy`                                                                     |
| node_status        | string array | No       | Node states: `init`, `running`, `damaged`, `busy`, `starting`, `upgrade`, `stopping`, `uninit`, `unknown` |
| node_version       | string array | No       | Node versions                                                                                             |
| bk_agent_id        | string array | No       | Agent IDs                                                                                                 |
| bk_networkunit_id  | int64 array  | No       | Network unit IDs                                                                                          |
| node_generation    | int64 array  | No       | Node generations                                                                                          |
| arch               | string array | No       | CPU architectures                                                                                         |
| proxy_tags         | string array | No       | Proxy tags: `dedicated_installer`, `cluster_tunnel`, `file_tunnel`, `data_tunnel`                         |

#### fuzzy_include_conditions

| Parameter          | Type         | Required | Description                   |
| ------------------ | ------------ | -------- | ----------------------------- |
| bk_host_name       | string array | No       | Host name patterns            |
| dept_name          | string array | No       | Business module name patterns |
| bk_host_innerip    | string array | No       | Private IPv4 patterns         |
| bk_host_innerip_v6 | string array | No       | Private IPv6 patterns         |
| bk_host_outerip    | string array | No       | Public IPv4 patterns          |
| bk_host_outerip_v6 | string array | No       | Public IPv6 patterns          |

### Request Example

```json
{
  "page": { "offset": 0, "limit": 10 },
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "node_role": ["agent"],
    "node_status": ["running"]
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
    "total": 1,
    "items": [
      {
        "tenant_id": "default",
        "bk_host_id": 1001,
        "info": {
          "bk_biz_id": 2,
          "bk_networkarea_id": 1,
          "bk_networkunit_id": 10,
          "bk_host_name": "agent-prod-01",
          "dept_name": "payment",
          "bk_host_innerip_list": ["10.0.0.12"],
          "bk_host_innerip_v6_list": [],
          "bk_host_outerip_list": [],
          "bk_host_outerip_v6_list": [],
          "bk_mac": "00:16:3e:12:34:56",
          "os_type": "linux",
          "cpu_arch": "x86_64",
          "login_ip": "10.0.0.12",
          "login_port": 22,
          "login_user": "root",
          "login_mode": "password",
          "login_credit_valid": true,
          "export_ip": "",
          "export_ip_v6": "",
          "advertise_ip": "10.0.0.12",
          "advertise_ip_v6": "",
          "relay_callback_port": 0,
          "relay_download_port": 0,
          "bk_addressing": "static"
        },
        "state": {
          "node_role": "agent",
          "node_status": "running",
          "node_version": "2.4.1",
          "bk_agent_id": "agent-1001",
          "node_generation": 2,
          "proxy_tags": [],
          "last_sync_at": 1795392000000
        }
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

| Parameter | Type         | Description                                             |
| --------- | ------------ | ------------------------------------------------------- |
| total     | int64        | Total matching hosts within the authorized scope        |
| items     | object array | Hosts on the current page; empty for count-only queries |

#### data.items[n]

| Parameter  | Type            | Description                                                                                                  |
| ---------- | --------------- | ------------------------------------------------------------------------------------------------------------ |
| tenant_id  | string          | Tenant ID                                                                                                    |
| bk_host_id | int64           | Host ID                                                                                                      |
| info       | object          | Static host and login information; login passwords and key content are not returned                          |
| state      | object          | Node state                                                                                                   |
| create_at  | string (uint64) | Creation time field; not populated by the current list response, so do not use it to determine creation time |
| updated_at | string (uint64) | Update time field; not populated by the current list response, so do not use it to determine update time     |

#### data.items[n].info

| Parameter               | Type         | Description                            |
| ----------------------- | ------------ | -------------------------------------- |
| bk_biz_id               | int64        | Business ID                            |
| bk_networkarea_id       | int64        | Network area ID                        |
| bk_networkunit_id       | int64        | Network unit ID                        |
| bk_host_name            | string       | Host name                              |
| dept_name               | string       | Business module name                   |
| bk_host_innerip_list    | string array | Private IPv4 addresses                 |
| bk_host_innerip_v6_list | string array | Private IPv6 addresses                 |
| bk_host_outerip_list    | string array | Public IPv4 addresses                  |
| bk_host_outerip_v6_list | string array | Public IPv6 addresses                  |
| bk_mac                  | string       | MAC address                            |
| os_type                 | string       | Operating system type                  |
| cpu_arch                | string       | CPU architecture                       |
| login_ip                | string       | Login IP                               |
| login_port              | int64        | Login port                             |
| login_user              | string       | Login user                             |
| login_mode              | string       | Login mode                             |
| login_credit_valid      | bool         | Whether login credentials are valid    |
| export_ip               | string       | NAT egress IPv4                        |
| export_ip_v6            | string       | NAT egress IPv6                        |
| advertise_ip            | string       | NAT advertised IPv4                    |
| advertise_ip_v6         | string       | NAT advertised IPv6                    |
| relay_callback_port     | int64        | Relay callback port                    |
| relay_download_port     | int64        | Relay download port                    |
| bk_addressing           | string       | Addressing mode: `static` or `dynamic` |

#### data.items[n].state

| Parameter       | Type         | Description                                                                                           |
| --------------- | ------------ | ----------------------------------------------------------------------------------------------------- |
| node_role       | string       | Node role, using the values described for the query filter                                            |
| node_status     | string       | Node state, using the values described for the query filter                                           |
| node_version    | string       | Node version                                                                                          |
| bk_agent_id     | string       | Agent ID                                                                                              |
| node_generation | int64        | Node generation                                                                                       |
| proxy_tags      | string array | Proxy tags, using the values described for the query filter                                           |
| last_sync_at    | int64        | Unix timestamp in milliseconds of the latest state sync from GSE; `0` means no sync time is available |

### Notes

Returns an empty list when the authorized business scope and query conditions do not intersect; authorization check failures return permission errors. Both pagination totals and count-only results use the same authorized scope.
