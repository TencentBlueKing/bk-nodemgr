### 描述

- 该接口提供版本：v3.0.1-alpha.105+
- 该接口所需权限：system_access（系统访问）、agent_operate（操作 Agent）。
- 该接口功能描述：以用户身份批量升级已有主机上的 Agent，并可指定升级后的管控单元。
- 用户身份由 API 网关认证并传递，请求体不需要用户名或租户参数。`agent_operate` 按目标主机所属业务校验；本接口不额外检查管控单元使用权限。

### URL

POST /api/v3/open/node/agent/upgrade

### 输入参数

| 参数名称 | 参数类型     | 必选 | 描述                                      |
| -------- | ------------ | ---- | ----------------------------------------- |
| host     | object array | 是   | 升级主机列表，不能为空；使用不同的主机 ID |

#### host[n]

| 参数名称                     | 参数类型 | 必选 | 描述                                                                       |
| ---------------------------- | -------- | ---- | -------------------------------------------------------------------------- |
| bk_host_id                   | int64    | 是   | 已存在的主机 ID，必须大于等于 `0`                                          |
| target_version               | string   | 否   | 目标 Agent 版本；为空时自动选择默认版本                                    |
| force                        | bool     | 否   | 是否强制重启，默认 `false`，传递给 GSE 重启操作                            |
| graceful_restart_timeout_sec | int64    | 否   | 优雅重启超时时间，单位秒；默认 `0`，传递给 GSE 重启操作                    |
| bk_networkunit_id            | int64    | 否   | 升级后的管控单元 ID；默认 `-1`，保留当前管控单元                           |
| cpu_arch                     | string   | 否   | 请求可携带该字段，但当前升级执行不会应用此值；按主机已记录的平台选择软件包 |

主机必须存在，且升级使用的管控单元 ID 必须大于等于 `0`。建议先调用 `OpenNodeAgent_OpenNodeAgentUpgradeCheck` 检查目标单元与平台版本映射；检查通过不代表异步升级一定成功。

### 调用示例

```json
{
  "host": [
    {
      "bk_host_id": 1001,
      "target_version": "2.1.0",
      "force": false,
      "graceful_restart_timeout_sec": 60
    }
  ]
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
| code       | int32    | 状态码，`0` 表示请求成功，不表示升级完成 |
| message    | string   | 请求信息                                 |
| request_id | string   | 请求 ID                                  |
| error      | object   | 错误信息                                 |
| permission | object   | 权限信息                                 |
| data       | object   | 响应数据                                 |

#### data

| 参数名称    | 参数类型 | 描述                                           |
| ----------- | -------- | ---------------------------------------------- |
| workflow_id | string   | 工作流 ID，用于跟踪异步升级；不是 `trigger_id` |

v3.0.1-alpha.114 起，可使用 `OpenNodeWorkflow_OpenNodeWorkflowList`，按 `exact_include_conditions.workflow_id` 查询执行状态；查询另需对应的历史查看权限。
