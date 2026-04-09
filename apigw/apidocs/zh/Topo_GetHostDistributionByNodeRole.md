### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：networkarea_view（查看管控区域）。
- 该接口功能描述：根据主机筛选条件统计各节点角色对应的主机数量分布。

### URL

POST /api/v3/topo/host/get_host_distribution_by_node_role

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| exact_include_conditions | object | 否 | 精确匹配包含条件 |
| fuzzy_include_conditions | object | 否 | 模糊匹配包含条件 |

#### exact_include_conditions

精确匹配包含条件。每个字段都是数组形式，服务端按主机条件进行筛选。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_host_id | int64 array | 否 | 主机 ID 列表 |
| bk_biz_id | int64 array | 否 | 业务 ID 列表 |
| bk_networkarea_id | int64 array | 否 | 管控区域 ID 列表 |
| os_type | string array | 否 | 操作系统类型列表 |
| node_role | string array | 否 | 节点角色列表 |
| node_status | string array | 否 | 节点状态列表 |
| node_version | string array | 否 | 节点版本列表 |
| bk_agent_id | string array | 否 | Agent ID 列表 |
| bk_networkunit_id | int64 array | 否 | 网络单元 ID 列表 |
| node_generation | int64 array | 否 | 节点代次列表 |
| arch | string array | 否 | CPU 架构列表 |

#### fuzzy_include_conditions

模糊匹配包含条件。每个字段都是数组形式，服务端按主机字符串字段进行模糊筛选。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_host_name | string array | 否 | 主机名列表，按主机名模糊匹配 |
| dept_name | string array | 否 | 业务模块名称列表，按模块名称模糊匹配 |
| bk_host_innerip | string array | 否 | 主机内网 IPv4 列表，按内网 IPv4 模糊匹配 |
| bk_host_innerip_v6 | string array | 否 | 主机内网 IPv6 列表，按内网 IPv6 模糊匹配 |
| bk_host_outerip | string array | 否 | 主机外网 IPv4 列表，按外网 IPv4 模糊匹配 |
| bk_host_outerip_v6 | string array | 否 | 主机外网 IPv6 列表，按外网 IPv6 模糊匹配 |

### 调用示例

统计业务 `2` 下各节点角色的主机数量分布。

```json
{
  "exact_include_conditions": {
    "bk_biz_id": [
      2
    ]
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "agent": 120,
    "proxy": 8,
    "blank": 15
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时通常为空 |
| permission | object | 权限信息，当前接口通常为空 |
| data | object | 节点角色到主机数量的映射 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| {node_role} | int64 | 某个节点角色对应的主机数量，键为节点角色字符串，当前领域模型中的角色值包括 `blank`、`agent`、`proxy` |

### 说明

- 该接口的请求契约与响应结构定义在 `proto/backend/api/v3/topo.proto`、`pkg/proto/backend/api/v3/host.go` 与 `docs/api/swagger/backend/api/v3/topo.swagger.json`。
- backend router `internal/backend/router/api-v3/topo/host.go` 对请求体执行 `BindJSON` 后，直接将筛选条件转换为 `types.HostCondition` 并调用存储层统计，返回值为节点角色字符串到数量的映射。
