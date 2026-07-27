### Description

- API Version: v3.0.1-alpha.60+.
- Required Permission: `agent_view` (View Agent), `proxy_view` (View Proxy).
- Function: Query the host list with pagination and filtering by host and node attributes, and return results in descending order by the latest update time.

### URL

POST /api/v3/topo/host/list

### Request Parameters

| Parameter                | Type   | Required | Description                                                                                                     |
| ------------------------ | ------ | -------- | --------------------------------------------------------------------------------------------------------------- |
| page                     | object | No       | Pagination configuration. `offset` must be greater than or equal to `0`, and `limit` must be within `(0, 1000]` |
| only_count               | bool   | No       | Whether to return only the total count. When `true`, host details are not returned                              |
| exact_include_conditions | object | No       | Exact match include conditions                                                                                  |
| fuzzy_include_conditions | object | No       | Fuzzy match include conditions                                                                                  |

#### page

| Parameter | Type  | Required | Description                                                    |
| --------- | ----- | -------- | -------------------------------------------------------------- |
| offset    | int32 | No       | Record start offset, starting from `0`                         |
| limit     | int32 | No       | Maximum number of returned records. Valid range is `(0, 1000]` |

#### exact_include_conditions

Exact match include conditions. Each field is an array and the server filters hosts by these conditions. The `node_role` field also affects permission narrowing.

| Parameter         | Type         | Required | Description                                                                                                                             |
| ----------------- | ------------ | -------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| bk_host_id        | int64 array  | No       | Host ID list                                                                                                                            |
| bk_biz_id         | int64 array  | No       | Business ID list                                                                                                                        |
| bk_networkarea_id | int64 array  | No       | Network area ID list                                                                                                                    |
| bk_host_innerip    | string array | No       | Host internal IPv4 list, matched exactly by full internal IPv4                                                                          |
| bk_host_innerip_v6 | string array | No       | Host internal IPv6 list, matched exactly by full internal IPv6                                                                          |
| bk_set_id         | int64 array  | No       | Set ID list                                                                                                                             |
| bk_module_id      | int64 array  | No       | Module ID list                                                                                                                          |
| os_type           | string array | No       | Operating system type list                                                                                                              |
| node_role         | string array | No       | Node role list. Common values include `blank`, `agent`, and `proxy`                                                                     |
| node_status       | string array | No       | Node status list. Valid values include `init`, `running`, `damaged`, `busy`, `starting`, `upgrade`, `stopping`, `uninit`, and `unknown` |
| node_version      | string array | No       | Node version list                                                                                                                       |
| bk_agent_id       | string array | No       | Agent ID list                                                                                                                           |
| bk_networkunit_id | int64 array  | No       | Network unit ID list                                                                                                                    |
| node_generation   | int64 array  | No       | Node generation list                                                                                                                    |
| arch              | string array | No       | CPU architecture list                                                                                                                   |
| proxy_tags        | string array | No       | Proxy tag list. Valid values: `dedicated_installer`, `cluster_tunnel`, `file_tunnel`, `data_tunnel`                                     |

#### fuzzy_include_conditions

Fuzzy match include conditions. Each field is an array and the server performs fuzzy filtering on host string fields.

| Parameter          | Type         | Required | Description                                               |
| ------------------ | ------------ | -------- | --------------------------------------------------------- |
| bk_host_name       | string array | No       | Host name list, matched fuzzily by host name              |
| dept_name          | string array | No       | Department name list, matched fuzzily by department name  |
| bk_host_innerip    | string array | No       | Host internal IPv4 list, matched fuzzily by internal IPv4 |
| bk_host_innerip_v6 | string array | No       | Host internal IPv6 list, matched fuzzily by internal IPv6 |
| bk_host_outerip    | string array | No       | Host external IPv4 list, matched fuzzily by external IPv4 |
| bk_host_outerip_v6 | string array | No       | Host external IPv6 list, matched fuzzily by external IPv6 |

### Request Example

Query running Agent hosts under business `2` and return the first 10 records.

