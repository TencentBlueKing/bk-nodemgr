### 描述

- 该接口提供版本：v3.0.1-alpha.107+。
- 该接口所需权限：system_access（系统访问）、networkunit_view（查看管控单元）。
- 该接口功能描述：通过用户态入口查询管控单元图节点、Agent/Proxy 数量和健康信息，结果按可查看的管控单元范围收窄。
- 用户身份及租户信息由 API 网关传递，无需在请求体中提供用户名或租户 ID。

### URL

POST /api/v3/open/topo/graph_node/get

### 输入参数

| 参数名称          | 参数类型    | 必选 | 描述                                                         |
| ----------------- | ----------- | ---- | ------------------------------------------------------------ |
| bk_networkunit_id | int64 array | 否   | 管控单元 ID 列表；省略或为空时查询用户可查看范围内的管控单元 |

### 调用示例

```json
{ "bk_networkunit_id": [10] }
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "graph_node_info": [
      {
        "bk_networkunit_id": 10,
        "running_proxy": 1,
        "total_proxy": 1,
        "running_agent": 10,
        "total_agent": 12,
        "is_healthy": true,
        "cycle_times": [
          {
            "bk_host_id": 1001,
            "bk_host_innerip_list": ["10.0.0.10"],
            "bk_host_innerip_v6_list": [],
            "bk_agent_id": "agent-1001",
            "time": "2ms"
          }
        ]
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                             |
| ---------- | -------- | -------------------------------- |
| code       | int32    | 状态码，`0` 表示成功             |
| message    | string   | 请求结果信息                     |
| request_id | string   | 请求 ID                          |
| error      | object   | 错误信息，成功时为空             |
| permission | object   | 权限申请信息，权限不足时可能返回 |
| data       | object   | 响应数据                         |

#### data

| 参数名称        | 参数类型     | 描述                                     |
| --------------- | ------------ | ---------------------------------------- |
| graph_node_info | object array | 管控单元图节点信息，顺序不保证与请求一致 |

#### data.graph_node_info[n]

| 参数名称          | 参数类型     | 描述                                                                                                                                                             |
| ----------------- | ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| bk_networkunit_id | int64        | 管控单元 ID                                                                                                                                                      |
| running_proxy     | int64        | 运行中的 Proxy 数量                                                                                                                                              |
| total_proxy       | int64        | Proxy 总数                                                                                                                                                       |
| running_agent     | int64        | 运行中的 Agent 数量                                                                                                                                              |
| total_agent       | int64        | Agent 总数                                                                                                                                                       |
| is_healthy        | bool         | 当前健康判定：存在运行中的 Proxy，且这些 Proxy 的标签合计覆盖 `dedicated_installer`、`cluster_tunnel`、`file_tunnel`、`data_tunnel`；不是所有 Agent 均正常的保证 |
| cycle_times       | object array | Proxy 连接往返时间信息                                                                                                                                           |

#### data.graph_node_info[n].cycle_times[n]

| 参数名称                | 参数类型     | 描述                                                                 |
| ----------------------- | ------------ | -------------------------------------------------------------------- |
| bk_host_id              | int64        | Proxy 主机 ID                                                        |
| bk_host_innerip_list    | string array | 内网 IPv4 列表                                                       |
| bk_host_innerip_v6_list | string array | 内网 IPv6 列表                                                       |
| bk_agent_id             | string       | Proxy 的 Agent ID                                                    |
| time                    | string       | GSE 上报的连接往返时间字符串，保留原始格式，不作为固定单位的数值返回 |

### 说明

授权范围收窄后没有匹配管控单元时返回空列表；权限校验失败时返回权限错误。
