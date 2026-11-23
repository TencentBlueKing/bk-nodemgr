### 描述

- 该接口提供版本：v3.0.1-alpha.107+。
- 该接口所需权限：system_access（系统访问）、networkarea_view（查看管控区域）。
- 该接口功能描述：通过用户态入口按 ID 查询管控区域详情，校验目标管控区域的查看权限。
- 用户身份及租户信息由 API 网关传递，无需在请求体中提供用户名或租户 ID。

### URL

POST /api/v3/open/topo/networkarea/get

### 输入参数

| 参数名称          | 参数类型 | 必选 | 描述                                               |
| ----------------- | -------- | ---- | -------------------------------------------------- |
| bk_networkarea_id | int64    | 否   | 管控区域 ID，必须大于或等于 `0`；省略时按 `0` 查询 |

### 调用示例

```json
{ "bk_networkarea_id": 1 }
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "tenant_id": "default",
    "bk_networkarea_id": 1,
    "bk_networkarea_name": "示例区域",
    "cloud_vendor": "1"
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                                                                        |
| ---------- | -------- | --------------------------------------------------------------------------- |
| code       | int32    | 状态码，`0` 表示成功                                                        |
| message    | string   | 请求结果信息                                                                |
| request_id | string   | 请求 ID                                                                     |
| error      | object   | 错误信息，包含 `system`、`message` 和 `details`（错误码及说明）；成功时为空 |
| permission | object   | 权限申请信息，权限不足时可能返回                                            |
| data       | object   | 管控区域详情                                                                |

#### data

| 参数名称            | 参数类型 | 描述            |
| ------------------- | -------- | --------------- |
| tenant_id           | string   | 租户 ID         |
| bk_networkarea_id   | int64    | 管控区域 ID     |
| bk_networkarea_name | string   | 管控区域名称    |
| cloud_vendor        | string   | CMDB 云厂商标识 |
