### Description

- API Version: v3.0.1-alpha.1+.
- Required Permission: plugin_view (View Plugin).
- Function: Query the process list with pagination, count-only mode, and exact or fuzzy filters.

### URL

POST /api/v3/process/list

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| page | object | No | Pagination parameters; can be omitted when `only_count=true` |
| only_count | bool | No | Whether to return only the matched count; when `true`, `items` is an empty array |
| exact_include_conditions | object | No | Exact-match include conditions |
| fuzzy_include_conditions | object | No | Fuzzy-match include conditions |

#### page

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| offset | int32 | No | Pagination offset; a value below `0` is treated as `0` |
| limit | int32 | No | Maximum number of returned records; valid range is `1` to `500` when querying items |

#### exact_include_conditions

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| bk_host_id | int64 array | No | Host ID list |
| bk_biz_id | int64 array | No | Business ID list |
| plugin_group | string array | No | Plugin group list; `default` is common but is not a fixed enum |
| generation | int64 array | No | Process generation list |
| platform_os | string array | No | Operating system type list |
| platform_arch | string array | No | CPU architecture list |
| status | string array | No | Process status list. Available values: `init`, `running`, `stopped`, `unregister`, `unknown` |
| agent_id | string array | No | Agent ID list |
| version | string array | No | Process version list |
| plugin_name | string array | No | Plugin name list |
| plugin_pkg_name | string array | No | Plugin package name list |

#### fuzzy_include_conditions

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| name | string array | No | Process binary name list, matched fuzzily |
| plugin_pkg_name | string array | No | Plugin package name list, matched fuzzily |

**Parameter Notes**:

- Different fields in one condition object are combined with AND; values in one array form the candidate set for that field.
- Both `exact_include_conditions` and `fuzzy_include_conditions` are inclusive filters. This API does not expose exclude filters.
- If `bk_biz_id` is omitted, the query is narrowed to businesses where the caller has `plugin_view`; if supplied, only authorized businesses in the list are queried.
- Defined `platform_os` values include `aix`, `aix6`, `aix7`, `android`, `darwin`, `dragonfly`, `freebsd`, `hurd`, `illumos`, `ios`, `js`, `linux`, `netbsd`, `openbsd`, `plan9`, `solaris`, `wasip1`, `windows`, `zos`, and `unknown`.
- Defined `platform_arch` values include `386`, `arm`, `arm64`, `amd64`, `loong64`, `mips`, `mipsle`, `mips64`, `mips64le`, `ppc`, `ppc64`, `ppc64le`, `riscv`, `riscv64`, `s390`, `s390x`, `sparc`, `sparc64`, `wasm`, and `unknown`.

### Request Example

Query running `bk-monitor-agent` processes in business `2`, returning the first 20 records.

```json
{
  "page": {
    "offset": 0,
    "limit": 20
  },
  "only_count": false,
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "plugin_name": ["bk-monitor-agent"],
    "status": ["running"]
  }
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260717-000001",
  "error": null,
  "permission": null,
  "data": {
    "total": 1,
    "items": [
      {
        "tenant_id": "default",
        "bk_host_id": 101,
        "bk_biz_id": 2,
        "plugin_name": "bk-monitor-agent",
        "plugin_pkg_name": "bkmonitoragent",
        "plugin_group": "default",
        "platform": {
          "os_type": "linux",
          "cpu_arch": "amd64"
        },
        "generation": 1,
        "process_info": {
          "pid": 2345,
          "version": "3.6.0",
          "agent_id": "agent-101",
          "auto_start": true,
          "status": "running"
        },
        "process_identity": {
          "name": "bk-monitor-agent",
          "setup_path": "/usr/local/gse/plugins/bin",
          "pid_path": "/var/run/bk-monitor-agent.pid",
          "config_path": "/usr/local/gse/plugins/etc",
          "log_path": "/var/log/gse",
          "user": "root"
        },
        "process_controller": {
          "start_cmd": "./start.sh",
          "stop_cmd": "./stop.sh",
          "restart_cmd": "./restart.sh",
          "reload_cmd": "./reload.sh",
          "kill_cmd": "./stop.sh --force",
          "version_cmd": "./bk-monitor-agent --version",
          "health_cmd": "./health.sh"
        },
        "process_resource": {
          "cpu_limit_percent": 80,
          "mem_limit_percent": 70
        },
        "process_monitor_policy": {
          "restart_type": "auto",
          "start_check_seconds": 10,
          "stop_check_seconds": 10,
          "operate_timeout_seconds": 60
        }
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
| total | int64 | Total number of processes matching the conditions |
| items | object array | Process list; an empty array when `only_count=true` |

#### data.items[n]

| Parameter | Type | Description |
|---------|----------|------|
| tenant_id | string | Tenant ID |
| bk_host_id | int64 | Host ID |
| bk_biz_id | int64 | Business ID |
| plugin_name | string | Plugin name |
| plugin_pkg_name | string | Plugin package name |
| plugin_group | string | Plugin group |
| platform | object | Runtime platform |
| generation | int64 | Process generation |
| process_info | object | Process runtime information |
| process_identity | object | Process identity and path information |
| process_controller | object | Process control commands |
| process_resource | object | Process resource limits |
| process_monitor_policy | object | Process monitoring policy |

#### platform

| Parameter | Type | Description |
|---------|----------|------|
| os_type | string | Operating system type |
| cpu_arch | string | CPU architecture |

#### process_info

| Parameter | Type | Description |
|---------|----------|------|
| pid | int32 | Process ID |
| version | string | Process version |
| agent_id | string | Agent ID |
| auto_start | bool | Whether the process starts automatically |
| status | string | Process status. Available values: `init`, `running`, `stopped`, `unregister`, `unknown` |

#### process_identity

| Parameter | Type | Description |
|---------|----------|------|
| name | string | Process binary name |
| setup_path | string | Installation path |
| pid_path | string | PID file path |
| config_path | string | Configuration path |
| log_path | string | Log path |
| user | string | Runtime user |

#### process_controller

| Parameter | Type | Description |
|---------|----------|------|
| start_cmd | string | Start command |
| stop_cmd | string | Stop command |
| restart_cmd | string | Restart command |
| reload_cmd | string | Reload command |
| kill_cmd | string | Force-stop command |
| version_cmd | string | Version query command |
| health_cmd | string | Health check command |

#### process_resource

| Parameter | Type | Description |
|---------|----------|------|
| cpu_limit_percent | double | CPU usage limit percentage |
| mem_limit_percent | double | Memory usage limit percentage |

#### process_monitor_policy

| Parameter | Type | Description |
|---------|----------|------|
| restart_type | string | Restart type. Available values: `auto`, `manual` |
| start_check_seconds | int64 | Wait time before the start check, in seconds |
| stop_check_seconds | int64 | Wait time before the stop check, in seconds |
| operate_timeout_seconds | int64 | Operation timeout, in seconds |

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
