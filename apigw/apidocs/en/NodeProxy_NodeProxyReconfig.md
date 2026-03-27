### Description

- API Version: v3.0.1+.
- Required Permission: proxy_operate (Operate Proxy).
- Function: Batch reconfigure node Proxy.

### URL

POST /api/v3/node/proxy/reconfig

### Input Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| host | array | Yes | Host list, see host parameters below |

**host[n]**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| bk_host_id | int64 | Yes | Host ID |
| force | bool | No | Force reconfigure, default false |
| graceful_restart_timeout_sec | int64 | No | Graceful restart timeout in seconds |

### Call Example

```json
{
    "host": [
        {
            "bk_host_id": 1001,
            "force": false,
            "graceful_restart_timeout_sec": 60
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
        "workflow_id": "wf-20240101-004"
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
| workflow_id | string | Workflow ID |

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
