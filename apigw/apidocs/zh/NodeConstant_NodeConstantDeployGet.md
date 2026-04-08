### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：。
- 该接口功能描述：根据节点代际和操作系统类型，获取节点安装器、Agent 和插件的默认部署配置。

### URL

POST /api/v3/node/constant/deploy/get

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| generation | int64 | 是 | 节点代际。当前仅支持 `2`，传入 `1` 或其他值会返回参数错误 |
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

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，`0` 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时通常为空 |
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
| base_work_dir | string | 安装器工作目录基础路径 |

#### data.default_deploy_config.node_runtime

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| base_deploy_dir | string | Agent 部署目录基础路径 |
| data_ipc | string | Agent 数据 IPC 路径或端口。仅 Windows 默认配置会返回该字段 |
| plugin_ipc | string | Agent 插件 IPC 路径或端口。仅 Windows 默认配置会返回该字段 |
| log_dir | string | Agent 日志目录 |

#### data.default_deploy_config.plugin_runtime

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| base_deploy_dir | string | 插件部署目录基础路径 |
| log_dir | string | 插件日志目录 |

### 说明

- 接口契约来源于 `proto/backend/api/v3/node_constant.proto`、`pkg/proto/backend/api/v3/constant.go` 和 `docs/api/swagger/backend/api/v3/node_constant.swagger.json`。
- `generation` 在当前实现中只接受 `2`；`1` 会因已不再支持而返回参数错误。
- `os_type` 不仅要通过枚举校验，还必须在后台已加载对应代际的部署常量；否则会返回参数错误。
- `data_ipc` 和 `plugin_ipc` 只会在 `os_type=windows` 时按默认端口 `27000` 和 `26000` 返回。非 Windows 系统通常不会返回这两个字段。
