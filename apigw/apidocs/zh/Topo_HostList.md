### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：agent_view（查看 Agent）、proxy_view（查看 Proxy）。
- 该接口功能描述：查询主机列表，支持分页、按主机/节点属性过滤，并按最近更新时间倒序返回结果。

### URL

POST /api/v3/topo/host/list

### 输入参数

| 参数名称                 | 参数类型 | 必选 | 描述                                                                   |
| ------------------------ | -------- | ---- | ---------------------------------------------------------------------- |
| page                     | object   | 否   | 分页配置；`offset` 必须大于等于 `0`，`limit` 必须在 `(0, 1000]` 范围内 |
| only_count               | bool     | 否   | 是否只返回总数；为 `true` 时仅统计数量，不返回主机详情列表             |
| exact_include_conditions | object   | 否   | 精确匹配包含条件                                                       |
| fuzzy_include_conditions | object   | 否   | 模糊匹配包含条件                                                       |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                                 |
| -------- | -------- | ---- | ------------------------------------ |
| offset   | int32    | 否   | 记录起始偏移量，起始值为 `0`         |
| limit    | int32    | 否   | 返回记录上限，取值范围为 `(0, 1000]` |

#### exact_include_conditions

精确匹配包含条件。每个字段都是数组形式，服务端按主机条件筛选；其中 `node_role` 会影响权限收敛逻辑。

| 参数名称          | 参数类型     | 必选 | 描述                                                                                                                  |
| ----------------- | ------------ | ---- | --------------------------------------------------------------------------------------------------------------------- |
| bk_host_id        | int64 array  | 否   | 主机 ID 列表                                                                                                          |
| bk_biz_id         | int64 array  | 否   | 业务 ID 列表                                                                                                          |
| bk_networkarea_id | int64 array  | 否   | 管控区域 ID 列表                                                                                                      |
| bk_set_id         | int64 array  | 否   | 集群 ID 列表                                                                                                          |
| bk_module_id      | int64 array  | 否   | 模块 ID 列表                                                                                                          |
| os_type           | string array | 否   | 操作系统类型列表                                                                                                      |
| node_role         | string array | 否   | 节点角色列表，可选值通常包括 `blank`、`agent`、`proxy`                                                                |
| node_status       | string array | 否   | 节点状态列表，可选值包括 `init`、`running`、`damaged`、`busy`、`starting`、`upgrade`、`stopping`、`uninit`、`unknown` |
| node_version      | string array | 否   | 节点版本列表                                                                                                          |
| bk_agent_id       | string array | 否   | Agent ID 列表                                                                                                         |
| bk_networkunit_id | int64 array  | 否   | 网络单元 ID 列表                                                                                                      |
| node_generation   | int64 array  | 否   | 节点代次列表                                                                                                          |
| arch              | string array | 否   | CPU 架构列表                                                                                                          |
| proxy_tags        | string array | 否   | Proxy 标签列表，可选值：`dedicated_installer`、`cluster_tunnel`、`file_tunnel`、`data_tunnel`                         |

#### fuzzy_include_conditions

模糊匹配包含条件。每个字段都是数组形式，服务端按主机字符串字段进行模糊筛选。

| 参数名称           | 参数类型     | 必选 | 描述                                     |
| ------------------ | ------------ | ---- | ---------------------------------------- |
| bk_host_name       | string array | 否   | 主机名列表，按主机名模糊匹配             |
| dept_name          | string array | 否   | 业务模块名称列表，按模块名称模糊匹配     |
| bk_host_innerip    | string array | 否   | 主机内网 IPv4 列表，按内网 IPv4 模糊匹配 |
| bk_host_innerip_v6 | string array | 否   | 主机内网 IPv6 列表，按内网 IPv6 模糊匹配 |
| bk_host_outerip    | string array | 否   | 主机外网 IPv4 列表，按外网 IPv4 模糊匹配 |
| bk_host_outerip_v6 | string array | 否   | 主机外网 IPv6 列表，按外网 IPv6 模糊匹配 |

### 调用示例

查询业务 `2` 下运行中的 Agent 主机列表，并返回前 10 条记录。

