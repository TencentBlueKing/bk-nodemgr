### 描述

- 该接口提供版本：v3.0.1-alpha.60+。
- 该接口所需权限：无。
- 该接口功能描述：触发同步指定主机的存活插件进程信息。

### URL

POST /api/v3/sync/gse/plugin/process/info

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| host_ids | int64 array | 否 | 主机 ID 列表 |

### 调用示例

```json
{
  "host_ids": [
    1
  ]
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
| trigger_id | string | 触发 ID |
