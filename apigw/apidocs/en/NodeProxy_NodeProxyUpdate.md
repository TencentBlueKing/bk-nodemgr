### Description

- API Version: v3.0.1+.
- Required Permission: proxy_operate (Operate Proxy).
- Function: Batch update dynamic configuration of node Proxy, including login info, network addresses, and tags. This is a metadata-only update and does not trigger remote operations.

### URL

POST /api/v3/node/proxy/update

### Input Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| host | array | Yes | Host list, see host parameters below |

**host[n]**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| bk_host_id | int64 | Yes | Host ID |
| login_ip | string | No | Login IP |
| login_port | int64 | No | Login port |
| login_user | string | No | Login username |
| login_mode | string | No | Login mode: password_vault / password / keyfile |
| export_ip | string | No | Export IPv4 address |
| export_ip_v6 | string | No | Export IPv6 address |
| advertise_ip | string | No | Advertise IPv4 address |
| advertise_ip_v6 | string | No | Advertise IPv6 address |
| proxy_tags | array[string] | No | Proxy tags: dedicated_installer / cluster_tunnel / file_tunnel / data_tunnel |
| relay_download_port | int64 | No | Relay download port |
| relay_callback_port | int64 | No | Relay callback port |

### Call Example

```json
{
    "host": [
        {
            "bk_host_id": 1001,
            "login_ip": "10.0.0.1",
            "login_port": 22,
            "login_user": "root",
            "login_mode": "password",
            "export_ip": "10.0.0.1",
            "proxy_tags": ["cluster_tunnel"]
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
    "data": {}
}
```

### Response Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| code | int32 | Status code, 0 means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error info, null on success |
| data | object | Response data (empty object on success) |

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
