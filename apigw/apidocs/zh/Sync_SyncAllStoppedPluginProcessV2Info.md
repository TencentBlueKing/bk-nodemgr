### 描述

- 该接口提供版本：v3.0.1-alpha.111+。
- 该接口所需权限：无。
- 该接口功能描述：触发同步全部主机的 V2 兼容插件已停止进程信息。

### URL

POST /api/v3/sync/gse/plugin/process_v2/info/stopped/all

### 输入参数

无业务参数，请提交 JSON 空对象 `{}`。租户 ID 和操作者用户名从已认证的请求上下文获取。

### 使用说明

- 由定时工作流 `scheduled_sync_stopped_plugin_process_v2_info` 使用同一个生成动作：先扫描当前租户下 Agent 不为空的主机，按批（每批最多 1000 台）生成同步任务，并发度由全局配置 `oper_sync_stopped_plugin_process_info_max_concurrency_num`（默认 100）控制。
- 每个同步任务只处理一批主机中 `process_v2` 记录状态为 `stopped` 的进程：Agent 在线时向 GSE 的 `nodeman` 命名空间查询真实信息，以 GSE 确认的结果回写，查询失败或 GSE 未上报的进程保留本地 `stopped` 记录；Agent 离线时标记为 `unknown`。
- 该接口只覆盖节点管理 V2 兼容插件进程；V3 插件进程请使用 `Sync_SyncAllStoppedPluginProcessInfo`。
- 接口成功仅表示工作流已创建并激活，不代表同步已完成；同步结果需查看工作流执行记录。

### 调用示例

```json
{}
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