```json
{
  "page": {
    "offset": 0,
    "limit": 10
  },
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "node_role": ["agent"],
    "node_status": ["running"],
    "bk_host_innerip": ["10.0.0.12"],
    "bk_host_innerip_v6": ["2001:db8::12"]
  },
  "fuzzy_include_conditions": {}
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 2,
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
          "bk_host_outerip_list": ["203.0.113.12"],
          "bk_host_outerip_v6_list": [],
          "bk_mac": "00:16:3e:12:34:56",
          "os_type": "linux",
          "cpu_arch": "x86_64",
          "login_ip": "10.0.0.12",
          "login_port": 22,
          "login_user": "root",
          "login_mode": "password",
          "login_credit_valid": true,
          "export_ip": "203.0.113.12",
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
          "node_generation": 1,
          "proxy_tags": []
        }
      }
    ]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                          |
| ---------- | ------ | ---------------------------------------------------- |
| code       | int32  | Status code, `0` means success                       |
| message    | string | Response message                                     |
| request_id | string | Request ID                                           |
| error      | object | Error information, usually empty on success          |
| permission | object | Permission information, typically empty for this API |
| data       | object | Response data                                        |

#### data

| Parameter | Type  | Description                                                               |
| --------- | ----- | ------------------------------------------------------------------------- |
| total     | int64 | Total record count matching the current criteria                          |
| items     | array | Returned host list. When `only_count=true`, this field is typically empty |

#### data.items[n]

| Parameter  | Type   | Description                                           |
| ---------- | ------ | ----------------------------------------------------- |
| tenant_id  | string | Tenant ID                                             |
| bk_host_id | int64  | Host ID                                               |
| info       | object | Static host information and login-related information |
| state      | object | Node state information                                |

#### data.items[n].info

| Parameter               | Type         | Description                                               |
| ----------------------- | ------------ | --------------------------------------------------------- |
| bk_biz_id               | int64        | Business ID                                               |
| bk_networkarea_id       | int64        | Network area ID                                           |
| bk_networkunit_id       | int64        | Network unit ID                                           |
| bk_host_name            | string       | Host name                                                 |
| dept_name               | string       | Department name                                           |
| bk_host_innerip_list    | string array | Internal IPv4 list                                        |
| bk_host_innerip_v6_list | string array | Internal IPv6 list                                        |
| bk_host_outerip_list    | string array | External IPv4 list                                        |
| bk_host_outerip_v6_list | string array | External IPv6 list                                        |
| bk_mac                  | string       | Host MAC address                                          |
| os_type                 | string       | Operating system type                                     |
| cpu_arch                | string       | CPU architecture                                          |
| login_ip                | string       | Login IP                                                  |
| login_port              | int64        | Login port                                                |
| login_user              | string       | Login username                                            |
| login_mode              | string       | Login mode                                                |
| login_credit_valid      | bool         | Whether the stored login credential is still valid        |
| export_ip               | string       | NAT export IPv4                                           |
| export_ip_v6            | string       | NAT export IPv6                                           |
| advertise_ip            | string       | NAT advertised IPv4                                       |
| advertise_ip_v6         | string       | NAT advertised IPv6                                       |
| relay_callback_port     | int64        | Relay callback port                                       |
| relay_download_port     | int64        | Relay download port                                       |
| bk_addressing           | string       | Addressing mode. Common values are `static` and `dynamic` |

#### data.items[n].state

| Parameter       | Type         | Description     |
| --------------- | ------------ | --------------- |
| node_role       | string       | Node role       |
| node_status     | string       | Node status     |
| node_version    | string       | Node version    |
| bk_agent_id     | string       | Agent ID        |
| node_generation | int64        | Node generation |
| proxy_tags      | string array | Proxy tag list  |

### Notes

- The request and response contract is defined by `proto/backend/api/v3/topo.proto`, `proto/backend/api/v3/common.proto`, `pkg/proto/backend/api/v3/host.go`, and `docs/api/swagger/backend/api/v3/topo.swagger.json`.
- In backend router `internal/backend/router/api-v3/topo/host.go`, the request is converted into `types.HostCondition`, then the handler narrows `bk_biz_id` by `agent_view` and/or `proxy_view` according to `node_role` before querying storage.
- If `node_role` is not provided, the server narrows the visible business scope against both Agent and Proxy permissions. If `node_role=["proxy"]`, only the Proxy-visible scope is used.
