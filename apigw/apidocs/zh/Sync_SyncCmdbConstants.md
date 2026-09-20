### 描述

- 该接口提供版本：v3.0.1-alpha.60+。
- 该接口所需权限：无。
- 该接口功能描述：触发同步 CMDB 常量数据。

### URL

POST /api/v3/sync/cmdb/constants

### 输入参数

无

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
| code | int32 | 状态码，0 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时为空 |
| permission | object | 权限信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| trigger_id | string | 同步任务的触发器 ID，不表示执行结果 |
