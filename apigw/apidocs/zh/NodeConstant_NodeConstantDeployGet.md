### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：无。
- 该接口功能描述：根据节点代际和操作系统类型，获取节点安装器、Agent 和插件的默认部署配置。

### URL

POST /api/v3/node/constant/deploy/get

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| generation | int64 | 是 | 节点代际。当前仅支持 `2`；传入 `1` 或其他不支持的值会返回参数错误 |
| os_type | string | 是 | 节点操作系统类型。必须是合法的 `os_type`，且当前 `generation` 下已加载对应部署配置。常见值包括 `linux`、`windows` |

### 调用示例

查询第 2 代 Linux 节点的默认部署配置。

```json
{
  "generation": 2,
  "os_type": "linux"
}
```

### 响应示例

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

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，`0` 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时通常为空 |
| permission | object | 权限信息，当前接口通常为空 |
| data | object | 默认部署配置 |

#### error

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| system | string | 错误来源系统 |
| message | string | 错误消息 |
| details | object array | 错误详情列表 |

#### error.details[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | string | 错误码 |
| message | string | 错误详情 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| default_deploy_config | object | 默认部署配置 |

#### data.default_deploy_config

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| installer_runtime | object | 安装器运行时配置 |
| node_runtime | object | 节点运行时配置 |
| plugin_runtime | object | 插件运行时配置 |

#### data.default_deploy_config.installer_runtime

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| base_work_dir | string | 安装器工作目录基础路径，来源于 backend `gseDeployConfs[*].baseWorkDir` |

#### data.default_deploy_config.node_runtime

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| base_deploy_dir | string | Agent 部署目录基础路径，来源于 backend `gseDeployConfs[*].baseDeployDir` |
| data_ipc | string | Agent 数据 IPC。未自定义时，Windows 返回端口，Unix/Linux 返回包含 `{node_role}` 的路径 |
| plugin_ipc | string | Agent 插件 IPC。未自定义时，Windows 返回端口，Unix/Linux 返回包含 `{node_role}` 的路径 |
| log_dir | string | Agent 日志目录 |
| zone_id | string | GSE 节点可用区 ID，未自定义时为 `default` |
| city_id | string | GSE 节点城市 ID，未自定义时为 `default` |

#### data.default_deploy_config.plugin_runtime

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| base_deploy_dir | string | 插件部署目录基础路径，来源于 backend `gseDeployConfs[*].baseDeployDir` |
| log_dir | string | 插件日志目录 |

### 说明

- 接口契约来源于 `proto/backend/api/v3/node_constant.proto`、`proto/backend/api/v3/common.proto`、`pkg/proto/backend/api/v3/constant.go` 和 `docs/api/swagger/backend/api/v3/node_constant.swagger.json`。
- 运行时行为定义在 `internal/backend/router/api-v3/node/constant/constant.go`：接口根据 `generation` 和 `os_type` 读取已加载的节点与插件部署配置，并返回 `data.default_deploy_config`。
- `generation` 在当前实现中只接受 `2`；`1` 会因已不再支持而返回参数错误。
- `os_type` 需要通过枚举校验，并且后台必须已加载对应代际和操作系统类型的部署配置；否则会返回参数错误。
- 返回值受 backend `gseDeployConfs` 配置和部署环境影响。未配置自定义 `dataIPC`、`pluginIPC`、`logDir`、`zoneID`、`cityID` 时，服务会按默认规则生成。
- `data_ipc` 和 `plugin_ipc` 在 Windows 默认配置中分别返回端口 `27000` 和 `26000`；在 Unix/Linux 默认配置中返回 IPC 路径。
