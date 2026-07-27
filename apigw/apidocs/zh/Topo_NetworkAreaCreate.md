### 描述

- 该接口提供版本：v3.0.1-alpha.60+。
- 该接口所需权限：networkarea_create（创建管控区域）。
- 该接口功能描述：创建管控区域，并同步写入本地拓扑数据。

### URL

POST /api/v3/topo/networkarea/create

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkarea_name | string | 是 | 管控区域名称 |
| cloud_vendor | string | 是 | 云区域标识 |

### 调用示例

```json
{
  "bk_networkarea_name": "default",
  "cloud_vendor": "string"
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
    "bk_networkarea_id": 1
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
| bk_networkarea_id | int64 | 管控区域 ID |