```json
{
  "page": {
    "offset": 0,
    "limit": 10
  },
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "node_role": ["agent"],
    "node_status": ["running"]
  },
  "fuzzy_include_conditions": {
    "bk_host_innerip": ["10.0.0."]
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
    "total": 2,
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
          "bk_host_outerip_list": ["203.0.113.12"],
          "bk_host_outerip_v6_list": [],
          "bk_mac": "00:16:3e:12:34:56",
          "os_type": "linux",
          "cpu_arch": "x86_64",
          "login_ip": "10.0.0.12",
          "login_port": 22,
          "login_user": "root",
          "login_mode": "password",
          "login_credit_valid": true,
          "export_ip": "203.0.113.12",
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
          "node_generation": 1,
          "proxy_tags": []
        }
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
| error      | object   | 错误信息，成功时通常为空   |
| permission | object   | 权限信息，当前接口通常为空 |
| data       | object   | 响应数据                   |

#### data

| 参数名称 | 参数类型 | 描述                                                |
| -------- | -------- | --------------------------------------------------- |
| total    | int64    | 当前规则能匹配到的总记录条数                        |
| items    | array    | 查询返回的主机列表；当 `only_count=true` 时通常为空 |

#### data.items[n]

| 参数名称   | 参数类型 | 描述                       |
| ---------- | -------- | -------------------------- |
| tenant_id  | string   | 租户 ID                    |
| bk_host_id | int64    | 主机 ID                    |
| info       | object   | 主机静态信息与登录相关信息 |
| state      | object   | 节点状态信息               |

#### data.items[n].info

| 参数名称                | 参数类型     | 描述                                     |
| ----------------------- | ------------ | ---------------------------------------- |
| bk_biz_id               | int64        | 业务 ID                                  |
| bk_networkarea_id       | int64        | 管控区域 ID                              |
| bk_networkunit_id       | int64        | 网络单元 ID                              |
| bk_host_name            | string       | 主机名                                   |
| dept_name               | string       | 业务模块名称                             |
| bk_host_innerip_list    | string array | 主机内网 IPv4 列表                       |
| bk_host_innerip_v6_list | string array | 主机内网 IPv6 列表                       |
| bk_host_outerip_list    | string array | 主机外网 IPv4 列表                       |
| bk_host_outerip_v6_list | string array | 主机外网 IPv6 列表                       |
| bk_mac                  | string       | 主机 MAC 地址                            |
| os_type                 | string       | 操作系统类型                             |
| cpu_arch                | string       | CPU 架构                                 |
| login_ip                | string       | 登录 IP                                  |
| login_port              | int64        | 登录端口                                 |
| login_user              | string       | 登录用户名                               |
| login_mode              | string       | 登录方式                                 |
| login_credit_valid      | bool         | 登录凭据是否仍然有效                     |
| export_ip               | string       | NAT 出口 IPv4                            |
| export_ip_v6            | string       | NAT 出口 IPv6                            |
| advertise_ip            | string       | NAT 公告 IPv4                            |
| advertise_ip_v6         | string       | NAT 公告 IPv6                            |
| relay_callback_port     | int64        | Relay 回调端口                           |
| relay_download_port     | int64        | Relay 下载端口                           |
| bk_addressing           | string       | 寻址方式，常见值为 `static` 或 `dynamic` |

#### data.items[n].state

| 参数名称        | 参数类型     | 描述           |
| --------------- | ------------ | -------------- |
| node_role       | string       | 节点角色       |
| node_status     | string       | 节点状态       |
| node_version    | string       | 节点版本       |
| bk_agent_id     | string       | Agent ID       |
| node_generation | int64        | 节点代次       |
| proxy_tags      | string array | Proxy 标签列表 |

### 说明

- 该接口的请求契约与响应结构定义在 `proto/backend/api/v3/topo.proto`、`proto/backend/api/v3/common.proto`、`pkg/proto/backend/api/v3/host.go` 与 `docs/api/swagger/backend/api/v3/topo.swagger.json`。
- backend router `internal/backend/router/api-v3/topo/host.go` 会先将请求转换为 `types.HostCondition`，再按 `node_role` 选择 `agent_view` 或 `proxy_view` 权限并对 `bk_biz_id` 做授权范围收敛。
- 当未显式传入 `node_role` 时，服务端按 Agent 与 Proxy 两类可见范围共同收敛业务列表；当显式传入 `node_role=["proxy"]` 时，仅按 Proxy 可见范围筛选。
