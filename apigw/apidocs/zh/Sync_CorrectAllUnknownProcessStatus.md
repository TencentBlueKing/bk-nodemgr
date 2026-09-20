### 描述

- 该接口提供版本：v3.0.1-alpha.90+。
- 该接口所需权限：无。
- 该接口功能描述：异步触发当前租户全部主机的插件进程状态过期检查，将过期状态修正为 `unknown`。

### 行为说明

- 以纠正任务执行时的时间为基准，进程最后同步时间早于 48 小时前时，将状态设为 `unknown`；恰好等于截止时间的记录不处理。
- 最后同步时间缺失、为 `null` 或为零时间的记录也会被处理。
- 仅修改进程状态，保留最后同步时间、自动托管标志、PID、版本和 Agent ID 等其他进程信息。
- 该接口依据已有同步时间纠正记录，不调用 GSE 查询实际进程状态，也不执行启动、停止或取消托管操作。
- 返回成功表示异步任务已创建并激活，不代表全部进程状态已完成更新。
- 全量检查按当前租户的主机清单分批执行，包含没有 Agent ID 的主机。

### URL

POST /api/v3/sync/gse/plugin/process/correct_unknown_status/all

### 输入参数

无。请求体传入 `{}`。

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

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0 表示成功触发任务 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时为空 |
| permission | object | 权限信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| trigger_id | string | 本次异步纠正任务的触发 ID |
