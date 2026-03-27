### Description

- API Version: v3.0.1+.
- Required Permission: proxy_operate (Operate Proxy).
- Function: Batch check whether hosts can have Proxy installed, returning pre-installation status for each host.

### URL

POST /api/v3/node/proxy/install_check

### Input Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| host | array | Yes | Host list, see host parameters below |

**host[n]**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| bk_networkunit_id | int64 | Yes | Network unit ID |
| bk_host_id | int64 | No | Host ID, -1 for new host registration |
| bk_host_innerip_list | array[string] | No | Inner IPv4 address list |
| bk_host_innerip_v6_list | array[string] | No | Inner IPv6 address list |
| bk_biz_id | int64 | No | Business ID |

### Call Example

```json
{
    "host": [
        {
            "bk_networkunit_id": 1,
            "bk_host_id": 1001,
            "bk_host_innerip_list": ["10.0.0.1"],
            "bk_host_innerip_v6_list": [],
            "bk_biz_id": 2
        },
        {
            "bk_networkunit_id": 1,
            "bk_host_id": -1,
            "bk_host_innerip_list": ["10.0.0.2"],
            "bk_biz_id": 2
        }
    ]
}
```

### Response Example

```json
{
    "code": 0,
    "message": "ok",
    "request_id": "abc123",
    "data": {
        "result": [
            {
                "bk_host_id": 1001,
                "status": "normal_install",
                "matched": {
                    "bk_host_id": 1001,
                    "bk_networkunit_id": 1,
                    "bk_host_innerip_list": ["10.0.0.1"],
                    "bk_host_innerip_v6_list": []
                }
            },
            {
                "bk_host_id": -1,
                "status": "register_to_cmdb_and_install",
                "matched": null
            }
        ]
    }
}
```

### Response Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| code | int32 | Status code, 0 means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error info, null on success |
| data | object | Response data |

**data**

| Parameter | Type | Description |
|-----------|------|-------------|
| result | array | Check results, in the same order as the request host list |

**data.result[n]**

| Parameter | Type | Description |
|-----------|------|-------------|
| bk_host_id | int64 | Host ID, -1 for new registration |
| status | string | Check status, see status values below |
| matched | object | Matched existing host info, null if no match |

**data.result[n].status values**

| Value | Description |
|-------|-------------|
| register_to_cmdb_and_install | Must register to CMDB before installation |
| normal_install | Ready for direct installation |
| host_not_found | Host does not exist |
| network_unit_not_found | Network unit does not exist |
| duplicated_inner_ip | Inner IPv4 address already in use |
| duplicated_inner_ip_v6 | Inner IPv6 address already in use |
| mismatched_inner_ip | Inner IPv4 address does not match |
| mismatched_inner_ip_v6 | Inner IPv6 address does not match |

**data.result[n].matched**

| Parameter | Type | Description |
|-----------|------|-------------|
| bk_host_id | int64 | Matched host ID |
| bk_networkunit_id | int64 | Network unit ID of matched host |
| bk_host_innerip_list | array[string] | Inner IPv4 address list of matched host |
| bk_host_innerip_v6_list | array[string] | Inner IPv6 address list of matched host |

**error**

| Parameter | Type | Description |
|-----------|------|-------------|
| system | string | Error source system |
| message | string | Error message |
| details | array | Detailed error list |

**error.details[n]**

| Parameter | Type | Description |
|-----------|------|-------------|
| code | string | Error code |
| message | string | Error detail |
