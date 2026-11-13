### 描述

- 该接口提供版本：v3.0.1-alpha.111+。
- 该接口所需权限：无。
- 该接口功能描述：触发同步指定主机的 V2 兼容插件已停止进程信息。

### URL

POST /api/v3/sync/gse/plugin/process_v2/info/stopped

### 输入参数

| 参数名称 | 参数类型    | 必选 | 描述         |
| -------- | ----------- | ---- | ------------ |
| host_ids | int64 array | 否   | 主机 ID 列表 |

### 使用说明

- 该接口只处理 `process_v2` 记录，即节点管理 V2 兼容插件的进程；V3 插件进程请使用 `Sync_SyncStoppedPluginProcessInfo`。
- 同步范围是本地记录中状态为 `stopped` 的进程，与 V2 存活进程同步（定时工作流 `scheduled_sync_alive_plugin_process_v2_info`）互补。
- Agent 在线的主机：向 GSE 的 `nodeman` 命名空间查询这些进程的真实信息，以 GSE 确认的结果回写本地记录；查询失败或 GSE 未上报的进程保留本地 `stopped` 记录，不会被标记为 `unknown`。
- Agent 离线的主机：无法向 GSE 求证，进程被标记为 `unknown`，不会保留本地 `stopped` 记录。
- 接口成功仅表示工作流已创建并激活，不代表同步已完成；同步结果需查看工作流执行记录。

### 调用示例

```json
{
  "host_ids": [1]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "error": null,
  "permission": null,
  "data": {
    "trigger_id": "trigger-001"
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                             |
| ---------- | -------- | -------------------------------- |
| code       | int32    | 状态码，0 表示工作流已创建并激活 |
| message    | string   | 请求信息                         |
| request_id | string   | 请求 ID                          |
| error      | object   | 错误信息，成功时为空             |
| permission | object   | 权限信息                         |
| data       | object   | 响应数据                         |

#### data

| 参数名称   | 参数类型 | 描述                                  |
| ---------- | -------- | ------------------------------------- |
| trigger_id | string   | 触发 ID（trigger ID），不表示执行结果 |
