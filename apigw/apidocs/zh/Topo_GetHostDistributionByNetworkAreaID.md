### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：networkarea_view（查看管控区域）。
- 该接口功能描述：根据主机筛选条件，统计各管控区域下的主机数量分布。

### URL

POST /api/v3/topo/host/get_host_distribution_by_networkarea_id

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| exact_include_conditions | object | 否 | 精确匹配条件 |
| fuzzy_include_conditions | object | 否 | 模糊匹配条件 |

#### exact_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_host_id | array<int64> | 否 | 主机 ID 列表 |
| bk_biz_id | array<int64> | 否 | 业务 ID 列表 |
| bk_networkarea_id | array<int64> | 否 | 管控区域 ID 列表 |
| os_type | array<string> | 否 | 主机操作系统类型列表 |
| node_role | array<string> | 否 | 节点角色列表 |
| node_status | array<string> | 否 | 节点状态列表 |
| node_version | array<string> | 否 | 节点版本列表 |
| bk_agent_id | array<string> | 否 | Agent ID 列表 |
| bk_networkunit_id | array<int64> | 否 | 接入点分组 ID 列表 |
| node_generation | array<int64> | 否 | 节点代际列表 |
| arch | array<string> | 否 | CPU 架构列表 |
| proxy_tags | array<string> | 否 | Proxy 标签列表，可选值：`dedicated_installer`、`cluster_tunnel`、`file_tunnel`、`data_tunnel` |

#### fuzzy_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_host_name | array<string> | 否 | 主机名列表，按模糊条件匹配 |
| dept_name | array<string> | 否 | 所属部门名称列表，按模糊条件匹配 |
| bk_host_innerip | array<string> | 否 | 主机内网 IPv4 地址列表，按模糊条件匹配 |
| bk_host_innerip_v6 | array<string> | 否 | 主机内网 IPv6 地址列表，按模糊条件匹配 |
| bk_host_outerip | array<string> | 否 | 主机外网 IPv4 地址列表，按模糊条件匹配 |
| bk_host_outerip_v6 | array<string> | 否 | 主机外网 IPv6 地址列表，按模糊条件匹配 |

### 调用示例

```json
{
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "node_status": ["RUNNING"],
    "bk_networkarea_id": [1, 2]
  },
  "fuzzy_include_conditions": {
    "bk_host_innerip": ["10.0."]
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
    "1": 12,
    "2": 5
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| data | object | 主机数量分布结果，键为管控区域 ID，值为命中的主机数量 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| {bk_networkarea_id} | int64 | 以管控区域 ID 作为键的动态字段 |
| {bk_networkarea_id} 对应的值 | int64 | 该管控区域下命中的主机数量 |

### 说明

- 该接口契约以 `proto/backend/api/v3/topo.proto`、`pkg/proto/backend/api/v3/host.go`、`internal/backend/router/api-v3/topo/host.go` 和 `docs/api/swagger/backend/api/v3/topo.swagger.json` 为准。
- `data` 在 proto 中定义为 `map<int64, int64>`。实际 JSON 返回时，对象键会序列化为字符串形式的管控区域 ID。
