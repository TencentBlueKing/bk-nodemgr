### 描述

- 该接口提供版本：v3.0.1-alpha.119+。
- 该接口所需权限：system_access（系统访问）和 platform_manage（平台管理），两者均须满足。平台管理资源类型为 `platform_resource`，资源 ID 为 `system_config`（系统配置）。
- 该接口功能描述：通过应用态接口查询命中的 Backend 实例在启动时加载的业务配置白名单，不返回凭据、连接端点、日志或追踪配置，不提供修改能力。
- 调用须具备 API 网关应用授权并通过应用态认证；无需用户态认证，但应用态认证链仍要求有效的用户名和租户上下文，不能通过请求体指定身份。本接口与用户态入口返回相同结构的业务配置视图，执行相同的权限检查。
- 查询结果是进程级配置，不按请求租户过滤，也不表示所有 Backend 副本配置一致。本接口不查询数据库中的全局配置。
- 默认事件 Data ID 不解析业务级覆盖；安装默认设置不代表目标主机实际状态，网络单元等配置或单次部署设置可能覆盖这些默认值；默认直连网络单元配置不代表网络单元已创建。
- 安装配置中的目录和文件路径不是解析后的实际路径。仅返回安装 HTTP 代理策略和防火墙功能，不返回代理地址或防火墙允许访问的 IP/CIDR。

### URL

POST /api/v3/system_config/get

### 输入参数

无业务参数，请发送 JSON 对象 `{}`。

### 调用示例

```json
{}
```

### 响应示例

```json
{
  "code": 0,
  "message": "OK",
  "request_id": "req-123456",
  "data": {
    "tenant_mode": "single",
    "system": { "env": "dev", "edition": "ce" },
    "iam": { "v3_enabled": false, "v4_enabled": false },
    "node_event_data_id": {
      "agent_base_alarm_event_data_id": 1000,
      "task_proc_event_data_id": 1100008
    },
    "default_direct_network_unit": { "enabled": false, "name": "default" },
    "gse_deploy_configs": [
      {
        "generation": 2,
        "os_type": "linux",
        "base_work_dir": "/tmp/bknm/",
        "base_deploy_dir": "/usr/local/",
        "extra_config_dir": "",
        "host_id_path": "",
        "v2_run_dir": "",
        "v2_data_dir": "",
        "zone_id": "",
        "city_id": "",
        "http_proxy_mode": "inherit",
        "firewall_features": []
      }
    ]
  }
}
```

示例仅说明字段结构，不代表当前环境的实际配置。

### 响应参数说明

| 参数名称   | 参数类型 | 描述                     |
| ---------- | -------- | ------------------------ |
| code       | int32    | 状态码，`0` 表示成功     |
| message    | string   | 请求信息                 |
| request_id | string   | 请求 ID                  |
| error      | object   | 错误信息，成功时省略     |
| permission | object   | 权限申请信息，成功时省略 |
| data       | object   | Backend 业务配置视图     |

#### data

| 参数名称                    | 参数类型 | 描述                                                        |
| --------------------------- | -------- | ----------------------------------------------------------- |
| tenant_mode                 | string   | 租户模式：`single`（单租户）、`multiple`（多租户）           |
| system                      | object   | 部署环境及版本形态                                          |
| iam                         | object   | IAM 权限管理开关                                            |
| node_event_data_id          | object   | 全局默认节点事件 Data ID                                    |
| default_direct_network_unit | object   | 默认直连网络单元自动创建配置                                |
| gse_deploy_configs          | array    | 按 GSE 代际和操作系统配置的安装默认设置；没有配置时为空数组 |

#### data.system

| 参数名称 | 参数类型 | 描述                                                        |
| -------- | -------- | ----------------------------------------------------------- |
| env      | string   | 配置的部署环境                                              |
| edition  | string   | 版本形态：`ce`（社区版）、`ee`（企业版）、`inner`（内部版） |

#### data.iam

| 参数名称   | 参数类型 | 描述            |
| ---------- | -------- | --------------- |
| v3_enabled | bool     | 是否启用 IAM V3 |
| v4_enabled | bool     | 是否启用 IAM V4 |

#### data.node_event_data_id

| 参数名称                       | 参数类型 | 描述                              |
| ------------------------------ | -------- | --------------------------------- |
| agent_base_alarm_event_data_id | int64    | 全局默认 Agent 基础告警事件 Data ID |
| task_proc_event_data_id        | int64    | 全局默认任务进程事件 Data ID      |

#### data.default_direct_network_unit

| 参数名称 | 参数类型 | 描述                                             |
| -------- | -------- | ------------------------------------------------ |
| enabled  | bool     | 是否启用默认直连网络单元的自动创建               |
| name     | string   | 自动创建时使用的网络单元名称；关闭时仍返回配置值 |

#### data.gse_deploy_configs[n]

| 参数名称          | 参数类型     | 描述                                                                                                                             |
| ----------------- | ------------ | -------------------------------------------------------------------------------------------------------------------------------- |
| generation        | int64        | 配置的 GSE 代际，当前仅支持 `2`                                                                                                  |
| os_type           | string       | 配置的操作系统类型，例如 `linux`、`windows`、`aix`；配置值不代表该系统上一定支持实际安装                                          |
| base_work_dir     | string       | 安装默认工作目录                                                                                                                 |
| base_deploy_dir   | string       | 安装默认部署根目录                                                                                                               |
| extra_config_dir  | string       | Agent 扩展配置目录覆盖值；空字符串表示未配置覆盖值                                                                               |
| host_id_path      | string       | 插件读取的主机 ID 文件路径覆盖值；空字符串表示未配置覆盖值                                                                       |
| v2_run_dir        | string       | 插件 V2 兼容安装的运行目录覆盖值；空字符串表示未配置覆盖值                                                                       |
| v2_data_dir       | string       | 插件 V2 兼容安装的数据目录覆盖值；空字符串表示未配置覆盖值                                                                       |
| zone_id           | string       | 节点部署的地域标识覆盖值；空字符串表示未配置覆盖值                                                                               |
| city_id           | string       | 节点部署的城市标识覆盖值；空字符串表示未配置覆盖值                                                                               |
| http_proxy_mode   | string       | 安装 HTTP 代理策略：`inherit` 沿用主机环境代理，`direct` 强制直连，`specified` 使用指定代理；不返回代理地址                        |
| firewall_features | string array | Windows 安装器管理的防火墙功能，仅 Windows 安装器生效；支持 `gse`、`nodemgr`，空数组表示不管理规则；不返回允许访问的 IP/CIDR      |
