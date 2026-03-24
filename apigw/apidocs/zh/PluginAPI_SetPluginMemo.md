### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：。
- 该接口功能描述：设置指定插件的备注信息。

### URL

POST /api/v3/plugin/set_memo

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| plugin_name | string | 是 | 插件名称 |
| memo | string | 否 | 插件备注内容；可为空字符串 |

**参数说明**：
- `plugin_name` 不能为空，否则请求校验失败。
- `memo` 当前未设置长度或格式限制，接口会直接更新插件备注字段。

### 调用示例

```json
{
  "plugin_name": "bk-monitor-agent",
  "memo": "用于主机监控采集"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260324-000004",
  "error": null
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求ID |
| error | object | 错误信息，成功时为null |

#### error

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| system | string | 错误系统标识 |
| message | string | 错误消息 |
| details | array | 错误详情列表 |

#### error.details[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | string | 错误代码 |
| message | string | 错误消息 |
