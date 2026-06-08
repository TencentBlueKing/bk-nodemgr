### 描述

- 该接口提供版本：v3.0.1-alpha.38+。
- 该接口所需权限：无。
- 该接口功能描述：根据业务 ID 查询业务实例拓扑，并返回各拓扑节点聚合后的主机数量。

### URL

POST /api/v3/topo/business/inst_topo/get

### 输入参数

| 参数名称  | 参数类型 | 必选 | 描述    |
| --------- | -------- | ---- | ------- |
| bk_biz_id | int64    | 是   | 业务 ID |

### 调用示例

查询业务 `2` 的业务实例拓扑。

```json
{
  "bk_biz_id": 2
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "items": {
      "topo_inst_id": 2,
      "topo_inst_name": "prod-payment",
      "topo_obj_id": "biz",
      "topo_obj_name": "业务",
      "host_count": 155,
      "children": [
        {
          "topo_inst_id": 10,
          "topo_inst_name": "default set",
          "topo_obj_id": "set",
          "topo_obj_name": "集群",
          "host_count": 120,
          "children": [
            {
              "topo_inst_id": 100,
              "topo_inst_name": "module-a",
              "topo_obj_id": "module",
              "topo_obj_name": "模块",
              "host_count": 120,
              "children": []
            }
          ]
        },
        {
          "topo_inst_id": 11,
          "topo_inst_name": "custom layer",
          "topo_obj_id": "custom_level",
          "topo_obj_name": "自定义层级",
          "host_count": 35,
          "children": [
            {
              "topo_inst_id": 101,
              "topo_inst_name": "module-b",
              "topo_obj_id": "module",
              "topo_obj_name": "模块",
              "host_count": 35,
              "children": []
            }
          ]
        }
      ]
    }
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

| 参数名称 | 参数类型 | 描述                                               |
| -------- | -------- | -------------------------------------------------- |
| items    | object   | 业务实例拓扑根节点。没有查询到拓扑根节点时可能为空 |

#### data.items / data.items.children[n]

| 参数名称       | 参数类型 | 描述                                                                      |
| -------------- | -------- | ------------------------------------------------------------------------- |
| topo_inst_id   | int64    | 拓扑实例 ID                                                               |
| topo_inst_name | string   | 拓扑实例名称                                                              |
| topo_obj_id    | string   | 拓扑对象 ID，例如 `biz`、`set`、`module`，也可能是 CMDB 自定义层级对象 ID |
| topo_obj_name  | string   | 拓扑对象名称                                                              |
| host_count     | int64    | 当前拓扑节点聚合后的主机数量                                              |
| children       | array    | 子拓扑节点列表，结构与当前节点一致                                        |

### 说明

- 该接口的请求契约与响应结构定义在 `proto/backend/api/v3/topo.proto`、`pkg/proto/backend/api/v3/business.go`、`internal/backend/router/api-v3/topo/business.go` 与 `docs/api/swagger/backend/api/v3/topo.swagger.json`。
- 请求参数 `bk_biz_id` 为 `0` 时，协议层校验会返回参数错误。
- 后端从 CMDB 查询业务实例拓扑，并对 `biz`、`set`、`module` 节点按实际主机关系统计主机数量；其他自定义层级节点的 `host_count` 由子节点数量向上汇总得到。
- `data.items` 是单棵业务拓扑树的根节点；每个 `children` 元素都使用同一节点结构。
