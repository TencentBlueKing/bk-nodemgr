### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：networkunit_view（查看管控单元）。
- 该接口功能描述：获取拓扑图节点信息，返回指定管控单元的 Agent 和 Proxy 统计数据及健康状态。

### URL

POST /api/v3/topo/graph_node/get

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkunit_id | array(int64) | 否 | 管控单元 ID 列表。为空时返回所有有权限的管控单元信息 |

### 调用示例

```json
{
  "bk_networkunit_id": [1, 2, 3]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "abc123",
  "data": {
    "graph_node_info": [
      {
        "bk_networkunit_id": 1,
        "running_proxy": 2,
        "total_proxy": 2,
        "running_agent": 150,
        "total_agent": 200,
        "is_healthy": true,
        "cycle_times": ["2024-01-01T10:00:00Z", "2024-01-01T10:05:00Z"]
      },
      {
        "bk_networkunit_id": 2,
        "running_proxy": 0,
        "total_proxy": 1,
        "running_agent": 80,
        "total_agent": 100,
        "is_healthy": false,
        "cycle_times": []
      }
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
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| graph_node_info | array | 图节点信息列表 |

#### graph_node_info[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| bk_networkunit_id | int64 | 管控单元 ID |
| running_proxy | int64 | 运行中的 Proxy 数量 |
| total_proxy | int64 | Proxy 总数 |
| running_agent | int64 | 运行中的 Agent 数量 |
| total_agent | int64 | Agent 总数 |
| is_healthy | bool | 节点健康状态。当 running_proxy 为 0 或存在未满足的必需 Proxy 标签时为 false |
| cycle_times | array(string) | Agent 心跳周期时间列表，ISO 8601 格式 |
