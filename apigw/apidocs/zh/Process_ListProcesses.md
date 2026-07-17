### 描述

- 该接口提供版本：v3.0.1-alpha.1+。
- 该接口所需权限：plugin_view（查看插件）。
- 该接口功能描述：分页查询进程列表，支持仅统计总数，以及按精确或模糊条件过滤。

### URL

POST /api/v3/process/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| page | object | 否 | 分页参数；当 `only_count=true` 时可省略 |
| only_count | bool | 否 | 是否只返回匹配总数；为 `true` 时 `items` 为空数组 |
| exact_include_conditions | object | 否 | 精确匹配包含条件 |
| fuzzy_include_conditions | object | 否 | 模糊匹配包含条件 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| offset | int32 | 否 | 分页起始偏移量；小于 `0` 时按 `0` 处理 |
| limit | int32 | 否 | 返回条数上限；查询列表时取值范围为 `1` 至 `500` |

#### exact_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_host_id | int64 array | 否 | 主机 ID 列表 |
| bk_biz_id | int64 array | 否 | 业务 ID 列表 |
| plugin_group | string array | 否 | 插件分组列表，常见值为 `default`，但不是固定枚举 |
| generation | int64 array | 否 | 进程代际列表 |
| platform_os | string array | 否 | 操作系统类型列表 |
| platform_arch | string array | 否 | CPU 架构列表 |
| status | string array | 否 | 进程状态列表，可选值：`init`、`running`、`stopped`、`unregister`、`unknown` |
| agent_id | string array | 否 | Agent ID 列表 |
| version | string array | 否 | 进程版本列表 |
| plugin_name | string array | 否 | 插件名称列表 |
| plugin_pkg_name | string array | 否 | 插件包名称列表 |

#### fuzzy_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| name | string array | 否 | 进程二进制名称列表，按模糊条件匹配 |
| plugin_pkg_name | string array | 否 | 插件包名称列表，按模糊条件匹配 |

**参数说明**：

- 同一条件对象中的不同字段按 AND 关系组合；同一数组内的值作为该字段的候选集合。
- `exact_include_conditions` 与 `fuzzy_include_conditions` 均为包含型条件，当前接口未暴露排除型条件。
- 未传 `bk_biz_id` 时，查询范围会收敛到调用方有 `plugin_view` 权限的业务；已传时只查询其中有权限的业务。
- `platform_os` 的已定义值包括 `aix`、`aix6`、`aix7`、`android`、`darwin`、`dragonfly`、`freebsd`、`hurd`、`illumos`、`ios`、`js`、`linux`、`netbsd`、`openbsd`、`plan9`、`solaris`、`wasip1`、`windows`、`zos`、`unknown`。
- `platform_arch` 的已定义值包括 `386`、`arm`、`arm64`、`amd64`、`loong64`、`mips`、`mipsle`、`mips64`、`mips64le`、`ppc`、`ppc64`、`ppc64le`、`riscv`、`riscv64`、`s390`、`s390x`、`sparc`、`sparc64`、`wasm`、`unknown`。

### 调用示例

查询业务 `2` 下运行中的 `bk-monitor-agent` 进程，并返回前 20 条数据。

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

### 响应示例

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

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，`0` 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时通常为空 |
| permission | object | 权限申请信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| total | int64 | 匹配条件的进程总数 |
| items | object array | 进程列表；当 `only_count=true` 时为空数组 |

#### data.items[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| tenant_id | string | 租户 ID |
| bk_host_id | int64 | 主机 ID |
| bk_biz_id | int64 | 业务 ID |
| plugin_name | string | 插件名称 |
| plugin_pkg_name | string | 插件包名称 |
| plugin_group | string | 插件分组 |
| platform | object | 运行平台 |
| generation | int64 | 进程代际 |
| process_info | object | 进程运行信息 |
| process_identity | object | 进程身份与路径信息 |
| process_controller | object | 进程控制命令 |
| process_resource | object | 进程资源限制 |
| process_monitor_policy | object | 进程监控策略 |

#### platform

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| os_type | string | 操作系统类型 |
| cpu_arch | string | CPU 架构 |

#### process_info

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| pid | int32 | 进程 ID |
| version | string | 进程版本 |
| agent_id | string | Agent ID |
| auto_start | bool | 是否自动启动 |
| status | string | 进程状态，可选值：`init`、`running`、`stopped`、`unregister`、`unknown` |

#### process_identity

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| name | string | 进程二进制名称 |
| setup_path | string | 安装路径 |
| pid_path | string | PID 文件路径 |
| config_path | string | 配置文件路径 |
| log_path | string | 日志路径 |
| user | string | 运行用户 |

#### process_controller

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| start_cmd | string | 启动命令 |
| stop_cmd | string | 停止命令 |
| restart_cmd | string | 重启命令 |
| reload_cmd | string | 重载命令 |
| kill_cmd | string | 强制终止命令 |
| version_cmd | string | 版本查询命令 |
| health_cmd | string | 健康检查命令 |

#### process_resource

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| cpu_limit_percent | double | CPU 使用率限制百分比 |
| mem_limit_percent | double | 内存使用率限制百分比 |

#### process_monitor_policy

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| restart_type | string | 重启类型，可选值：`auto`、`manual` |
| start_check_seconds | int64 | 启动检查等待时间，单位为秒 |
| stop_check_seconds | int64 | 停止检查等待时间，单位为秒 |
| operate_timeout_seconds | int64 | 操作超时时间，单位为秒 |

#### error

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| system | string | 错误系统标识 |
| message | string | 错误消息 |
| details | object array | 错误详情列表 |

#### error.details[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | string | 错误代码 |
| message | string | 错误消息 |
