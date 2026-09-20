### 描述

- 该接口提供版本：v3.0.1-alpha.85+。
- 该接口所需权限：无。
- 该接口功能描述：触发一次性工作流，将 system 租户的共享制品同步到当前租户。

### URL

POST /api/v3/sync/release/shared

### 输入参数

无业务参数，请提交 JSON 空对象 `{}`。租户 ID 和操作者用户名从已认证的请求上下文获取，两者均不能为空，无需在请求体中传入。

### 使用说明

- 无额外 IAM 操作权限要求；仍需通过接口身份认证。
- 当前租户不能为 system；system 租户调用会在创建工作流前被拒绝。
- 同步范围包括 system 租户标记为共享的 Cert、BinTool、PluginBinTool、Agent、Proxy 和 Plugin 发布包。
- 接口成功仅表示工作流已创建并激活，不代表共享制品同步已完成。

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
    "trigger_id": "trig:6c0379fe82b44920a8c49508fb744d72"
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0 表示工作流已成功创建并激活 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时为空 |
| permission | object | 权限信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| trigger_id | string | 本次一次性工作流的触发器 ID（trigger ID），不表示执行结果 |
