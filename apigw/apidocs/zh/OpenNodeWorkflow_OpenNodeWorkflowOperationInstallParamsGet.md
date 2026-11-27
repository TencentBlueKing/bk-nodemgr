### 描述

- 该接口提供版本：v3.0.1-alpha.121+。
- 该接口所需权限：system_access（系统访问）、agent_history_view（查看 Agent 历史）、proxy_history_view（查看 Proxy 历史）。根据历史任务流的节点角色，在任务流涉及的业务上校验对应权限；`blank`、`agent` 需要 Agent 历史权限，`proxy` 需要 Proxy 历史权限；涉及两者、角色缺失或未知时需要两者。
- 该接口功能描述：读取所选历史 Agent / Proxy 安装操作可恢复的参数，用于执行诊断（例如核对历史安装时的目标版本、登录认证方式及连接参数）；不创建安装任务、不重试操作，也不修改历史任务。
- 用户身份由 API 网关认证并传递，请求体不需要用户名或租户参数。

**使用边界**：

- 仅支持 `install_agent` 和 `install_proxy` 任务流，不要求操作必须处于失败状态。可先调用[操作列表接口](OpenNodeWorkflow_OpenNodeWorkflowOperationList.md)取得操作 ID。
- 认证方式可以复制，但响应不包含密码、私钥、密码库凭据 ID、部署 Token 或实例运行状态。
- 参数来自历史部署记录，并非完整的原始请求快照。`missing_fields` 标识未能恢复的字段；不能把这些字段的空值当成上次的选择。缺失整条部署数据的操作放入 `unavailable_operation_ids`，不出现在 `items` 中。
- 返回的主机 ID、业务、管控区域、管控单元和目标版本等历史信息不代表当前仍然有效。这些参数也不能证明实际执行时采用的配置或安装结果，需结合执行实例及步骤日志判断。

### URL

POST /api/v3/open/node/workflow/operation/install_params/get

### 输入参数

| 参数名称      | 参数类型     | 必选 | 描述                                                                                                                    |
| ------------- | ------------ | ---- | ----------------------------------------------------------------------------------------------------------------------- |
| workflow_id   | string       | 是   | 当前租户下的历史安装任务流 ID，不能为空或全为空白字符                                                                   |
| operation_ids | string array | 是   | 该任务流下选中的操作 ID；列表非空，元素不能为空或全为空白字符，且不能重复。不存在或不属于该任务流的操作会使整个请求失败 |

### 调用示例

```json
{
  "workflow_id": "wf:00000000000000000000000000000001",
  "operation_ids": ["oper:00000000000000000000000000000001", "oper:00000000000000000000000000000002"]
}
```

### 响应示例

示例中第一个操作可以恢复部分参数；第二个操作的部署数据已不可用。

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "type": "install_agent",
    "items": [
      {
        "operation_id": "oper:00000000000000000000000000000001",
        "host": {
          "bk_host_id": 1001,
          "bk_biz_id": 2,
          "bk_networkarea_id": 0,
          "bk_networkunit_id": 1,
          "bk_addressing": "static",
          "bk_host_innerip": ["10.0.0.1"],
          "bk_host_innerip_v6": [],
          "os_type": "linux",
          "login_ip": "10.0.0.1",
          "login_user": "root",
          "login_mode": "password",
          "install_method": "",
          "login_port": 22,
          "is_sudo": false,
          "re_register": false,
          "install_pre_ordered_plugins": true,
          "renew_gse_task": false,
          "renew_gse_proc": false,
          "export_ip": "",
          "export_ip_v6": "",
          "advertise_ip": "",
          "advertise_ip_v6": "",
          "proxy_tags": [],
          "proxy_install_origin_unit_id": 0,
          "relay_download_port": 0,
          "relay_callback_port": 0,
          "cpu_arch": "x86_64"
        },
        "target_version": [
          { "os_type": "linux", "cpu_arch": "x86_64", "version": "2.0.0" }
        ],
        "is_manual": false,
        "is_offline": false,
        "missing_fields": ["bk_host_innerip_v6", "install_method"]
      }
    ],
    "unavailable_operation_ids": ["oper:00000000000000000000000000000002"]
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                             |
| ---------- | -------- | -------------------------------- |
| code       | int32    | 状态码，`0` 表示成功             |
| message    | string   | 请求信息                         |
| request_id | string   | 请求 ID                          |
| error      | object   | 错误信息，成功时通常为空         |
| permission | object   | 权限不足时的权限申请信息         |
| data       | object   | 可恢复的安装参数及不可用操作列表 |

#### data

| 参数名称                  | 参数类型     | 描述                                                                                                            |
| ------------------------- | ------------ | --------------------------------------------------------------------------------------------------------------- |
| type                      | string       | 历史任务流类型：`install_agent`、`install_proxy`                                                                |
| items                     | object array | 可恢复的操作参数，按 `operation_ids` 的请求顺序返回，排除不可用操作；可以为空                                   |
| unavailable_operation_ids | string array | 操作存在且属于该任务流，但部署记录不存在或主机静态、动态信息缺失的操作 ID，按请求顺序返回。为空表示没有此类操作 |

#### items[n]

