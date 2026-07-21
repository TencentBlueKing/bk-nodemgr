### Description

- API Version: v3.0.1+.
- Required Permission: None.
- Function: Get the default deployment configuration for the installer, node agent, and plugins by node generation and operating system type.

### URL

POST /api/v3/node/constant/deploy/get

### Request Parameters

| Parameter | Type | Required | Description |
|---------|----------|------|------|
| generation | int64 | Yes | Node generation. Only `2` is currently supported. Passing `1` or any other unsupported value returns an invalid parameter error |
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
        "base_work_dir": "/tmp/bknm/"
      },
      "node_runtime": {
        "base_deploy_dir": "/usr/local/",
        "data_ipc": "/usr/local/dev/{node_role}/lib/ipc.state.report",
        "plugin_ipc": "/usr/local/dev/{node_role}/lib/ipc.state.message",
        "log_dir": "/var/log/dev/",
        "zone_id": "default",
        "city_id": "default"
      },
      "plugin_runtime": {
        "base_deploy_dir": "/usr/local/",
        "log_dir": "/var/log/dev/plugin/"
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
| permission | object | Permission information, typically empty for this API |
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
| base_work_dir | string | Base working directory for the installer, from backend `gseDeployConfs[*].baseWorkDir` |

#### data.default_deploy_config.node_runtime

| Parameter | Type | Description |
|---------|----------|------|
| base_deploy_dir | string | Base deployment directory for the agent, from backend `gseDeployConfs[*].baseDeployDir` |
| data_ipc | string | Agent data IPC. When not customized, Windows returns a port, and Unix/Linux returns a path containing `{node_role}` |
| plugin_ipc | string | Agent plugin IPC. When not customized, Windows returns a port, and Unix/Linux returns a path containing `{node_role}` |
| log_dir | string | Agent log directory |
| zone_id | string | GSE node zone ID. Defaults to `default` when not customized |
| city_id | string | GSE node city ID. Defaults to `default` when not customized |

#### data.default_deploy_config.plugin_runtime

| Parameter | Type | Description |
|---------|----------|------|
| base_deploy_dir | string | Base deployment directory for plugins, from backend `gseDeployConfs[*].baseDeployDir` |
| log_dir | string | Plugin log directory |

### Notes

- The API contract is defined by `proto/backend/api/v3/node_constant.proto`, `proto/backend/api/v3/common.proto`, `pkg/proto/backend/api/v3/constant.go`, and `docs/api/swagger/backend/api/v3/node_constant.swagger.json`.
- Runtime behavior is implemented in `internal/backend/router/api-v3/node/constant/constant.go`: the API loads the node and plugin deployment configurations by `generation` and `os_type`, then returns `data.default_deploy_config`.
- In the current implementation, `generation` only accepts `2`; `1` returns an invalid parameter error because generation 1 is no longer supported.
- `os_type` must pass enum validation, and the backend must have deployment configuration loaded for the given generation and OS type. Otherwise, the API returns an invalid parameter error.
- Returned values depend on backend `gseDeployConfs` and the deployment environment. When custom `dataIPC`, `pluginIPC`, `logDir`, `zoneID`, or `cityID` values are not configured, the service derives them from default rules.
- In the Windows default configuration, `data_ipc` and `plugin_ipc` return ports `27000` and `26000`; in the Unix/Linux default configuration, they return IPC paths.
