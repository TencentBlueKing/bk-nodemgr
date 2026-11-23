### Description

- API Version: v3.0.1-alpha.105+
- Required Permissions: system_access (System Access), agent_operate (Operate Agent).
- Function: Check Agent installation prerequisites in batches as the authenticated user. Returns matched hosts and installation categories without performing installation.
- The API Gateway authenticates and forwards the user identity. Do not include a username or tenant parameter in the request body. `agent_operate` is checked against the requested businesses. This check does not verify network unit use permission; actual installation still requires it.

### URL

POST /api/v3/open/node/agent/install_check

### Input Parameters

| Parameter Name | Parameter Type | Required | Description                 |
| -------------- | -------------- | -------- | --------------------------- |
| host           | object array   | Yes      | Nonempty host list to check |

#### host[n]

| Parameter Name          | Parameter Type | Required | Description                                                                                          |
| ----------------------- | -------------- | -------- | ---------------------------------------------------------------------------------------------------- |
| bk_host_id              | int64          | No       | Existing host ID; default `-1` checks a new host                                                     |
| bk_biz_id               | int64          | Yes      | Business ID, greater than or equal to `0`                                                            |
| bk_host_innerip_list    | string array   | No       | Inner IPv4 list; this or the IPv6 list must be nonempty; entries must be valid IP addresses          |
| bk_host_innerip_v6_list | string array   | No       | Inner IPv6 list; this or the IPv4 list must be nonempty; entries must be valid IP addresses          |
| bk_networkunit_id       | int64          | No       | Target network unit ID, default `-1`; omission or a nonexistent unit returns `networkunit_not_found` |

Checks whether the target unit exists and supports installation, and verifies host business, network area, IP, and node role consistency. A new host may match an existing host or require CMDB registration. When changing an existing host's network unit, `status` can be `normal_install` while `category` is `need_confirm`. Read both fields.

### Request Example

```json
{
  "host": [
    {
      "bk_host_id": 1001,
      "bk_biz_id": 100,
      "bk_host_innerip_list": ["10.0.0.1"],
      "bk_networkunit_id": 1
    }
  ]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "results": [
      {
        "status": "normal_install",
        "category": "normal_install",
        "message_en": "",
        "message_zh": "",
        "matched": {
          "bk_host_id": 1001,
          "bk_biz_id": 100,
          "bk_networkarea_id": 1,
          "bk_networkunit_id": 1,
          "os_type": "linux",
          "node_role": "agent",
          "bk_host_innerip_list": ["10.0.0.1"],
          "bk_host_innerip_v6_list": []
        }
      }
    ]
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description                                                                              |
| -------------- | -------------- | ---------------------------------------------------------------------------------------- |
| code           | int32          | Status code; `0` means the check request succeeded, not that every host can be installed |
| message        | string         | Request message                                                                          |
| request_id     | string         | Request ID                                                                               |
| error          | object         | Error information                                                                        |
| permission     | object         | Permission information                                                                   |
| data           | object         | Response data                                                                            |
| data.results   | object array   | Check results in the same order as the input hosts                                       |

#### data.results[n]

| Parameter Name | Parameter Type | Description                                                                  |
| -------------- | -------------- | ---------------------------------------------------------------------------- |
| status         | string         | Check status, listed below                                                   |
| category       | string         | `normal_install`, `register_to_cmdb_and_install`, `need_confirm`, or `error` |
| message_en     | string         | English message                                                              |
| message_zh     | string         | Chinese message                                                              |
| matched        | object         | Matched existing host; may be empty if no host matches                       |

#### status

| Value                           | Description                                                                  |
| ------------------------------- | ---------------------------------------------------------------------------- |
| duplicated_inner_ip             | Conflicting inner IPv4 address                                               |
| duplicated_inner_ipv6           | Conflicting inner IPv6 address                                               |
| host_not_found                  | Specified host does not exist                                                |
| networkunit_not_found           | Network unit does not exist or was not specified                             |
| mismatched_inner_ip             | Inner IPv4 addresses do not match the host                                   |
| mismatched_inner_ipv6           | Inner IPv6 addresses do not match the host                                   |
| mismatched_biz_id               | Business does not match                                                      |
| mismatched_networkarea_id       | Network area does not match                                                  |
| invalid_node_role               | Node role does not allow Agent installation, for example an existing Proxy   |
| networkunit_not_support_install | Network unit does not support installation                                   |
| normal_install                  | Installation is allowed; also check `category` for confirmation requirements |
| register_to_cmdb_and_install    | Register the host in CMDB before installation                                |

#### data.results[n].matched

| Parameter Name          | Parameter Type | Description                                            |
| ----------------------- | -------------- | ------------------------------------------------------ |
| bk_host_id              | int64          | Matched host ID                                        |
| bk_biz_id               | int64          | Matched host business ID                               |
| bk_networkarea_id       | int64          | Matched host network area ID                           |
| bk_networkunit_id       | int64          | Matched host network unit ID                           |
| os_type                 | string         | Operating system                                       |
| node_role               | string         | Node role: `agent` or `proxy`; empty string when unset |
| bk_host_innerip_list    | string array   | Matched host inner IPv4 list                           |
| bk_host_innerip_v6_list | string array   | Matched host inner IPv6 list                           |
