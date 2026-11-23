### 描述

- 该接口提供版本：v3.0.1-alpha.107+。
- 该接口所需权限：system_access（系统访问）；根据 `node_role` 还需 agent_view（查看 Agent）或 proxy_view（查看 Proxy）。`blank`、`agent` 使用 Agent 权限，`proxy` 使用 Proxy 权限；未指定角色或同时查询两类角色时，取所需权限的授权业务范围交集。
- 该接口功能描述：通过用户态入口查询指定业务的实例拓扑，并返回各节点符合角色条件的主机数量。
- 用户身份及租户信息由 API 网关传递，无需在请求体中提供用户名或租户 ID。

### URL

POST /api/v3/open/topo/business/inst_topo/get

### 输入参数

| 参数名称  | 参数类型     | 必选 | 描述                                                                          |
| --------- | ------------ | ---- | ----------------------------------------------------------------------------- |
| bk_biz_id | int64        | 是   | 业务 ID，不能为 `0`；需要对应业务的查看权限                                   |
| node_role | string array | 否   | 节点角色：`blank`（未安装节点）、`agent`、`proxy`；不传或为空表示不按角色过滤 |

### 调用示例

```json
{ "bk_biz_id": 2, "node_role": ["agent"] }
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "items": {
      "topo_obj_id": "biz",
      "topo_inst_id": 2,
      "topo_inst_name": "示例业务",
      "host_count": 10,
      "children": [
        {
          "topo_obj_id": "set",
          "topo_inst_id": 20,
          "topo_inst_name": "示例集群",
          "host_count": 10,
          "children": []
        }
      ]
    }
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
| data       | object   | 响应数据                                                                    |

#### data

| 参数名称 | 参数类型 | 描述                                                    |
| -------- | -------- | ------------------------------------------------------- |
| items    | object   | 单个业务拓扑根节点，**不是数组**；CMDB 未返回拓扑时为空 |

#### data.items / children[n]

子节点递归使用相同结构。

| 参数名称       | 参数类型     | 描述                                                                 |
| -------------- | ------------ | -------------------------------------------------------------------- |
| topo_obj_id    | string       | 拓扑对象类型，如 `biz`、`set`、`module`，也可能为 CMDB 自定义层级    |
| topo_inst_id   | int64        | 拓扑实例 ID                                                          |
| topo_inst_name | string       | 拓扑实例名称                                                         |
| host_count     | int64        | 节点范围内符合角色条件的主机数；自定义层级按其后代集群的主机去重统计 |
| children       | object array | 子节点列表                                                           |

### 说明

- 未获得该业务所需查看权限时，返回权限错误，不以空拓扑替代权限校验。
- 拓扑来自 CMDB，主机数量来自节点管理的主机记录，可能存在同步时差。
