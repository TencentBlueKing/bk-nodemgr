### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：无。
- 该接口功能描述：按需获取拓扑相关常量值，目前支持云厂商列表和操作系统类型列表。

### URL

POST /api/v3/topo/constant/get

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| cloud_vendor | bool | 否 | 是否返回云厂商列表。为 `true` 时返回 `data.cloud_vendor`，为 `false` 或不传时返回空数组 |
| os_type | bool | 否 | 是否返回操作系统类型列表。为 `true` 时返回 `data.os_type`，为 `false` 或不传时返回空数组 |

### 调用示例

```json
{
  "cloud_vendor": true,
  "os_type": true
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "cloud_vendor": [
      "tencent",
      "aws"
    ],
    "os_type": [
      "LINUX",
      "WINDOWS"
    ]
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
| permission | object | 权限信息，当前接口通常为空 |
| data | object | 常量数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| cloud_vendor | string array | 云厂商列表 |
| os_type | string array | 操作系统类型列表 |

### 说明

- 该接口的请求契约与响应结构定义在 `proto/backend/api/v3/topo.proto` 和 `docs/api/swagger/backend/api/v3/topo.swagger.json`。
- 运行时行为定义在 `internal/backend/router/api-v3/topo/constant.go`：接口仅返回请求中显式开启的常量项，未开启的项返回空数组。