| 参数名称       | 参数类型     | 描述                                                                 |
| -------------- | ------------ | -------------------------------------------------------------------- |
| operation_id   | string       | 历史操作 ID，用于与请求中的操作对应                                  |
| host           | object       | 可复制的历史主机安装参数，不包含凭据内容                             |
| target_version | object array | 历史安装目标版本列表，不是当前默认版本，也不是安装完成后检测到的版本 |
| is_manual      | bool         | 历史是否选择手动安装                                                 |
| is_offline     | bool         | 历史是否选择离线安装                                                 |
| missing_fields | string array | 未恢复的字段名，见下方说明                                           |

#### items[n].host

| 参数名称                     | 参数类型     | 描述                                                                                                                                  |
| ---------------------------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------- |
| bk_host_id                   | int64        | 历史主机 ID，仅作为历史信息；新安装请求中的主机标识按普通安装规则处理                                                                 |
| bk_biz_id                    | int64        | 历史业务 ID                                                                                                                           |
| bk_networkarea_id            | int64        | 历史管控区域 ID；`0` 是有效 ID，不能当作缺失                                                                                          |
| bk_networkunit_id            | int64        | 历史管控单元 ID                                                                                                                       |
| bk_addressing                | string       | 寻址方式：`dynamic`、`static`                                                                                                         |
| bk_host_innerip              | string array | 历史内网 IPv4 地址列表                                                                                                                |
| bk_host_innerip_v6           | string array | 内网 IPv6 地址列表；当前历史记录未保存此字段，见 `missing_fields`                                                                     |
| os_type                      | string       | 历史操作系统类型，如 `linux`、`windows`、`darwin`                                                                                     |
| login_ip                     | string       | 登录地址                                                                                                                              |
| login_user                   | string       | 登录用户                                                                                                                              |
| login_mode                   | string       | 认证方式：`password`（密码）、`keyfile`（密钥）、`password_vault`（密码库）；不含凭据内容                                             |
| install_method               | string       | 安装连接方式，普通安装允许空字符串（按 OS 自动选择）、`ssh`、`wmi`；历史返回空字符串时会标记为缺失，不能据此认定上次选择了自动方式    |
| login_port                   | int64        | 登录端口                                                                                                                              |
| is_sudo                      | bool         | 是否使用 sudo                                                                                                                         |
| re_register                  | bool         | 是否重新注册主机                                                                                                                      |
| install_pre_ordered_plugins  | bool         | 是否安装预订阅插件                                                                                                                    |
| renew_gse_task               | bool         | 是否重新生成 GSE `.task` runtime 文件                                                                                                 |
| renew_gse_proc               | bool         | 是否重新生成 GSE `.proc` runtime 文件                                                                                                 |
| export_ip                    | string       | 对外服务 IPv4 地址                                                                                                                    |
| export_ip_v6                 | string       | 对外服务 IPv6 地址                                                                                                                    |
| advertise_ip                 | string       | 广播 IPv4 地址                                                                                                                        |
| advertise_ip_v6              | string       | 广播 IPv6 地址                                                                                                                        |
| proxy_tags                   | string array | Proxy 标签：`dedicated_installer`（专用安装通道）、`cluster_tunnel`（集群通道）、`file_tunnel`（文件通道）、`data_tunnel`（数据通道） |
| proxy_install_origin_unit_id | int64        | Proxy 安装来源管控单元 ID                                                                                                             |
| relay_download_port          | int64        | Relay 下载端口                                                                                                                        |
| relay_callback_port          | int64        | Relay 回调端口                                                                                                                        |
| cpu_arch                     | string       | 历史主机 CPU 架构，如 `x86_64`                                                                                                        |

#### items[n].target_version[n]

| 参数名称 | 参数类型 | 描述                                                              |
| -------- | -------- | ----------------------------------------------------------------- |
| os_type  | string   | 目标版本对应的操作系统类型                                        |
| cpu_arch | string   | 目标版本对应的 CPU 架构；保留历史标识，不应自行改写为其他架构别名 |
| version  | string   | 历史指定的目标版本号；提交新安装前仍需确认该版本当前可用          |

#### missing_fields

当前可能返回以下字段名：

| 字段名             | 含义                                                                                 |
| ------------------ | ------------------------------------------------------------------------------------ |
| bk_host_innerip_v6 | 当前历史部署记录未保存内网 IPv6 列表，因此始终标记；空列表不证明原请求没有 IPv6 地址 |
| install_method     | 未恢复安装连接方式，需要用户确认，不能将空字符串当作已恢复的自动方式选择             |
| target_version     | 未恢复历史目标版本，需要重新选择；接口不会替换成最新版本                             |

`missing_fields` 不表示凭据填写状态，也不是对所有历史参数完整性的保证。新安装前仍需执行正常表单校验。

### 错误场景

- 必填 ID 为空、操作列表为空或包含空白、重复 ID：请求参数校验失败。
- 无对应历史查看权限：返回权限错误。
- 任务流不存在、不属于当前租户、不是 Agent / Proxy 安装类型，或操作不存在、不属于指定任务流：整个请求失败，不作为部分成功返回。
- 无法读取操作参数、缺少部署关联信息或发生存储错误：整个请求失败。部署记录不可用属于前述 `unavailable_operation_ids` 情况，可与其他可恢复操作一起成功返回。
