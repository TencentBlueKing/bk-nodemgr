### Description

- API Version: v3.0.0+.
- Required Permission: agent_operate (Operate Agent).
- Function: Batch check whether hosts meet agent installation requirements, including network unit availability, IP conflict detection, and host attribute consistency validation.

### URL

POST /api/v3/node/agent/install_check

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| host | object array | Yes | List of host information to check |

#### host[n]

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_biz_id | int64 | Yes | Business ID, must be greater than or equal to 0 |
| bk_host_id | int64 | No | Host ID, -1 or omitted means a new host (not yet in CMDB) |
| bk_host_innerip_list | string array | No | Host inner network IPv4 address list, at least one of bk_host_innerip_list and bk_host_innerip_v6_list must be provided |
| bk_host_innerip_v6_list | string array | No | Host inner network IPv6 address list, at least one of bk_host_innerip_list and bk_host_innerip_v6_list must be provided |
| bk_networkunit_id | int64 | No | Network unit ID, defaults to -1 if not provided |

**Parameter Notes**:
- `bk_host_innerip_list` and `bk_host_innerip_v6_list` cannot both be empty, and neither list may contain empty strings
- When `bk_host_id` is -1 or not provided, the host is treated as not yet registered in CMDB; the system checks for IP conflicts with existing hosts
- When `bk_host_id` >= 0, the host is treated as already in CMDB; the system validates host attributes (business ID, network area ID, IP addresses, etc.) for consistency

### Request Example

Check installation conditions for two hosts: one already in CMDB and one new host.

```json
{
  "host": [
    {
      "bk_biz_id": 100,
      "bk_host_id": 12345,
      "bk_host_innerip_list": ["10.0.0.1"],
      "bk_networkunit_id": 1
    },
    {
      "bk_biz_id": 100,
      "bk_host_id": -1,
      "bk_host_innerip_list": ["10.0.0.2"],
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
  "request_id": "req-1234567890",
  "data": {
    "results": [
      {
        "status": "normal_install",
        "message_en": "Install Agent",
        "message_zh": "安装Agent",
        "category": "normal_install",
        "matched": {
          "bk_host_id": 12345,
          "bk_biz_id": 100,
          "bk_networkarea_id": 0,
          "bk_networkunit_id": 1,
          "os_type": "linux",
          "node_role": "agent",
          "bk_host_innerip_list": ["10.0.0.1"],
          "bk_host_innerip_v6_list": []
        }
      },
      {
        "status": "register_to_cmdb_and_install",
        "message_en": "Import node to CMDB and install Agent",
        "message_zh": "将节点导入CMDB并安装Agent",
        "category": "register_to_cmdb_and_install"
      }
    ]
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, 0 indicates success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, null on success |
| data | object | Response data |

#### data

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| results | object array | Check results list, in one-to-one correspondence with the request host array |

#### data.results[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| status | string | Check status (see enum values below) |
| message_en | string | English description of the check result |
| message_zh | string | Chinese description of the check result |
| category | string | Result category (see enum values below) |
| matched | object | Matched existing host information, null for some statuses |

**status enum values**:

| Value | Description |
|-------|-------------|
| normal_install | Normal install, host is in CMDB and all validations passed |
| register_to_cmdb_and_install | Needs to register in CMDB first then install, new host with no IP conflicts |
| duplicated_inner_ip | Inner IPv4 address conflict, a host with the same IP already exists in the same network area |
| duplicated_inner_ipv6 | Inner IPv6 address conflict, a host with the same IPv6 already exists in the same network area |
| host_not_found | The specified host ID does not exist in the system |
| networkunit_not_found | The specified network unit ID does not exist |
| networkunit_not_support_install | The network unit does not support installation (non-direct network unit without an available installer proxy) |
| mismatched_inner_ip | Host inner IPv4 address does not match the CMDB record |
| mismatched_inner_ipv6 | Host inner IPv6 address does not match the CMDB record |
| mismatched_biz_id | The business ID in the request does not match the host's actual business ID |
| mismatched_networkarea_id | The network area ID of the network unit does not match the host's actual network area ID |
| invalid_node_role | The host's node role is Proxy and cannot install Agent |

**category enum values**:

| Value | Description |
|-------|-------------|
| normal_install | Normal installation, no user confirmation needed |
| register_to_cmdb_and_install | New host needs CMDB registration before installation |
| need_confirm | Requires user confirmation (e.g., IP conflict will trigger a reinstall) |
| error | Check failed, installation cannot proceed |

#### data.results[n].matched

Returned when the check involves an existing host; null when a new host has no conflicts.

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| bk_host_id | int64 | Host ID |
| bk_biz_id | int64 | Business ID |
| bk_networkarea_id | int64 | Network area ID |
| bk_networkunit_id | int64 | Network unit ID |
| os_type | string | Operating system type (enum values: linux, windows, darwin) |
| node_role | string | Node role (enum values: blank, agent, proxy) |
| bk_host_innerip_list | string array | Host inner network IPv4 address list |
| bk_host_innerip_v6_list | string array | Host inner network IPv6 address list |

**node_role enum values**:
- `blank`: Blank node, nothing installed
- `agent`: Agent node
- `proxy`: Proxy node

#### error

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| system | string | Error system identifier |
| message | string | Error message |
| details | array | Error details list |

#### error.details[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | string | Error code |
| message | string | Error message |
