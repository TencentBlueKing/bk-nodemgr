### Description

- API Version: v3.0.1+.
- Required Permission: .
- Function: Get the default deployment configuration for the installer, node agent, and plugins by node generation and operating system type.

### URL

POST /api/v3/node/constant/deploy/get

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| generation | int64 | Yes | Node generation. Only `2` is currently supported. Passing `1` or any other value returns an invalid parameter error |
| os_type | string | Yes | Node operating system type. It must be a valid `os_type` and must have deploy constants loaded for the given `generation`. Common values include `linux` and `windows` |

### Request Example

Query the default deployment configuration for a generation-2 Linux node.

```json
{
  "generation": 2,
  "os_type": "linux"
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "default_deploy_config": {
      "installer_runtime": {
        "base_work_dir": "/data/bknodeman/workdir"
      },
      "node_runtime": {
        "base_deploy_dir": "/usr/local/gse",
        "log_dir": "/var/log/bk-gse/"
      },
      "plugin_runtime": {
        "base_deploy_dir": "/usr/local/gse",
        "log_dir": "/var/log/bk-gse/plugin/"
      }
    }
  }
}
```

### Response Parameters

| Parameter | Type | Description |
|---------|----------|------|
| code | int32 | Status code. `0` means success |
| message | string | Response message |
| request_id | string | Request ID |
| error | object | Error information. Usually empty on success |
| data | object | Default deployment configuration |

#### error

| Parameter | Type | Description |
|---------|----------|------|
| system | string | Source system of the error |
| message | string | Error message |
| details | object array | Error detail list |

#### error.details[n]

| Parameter | Type | Description |
|---------|----------|------|
| code | string | Error code |
| message | string | Error detail |

#### data

| Parameter | Type | Description |
|---------|----------|------|
| default_deploy_config | object | Default deployment configuration |

#### data.default_deploy_config

| Parameter | Type | Description |
|---------|----------|------|
| installer_runtime | object | Installer runtime configuration |
| node_runtime | object | Node runtime configuration |
| plugin_runtime | object | Plugin runtime configuration |

#### data.default_deploy_config.installer_runtime

| Parameter | Type | Description |
|---------|----------|------|
| base_work_dir | string | Base working directory for the installer |

#### data.default_deploy_config.node_runtime

| Parameter | Type | Description |
|---------|----------|------|
| base_deploy_dir | string | Base deployment directory for the agent |
| data_ipc | string | Agent data IPC path or port. This field is only returned for Windows default config |
| plugin_ipc | string | Agent plugin IPC path or port. This field is only returned for Windows default config |
| log_dir | string | Agent log directory |

#### data.default_deploy_config.plugin_runtime

| Parameter | Type | Description |
|---------|----------|------|
| base_deploy_dir | string | Base deployment directory for plugins |
| log_dir | string | Plugin log directory |

### Notes

- The API contract is defined by `proto/backend/api/v3/node_constant.proto`, `pkg/proto/backend/api/v3/constant.go`, and `docs/api/swagger/backend/api/v3/node_constant.swagger.json`.
- In the current implementation, `generation` only accepts `2`; `1` returns an invalid parameter error because generation 1 is no longer supported.
- `os_type` must pass enum validation and must also have deploy constants loaded in the backend for the given generation. Otherwise the API returns an invalid parameter error.
- `data_ipc` and `plugin_ipc` are only populated when `os_type=windows`, using default ports `27000` and `26000`. These fields are usually omitted for non-Windows systems.