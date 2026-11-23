### 描述

- 该接口提供版本：v3.0.1-alpha.107+。
- 该接口所需权限：system_access（系统访问）；按 `node_role` 还需 agent_view（查看 Agent）或 proxy_view（查看 Proxy）。`blank`、`agent` 使用 Agent 权限，`proxy` 使用 Proxy 权限；未指定角色或同时查询两类角色时，取所需权限的授权业务范围交集。
- 该接口功能描述：通过用户态入口查询主机列表，支持分页和属性过滤，结果按授权业务范围收窄，并按最近更新时间倒序返回。
- 用户身份及租户信息由 API 网关传递，无需在请求体中提供用户名或租户 ID。

### URL

POST /api/v3/open/topo/host/list

### 输入参数

| 参数名称                 | 参数类型 | 必选     | 描述                                               |
| ------------------------ | -------- | -------- | -------------------------------------------------- |
| page                     | object   | 条件必选 | `only_count` 不为 `true` 时必须提供有效的 `limit`  |
| only_count               | bool     | 否       | 默认 `false`；为 `true` 时只统计总数，`items` 为空 |
| exact_include_conditions | object   | 否       | 精确匹配条件                                       |
| fuzzy_include_conditions | object   | 否       | 模糊匹配条件                                       |

#### page

| 参数名称 | 参数类型 | 必选     | 描述                                                         |
| -------- | -------- | -------- | ------------------------------------------------------------ |
| offset   | int32    | 否       | 起始记录偏移量，默认 `0`，必须大于等于 `0`                   |
| limit    | int32    | 条件必选 | 列表查询时范围为 `(0, 1000]`；仅计数时可省略，传入值不能为负 |

#### exact_include_conditions

各字段使用数组提供候选值；`node_role` 同时决定所需查看权限。

| 参数名称           | 参数类型     | 必选 | 描述                                                                                                   |
| ------------------ | ------------ | ---- | ------------------------------------------------------------------------------------------------------ |
| bk_host_id         | int64 array  | 否   | 主机 ID 列表                                                                                           |
| bk_biz_id          | int64 array  | 否   | 业务 ID 列表，与授权业务范围取交集                                                                     |
| bk_networkarea_id  | int64 array  | 否   | 管控区域 ID 列表                                                                                       |
| bk_set_id          | int64 array  | 否   | CMDB 集群 ID 列表                                                                                      |
| bk_module_id       | int64 array  | 否   | CMDB 模块 ID 列表                                                                                      |
| bk_host_innerip    | string array | 否   | 内网 IPv4 精确匹配列表                                                                                 |
| bk_host_innerip_v6 | string array | 否   | 内网 IPv6 精确匹配列表                                                                                 |
| os_type            | string array | 否   | 操作系统类型列表                                                                                       |
| node_role          | string array | 否   | 节点角色：`blank`、`agent`、`proxy`                                                                    |
| node_status        | string array | 否   | 节点状态：`init`、`running`、`damaged`、`busy`、`starting`、`upgrade`、`stopping`、`uninit`、`unknown` |
| node_version       | string array | 否   | 节点版本列表                                                                                           |
| bk_agent_id        | string array | 否   | Agent ID 列表                                                                                          |
| bk_networkunit_id  | int64 array  | 否   | 管控单元 ID 列表                                                                                       |
| node_generation    | int64 array  | 否   | 节点代际列表                                                                                           |
| arch               | string array | 否   | CPU 架构列表                                                                                           |
| proxy_tags         | string array | 否   | Proxy 标签：`dedicated_installer`、`cluster_tunnel`、`file_tunnel`、`data_tunnel`                      |

#### fuzzy_include_conditions

| 参数名称           | 参数类型     | 必选 | 描述                     |
| ------------------ | ------------ | ---- | ------------------------ |
| bk_host_name       | string array | 否   | 主机名模糊匹配列表       |
| dept_name          | string array | 否   | 业务模块名称模糊匹配列表 |
| bk_host_innerip    | string array | 否   | 内网 IPv4 模糊匹配列表   |
| bk_host_innerip_v6 | string array | 否   | 内网 IPv6 模糊匹配列表   |
| bk_host_outerip    | string array | 否   | 外网 IPv4 模糊匹配列表   |
| bk_host_outerip_v6 | string array | 否   | 外网 IPv6 模糊匹配列表   |

### 调用示例

