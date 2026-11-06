### Description

- API Version: v3.0.1-alpha.106+.
- Required Permission: plugin_view (View Plugin).
- Function: Query the plugin process status of hosts by business and host, returning the same content as the Node Manager V2 plugin search.
- Scope: Used only by the monitoring platform; not open for external application.

### URL

POST /api/v3/process/list_v2_compatible

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_biz_id | int64 array | Yes | Business ID list |
| bk_host_id | int64 array | No | Host ID list; when omitted, all hosts in the business are returned |

**Parameter Notes**:

- `bk_biz_id` is required; a missing or empty array returns a parameter error.
- This API is not paginated: one call returns every matched host within the authorized scope, and `data.total` is the number of matched hosts.
- `bk_biz_id` is narrowed by the caller's `plugin_view` business permission. If the caller is authorized for all businesses, the requested businesses are queried; with partial authorization only the authorized ones among them are queried; if none of the requested businesses is authorized, `total=0` and an empty `list` are returned without a permission error. If the caller has no business instance authorization at all, the API checks permission against the requested businesses and returns a permission error.
- `bk_host_id` filters hosts inside the narrowed business scope; if no host matches, `list` is empty.
- Only the `bkmonitorbeat` plugin process is returned, matching the Node Manager V2 plugin search scope: `plugin_status` always contains exactly one entry and never reports other plugins.

### Request Example

Query the plugin process status of hosts `101` and `102` in business `2`.

```json
{
  "bk_biz_id": [2],
  "bk_host_id": [101, 102]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20261106-000001",
  "error": null,
  "permission": null,
  "data": {
    "total": 2,
    "list": [
      {
        "bk_host_id": 101,
        "inner_ip": "10.0.0.1",
        "inner_ipv6": "",
        "bk_cloud_id": 0,
        "setup_path": "/usr/local/gse/gse_agent",
        "plugin_status": [
          {
            "name": "bkmonitorbeat",
            "version": "3.6.0",
            "status": "running"
          }
        ]
      },
      {
        "bk_host_id": 102,
        "inner_ip": "10.0.0.2,10.0.1.2",
        "inner_ipv6": "fd00::2",
        "bk_cloud_id": 0,
        "setup_path": "/usr/local/gse/gse_agent",
        "plugin_status": [
          {
            "name": "bkmonitorbeat",
            "version": "",
            "status": "unregister"
          }
        ]
      }
    ]
  }
}
```

### Response Parameters

| Parameter | Type | Description |
|---------|----------|------|
| code | int32 | Status code, `0` means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, usually empty on success |
| permission | object | Permission application information |
| data | object | Response data |

#### data

| Parameter | Type | Description |
|---------|----------|------|
| total | int64 | Total number of hosts matching the conditions |
| list | object array | Host list with plugin process status; an empty array when no host matches |

#### data.list[n]

| Parameter | Type | Description |
|---------|----------|------|
| bk_host_id | int64 | Host ID |
| inner_ip | string | Internal IPv4 addresses, comma-separated when there are several |
| inner_ipv6 | string | Internal IPv6 addresses, comma-separated when there are several; empty when the host has none |
| bk_cloud_id | int64 | Cloud area ID |
| setup_path | string | Agent installation path on the host, computed from the deployment configuration and the cloud area custom deployment configuration; empty when the host has no node installation information; computed from the second generation (generation=2) and `linux` deployment configuration when the reported node generation or os type is missing |
| plugin_status | object array | Plugin process status list, always containing exactly one entry |

#### plugin_status[n]

| Parameter | Type | Description |
|---------|----------|------|
| name | string | Plugin name, currently always `bkmonitorbeat` |
| version | string | Plugin version; empty when no process record exists for the plugin |
| status | string | Plugin process status. Available values: `init`, `running`, `stopped`, `unknown`, `unregister` |

**status values**:

- `running`: the process is running.
- `stopped`: the process is stopped.
- `unknown`: the process status cannot be determined.
- `init`: the process record has been created but its actual status has not been synchronized yet.
- `unregister`: the host has no process record for the plugin.

#### error

| Parameter | Type | Description |
|---------|----------|------|
| system | string | Error system identifier |
| message | string | Error message |
| details | object array | Error detail list |

#### error.details[n]

| Parameter | Type | Description |
|---------|----------|------|
| code | string | Error code |
| message | string | Error message |
