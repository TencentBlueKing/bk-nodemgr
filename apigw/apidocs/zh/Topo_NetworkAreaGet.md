### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：networkarea_view（查看管控区域详情）。
- 该接口功能描述：根据管控区域 ID 查询单个管控区域详情。

### URL

POST /api/v3/topo/networkarea/get

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkarea_id | int64 | 是 | 管控区域 ID |

### 调用示例

```json
{
  "bk_networkarea_id": 1
}
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
    "bk_networkarea_name": "prod-default",
    "cloud_vendor": "tencent"
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| data | object | 管控区域详情 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| tenant_id | string | 租户 ID |
| bk_networkarea_id | int64 | 管控区域 ID |
| bk_networkarea_name | string | 管控区域名称 |
| cloud_vendor | string | 云区域标识 |

### 说明

- 该接口的契约以 `proto/application/api/v3/topo.proto`、`pkg/proto/application/api/v3/networkarea.go` 和 `docs/api/swagger/application/api/v3/topo.swagger.json` 为准。
- 当前 application handler `internal/application/router/api-v3/topo/networkarea.go` 存在实现偏差：代码路径目前返回的是 create-style 数据结构，仅包含 `bk_networkarea_id`。本文档按 proto / swagger 契约描述标准返回结构；若依赖运行时行为，请先确认服务实现是否已修正。
