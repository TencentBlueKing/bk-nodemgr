### Description

- API Version: v3.0.1-alpha.60+.
- Required Permission: proxy_operate (Operate Proxy).
- Function: Batch-check whether Proxies can be upgraded.

### URL

POST /api/v3/node/proxy/upgrade_check

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| host | object array | Yes | Host information list |
| target_version | object array | No | Target version list |

#### host[n]

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_host_id | int64 | No | Host ID |
| bk_networkunit_id | int64 | No | Network unit ID |
| cpu_arch | string | No | CPU architecture |

#### target_version[n]

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| version | string | No | Version |
| cpu_arch | string | No | CPU architecture |
| os_type | string | No | Operating system type |

### Request Example

```json
{
  "host": [
    {
      "bk_host_id": 1,
      "bk_networkunit_id": 1,
      "cpu_arch": "amd64"
    }
  ],
  "target_version": [
    {
      "version": "3.2.1",
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
  "request_id": "req-123456",
  "error": null,
  "permission": null,
  "data": {
    "results": [
      {
        "status": "running",
        "matched": {
          "bk_host_id": 1,
          "bk_biz_id": 1,
          "bk_networkarea_id": 1,
          "bk_networkunit_id": 1,
          "os_type": "linux",
          "node_role": "agent",
          "bk_host_innerip_list": [
            "10.0.0.1"
          ],
          "bk_host_innerip_v6_list": [
            "2001:db8::1"
          ]
        }
      }
    ]
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, 0 means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, empty on success |
| permission | object | Permission information |
| data | object | Response data |

#### data

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| results | object array | Result list |

#### data.results[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| status | string | Status |
| matched | object | Matched host information |