```json
{
  "page": { "offset": 0, "limit": 10 },
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "node_role": ["agent"],
    "node_status": ["running"]
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
    "total": 1,
    "items": [
      {
        "tenant_id": "default",
        "bk_host_id": 1001,
        "info": {
          "bk_biz_id": 2,
          "bk_networkarea_id": 1,
          "bk_networkunit_id": 10,
          "bk_host_name": "agent-prod-01",
          "dept_name": "payment",
          "bk_host_innerip_list": ["10.0.0.12"],
          "bk_host_innerip_v6_list": [],
          "bk_host_outerip_list": [],
          "bk_host_outerip_v6_list": [],
          "bk_mac": "00:16:3e:12:34:56",
          "os_type": "linux",
          "cpu_arch": "x86_64",
          "login_ip": "10.0.0.12",
          "login_port": 22,
          "login_user": "root",
          "login_mode": "password",
          "login_credit_valid": true,
          "export_ip": "",
          "export_ip_v6": "",
          "advertise_ip": "10.0.0.12",
          "advertise_ip_v6": "",
          "relay_callback_port": 0,
          "relay_download_port": 0,
          "bk_addressing": "static"
        },
        "state": {
          "node_role": "agent",
          "node_status": "running",
          "node_version": "2.4.1",
          "bk_agent_id": "agent-1001",
          "node_generation": 2,
          "proxy_tags": [],
          "last_sync_at": 1795392000000
        }
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

| 参数名称 | 参数类型     | 描述                         |
| -------- | ------------ | ---------------------------- |
| total    | int64        | 授权范围内满足条件的主机总数 |
| items    | object array | 当前页主机列表；仅计数时为空 |

#### data.items[n]

| 参数名称   | 参数类型        | 描述                                                       |
| ---------- | --------------- | ---------------------------------------------------------- |
| tenant_id  | string          | 租户 ID                                                    |
| bk_host_id | int64           | 主机 ID                                                    |
| info       | object          | 主机静态信息及登录信息，不返回登录密码或密钥内容           |
| state      | object          | 节点状态                                                   |
| create_at  | string (uint64) | 创建时间字段；当前列表响应转换未赋值，不应据此判断创建时间 |
| updated_at | string (uint64) | 更新时间字段；当前列表响应转换未赋值，不应据此判断更新时间 |

#### data.items[n].info

| 参数名称                | 参数类型     | 描述                            |
| ----------------------- | ------------ | ------------------------------- |
| bk_biz_id               | int64        | 业务 ID                         |
| bk_networkarea_id       | int64        | 管控区域 ID                     |
| bk_networkunit_id       | int64        | 管控单元 ID                     |
| bk_host_name            | string       | 主机名                          |
| dept_name               | string       | 业务模块名称                    |
| bk_host_innerip_list    | string array | 内网 IPv4 列表                  |
| bk_host_innerip_v6_list | string array | 内网 IPv6 列表                  |
| bk_host_outerip_list    | string array | 外网 IPv4 列表                  |
| bk_host_outerip_v6_list | string array | 外网 IPv6 列表                  |
| bk_mac                  | string       | MAC 地址                        |
| os_type                 | string       | 操作系统类型                    |
| cpu_arch                | string       | CPU 架构                        |
| login_ip                | string       | 登录 IP                         |
| login_port              | int64        | 登录端口                        |
| login_user              | string       | 登录用户                        |
| login_mode              | string       | 登录方式                        |
| login_credit_valid      | bool         | 登录凭据是否有效                |
| export_ip               | string       | NAT 出口 IPv4                   |
| export_ip_v6            | string       | NAT 出口 IPv6                   |
| advertise_ip            | string       | NAT 公告 IPv4                   |
| advertise_ip_v6         | string       | NAT 公告 IPv6                   |
| relay_callback_port     | int64        | Relay 回调端口                  |
| relay_download_port     | int64        | Relay 下载端口                  |
| bk_addressing           | string       | 寻址方式：`static` 或 `dynamic` |

#### data.items[n].state

| 参数名称        | 参数类型     | 描述                                                            |
| --------------- | ------------ | --------------------------------------------------------------- |
| node_role       | string       | 节点角色，与查询条件中的角色值相同                              |
| node_status     | string       | 节点状态，与查询条件中的状态值相同                              |
| node_version    | string       | 节点版本                                                        |
| bk_agent_id     | string       | Agent ID                                                        |
| node_generation | int64        | 节点代际                                                        |
| proxy_tags      | string array | Proxy 标签，与查询条件中的标签值相同                            |
| last_sync_at    | int64        | 最近一次从 GSE 同步状态的 Unix 毫秒时间戳，`0` 表示尚无同步时间 |

### 说明

授权业务范围与查询条件没有交集时返回空列表；权限校验失败时返回权限错误。分页总数及仅计数结果均受相同授权范围限制。
