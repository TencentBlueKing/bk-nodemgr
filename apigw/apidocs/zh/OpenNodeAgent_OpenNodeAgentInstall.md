### 描述

- 该接口提供版本：v3.0.1-alpha.105+；`is_sudo` 从 v3.0.1-alpha.106 起支持。
- 该接口所需权限：system_access（系统访问）、agent_operate（操作 Agent）、networkunit_use_for_agent（使用管控单元部署 Agent）。
- 该接口功能描述：以用户身份批量安装节点 Agent，支持按平台指定目标版本及手动安装。
- 用户身份由 API 网关认证并传递，请求体不需要用户名或租户参数；`login_user` 是目标主机登录账号，不是调用用户。
- `agent_operate` 按请求中的业务校验，`networkunit_use_for_agent` 按目标管控单元校验。

### URL

POST /api/v3/open/node/agent/install

### 输入参数

| 参数名称       | 参数类型     | 必选 | 描述                                                                         |
| -------------- | ------------ | ---- | ---------------------------------------------------------------------------- |
| host           | object array | 是   | 安装主机列表，不能为空；不是 `info`                                          |
| target_version | object array | 否   | 按 `os_type` + `cpu_arch` 指定版本；同一平台不得重复。省略时自动选择默认版本 |
| is_manual      | bool         | 否   | 手动安装模式，默认 `false`；仍须提供下表中的必填主机字段                     |

#### host[n]

| 参数名称                    | 参数类型     | 必选     | 描述                                                                                                     |
| --------------------------- | ------------ | -------- | -------------------------------------------------------------------------------------------------------- |
| bk_addressing               | string       | 是       | 寻址方式：`dynamic`、`static`                                                                            |
| bk_biz_id                   | int64        | 是       | 业务 ID，必须大于等于 `0`                                                                                |
| bk_host_innerip             | string array | 否       | 内网 IPv4 列表；与 `bk_host_innerip_v6` 至少一个非空；元素须为有效 IP                                    |
| bk_host_innerip_v6          | string array | 否       | 内网 IPv6 列表；与 `bk_host_innerip` 至少一个非空；元素须为有效 IP                                       |
| login_ip                    | string       | 是       | 登录目标主机的 IP，不能为空                                                                              |
| login_port                  | int64        | 是       | 登录端口，必须大于等于 `0`；例如 SSH 使用 `22`                                                           |
| login_user                  | string       | 是       | 目标主机登录账号，不能为空                                                                               |
| login_mode                  | string       | 是       | 凭证类型：`password_vault`（密码库）、`password`（密码）、`keyfile`（密钥文件）                          |
| login_password              | string       | 条件必选 | 自动安装且使用 `password` 时提供加密密码；已有可用凭证时可省略并复用；不是明文密码                       |
| login_key_file              | string       | 条件必选 | 自动安装且使用 `keyfile` 时，先将密钥文件内容 Base64 编码，再加密；已有可用凭证时可省略并复用            |
| bk_networkunit_id           | int64        | 是       | 已存在的目标管控单元 ID，必须大于等于 `0`                                                                |
| os_type                     | string       | 是       | 操作系统，不能为空；例如 `linux`、`windows`、`darwin`，实际安装须有对应平台的软件包                      |
| bk_host_id                  | int64        | 否       | 已有主机 ID；省略时为 `-1`，表示未指定已有主机                                                           |
| re_register                 | bool         | 否       | 是否重新注册，默认 `false`                                                                               |
| install_pre_ordered_plugins | bool         | 否       | 是否安装预设插件，默认 `false`                                                                           |
| renew_gse_task              | bool         | 否       | 是否重新生成 GSE `.task` runtime 文件；默认 `false`，保留已有文件                                        |
| renew_gse_proc              | bool         | 否       | 是否重新生成 GSE `.proc` runtime 文件；默认 `false`，保留已有文件                                        |
| install_method              | string       | 否       | `""`（默认，按 OS 自动选择）、`ssh`、`wmi`；`wmi` 仅 Windows 可用                                        |
| is_sudo                     | bool         | 否       | 非 root 登录且配置免密 sudo 时设为 `true`，通过 `sudo -n` 执行安装器；默认 `false`，仅对类 Unix 系统生效 |

自动安装的凭证必须使用当前部署启用的凭证公钥和加密协议生成密文；不要直接发送明文密码或密钥。手动模式不处理登录凭证，但不会免除必填字段校验。手动模式提交后仍需获取并执行安装命令，本接口不直接返回命令。

`install_method=""` 时，Linux/Darwin 使用 SSH；直接 Windows Agent 安装先尝试 SSH，在启动安装器前 SSH 准备失败时回退 WMI。显式指定 `ssh` 或 `wmi` 不执行该自动回退。

#### target_version[n]

| 参数名称 | 参数类型 | 必选 | 描述                                                           |
| -------- | -------- | ---- | -------------------------------------------------------------- |
| version  | string   | 是   | 目标 Agent 版本，应已上传并可用于目标平台                      |
| cpu_arch | string   | 是   | 目标平台 CPU 架构，例如 `amd64`、`arm64`；须与探测到的平台一致 |
| os_type  | string   | 是   | 目标平台操作系统，例如 `linux`、`windows`、`darwin`            |

### 调用示例

使用已配置的密码库安装 Linux Agent。

```json
{
  "host": [
    {
      "bk_addressing": "static",
      "bk_biz_id": 100,
      "bk_host_innerip": ["10.0.0.1"],
      "login_ip": "10.0.0.1",
      "login_port": 22,
      "login_user": "root",
      "login_mode": "password_vault",
      "bk_networkunit_id": 1,
      "os_type": "linux",
      "install_method": "ssh"
    }
  ],
  "target_version": [
    { "version": "2.1.0", "cpu_arch": "amd64", "os_type": "linux" }
  ],
  "is_manual": false
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": { "workflow_id": "workflow-abc123" }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                                     |
| ---------- | -------- | ---------------------------------------- |
| code       | int32    | 状态码，`0` 表示请求成功，不表示安装完成 |
| message    | string   | 请求信息                                 |
| request_id | string   | 请求 ID                                  |
| error      | object   | 错误信息                                 |
| permission | object   | 权限信息                                 |
| data       | object   | 响应数据                                 |

#### data

| 参数名称    | 参数类型 | 描述                                                               |
| ----------- | -------- | ------------------------------------------------------------------ |
| workflow_id | string   | 工作流 ID；用于跟踪异步操作，不是 `trigger_id`，返回时安装尚未完成 |

v3.0.1-alpha.114 起，可使用 `OpenNodeWorkflow_OpenNodeWorkflowList`，按 `exact_include_conditions.workflow_id` 查询执行状态；查询另需对应的历史查看权限。
