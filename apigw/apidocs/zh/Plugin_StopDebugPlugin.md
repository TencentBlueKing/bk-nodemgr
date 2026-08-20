### 描述

- 该接口提供版本：v3.0.1-alpha.72+。
- 该接口所需权限：⚠️ 待确认（接口尚未实现 handler，权限待补充）。
- 该接口功能描述：停止一个正在运行的插件调试工作流。

### URL

POST /api/v3/plugin/stop_debug

### 输入参数

| 参数名称    | 参数类型 | 必选 | 描述                                     |
| ----------- | -------- | ---- | ---------------------------------------- |
| workflow_id | string   | 是   | 待停止的调试工作流ID，由启动调试接口返回 |

### 调用示例

```json
{
  "workflow_id": "workflow-abc123def456"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {}
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                   |
| ---------- | -------- | ---------------------- |
| code       | int32    | 状态码，0表示成功      |
| message    | string   | 请求信息               |
| request_id | string   | 请求ID                 |
| error      | object   | 错误信息，成功时为null |
| data       | object   | 响应数据，为空对象     |

#### error

| 参数名称 | 参数类型 | 描述         |
| -------- | -------- | ------------ |
| system   | string   | 错误系统标识 |
| message  | string   | 错误消息     |
| details  | array    | 错误详情列表 |

#### error.details[n]

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| code     | string   | 错误代码 |
| message  | string   | 错误消息 |
