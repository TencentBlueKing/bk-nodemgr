### Description

- API Version: v3.0.1+.
- Required Permission: proxy_operate (Operate Proxy).
- Function: Batch upgrade node Proxy to the specified version.

### URL

POST /api/v3/node/proxy/upgrade

### Input Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| host | array | Yes | Host list, see host parameters below |
| target_version | array | No | Target version list, see target_version parameters below |

**host[n]**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| bk_host_id | int64 | Yes | Host ID |
| force | bool | No | Force upgrade, default false |
| graceful_restart_timeout_sec | int64 | No | Graceful restart timeout in seconds |

**target_version[n]**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| version | string | Yes | Version string |
| cpu_arch | string | Yes | CPU architecture: 386 / arm / arm64 / amd64 |
| os_type | string | Yes | OS type: linux / windows / darwin |

### Call Example

```json
{
    "host": [
        {
            "bk_host_id": 1001,
            "force": false,
            "graceful_restart_timeout_sec": 60
        }
    ],
    "target_version": [
        {
            "version": "2.1.0",
            "cpu_arch": "amd64",
            "os_type": "linux"
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
        "workflow_id": "wf-20240101-002"
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
