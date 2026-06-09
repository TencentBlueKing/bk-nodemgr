### 描述

- 该接口提供版本：v3.0.1-alpha.38+。
- 该接口所需权限：`agent_view（查看 Agent）`、`proxy_view（查看 Proxy）`。
- 该接口功能描述：根据业务 ID 列表，统计每个业务下的主机数量。

### URL

POST /api/v3/topo/business/host_count/get

### 输入参数

| 参数名称  | 参数类型    | 必选 | 描述                              |
| --------- | ----------- | ---- | --------------------------------- |
| bk_biz_id | int64 array | 是   | 业务 ID 列表，至少传入一个业务 ID |

### 调用示例

查询业务 `2` 和业务 `7` 下的主机数量。

```json
{
  "bk_biz_id": [2, 7]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "items": [
      {
        "bk_biz_id": 2,
        "host_count": 120
      },
      {
        "bk_biz_id": 7,
        "host_count": 35
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                       |
| ---------- | -------- | -------------------------- |
| code       | int32    | 状态码，0 表示成功         |
| message    | string   | 请求信息                   |
| request_id | string   | 请求 ID                    |
| error      | object   | 错误信息，成功时为空       |
| permission | object   | 权限信息，当前接口通常为空 |
| data       | object   | 响应数据                   |

#### data

| 参数名称 | 参数类型 | 描述                       |
| -------- | -------- | -------------------------- |
| items    | array    | 按业务聚合后的主机数量列表 |

#### data.items[n]

| 参数名称   | 参数类型 | 描述               |
| ---------- | -------- | ------------------ |
| bk_biz_id  | int64    | 业务 ID            |
| host_count | int64    | 该业务下的主机数量 |

### 说明

- 该接口的请求契约与响应结构定义在 `proto/backend/api/v3/topo.proto` 与 `docs/api/swagger/backend/api/v3/topo.swagger.json`。
- 请求参数 `bk_biz_id` 为空时，协议层校验会返回参数错误。
- 响应中的 `items` 由后端按业务 ID 聚合主机数量生成；未命中的业务通常不会出现在返回列表中。
