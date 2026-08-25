# list_host

## 目的与适用场景

当调用方需要按拓扑节点 / 业务 / 网络区域 / 节点类型 / IP / 主机名等条件分页查询主机列表时，使用 `list_host`。

`list_host` 支持两种使用模式：分页（`page`）拉取满足过滤条件的主机数组，仅取总数（`only_count`）跳过主机行只读 `data.total`。也可同时使用 `exact_include_conditions` 与 `fuzzy_include_conditions`，同一字段的精确与模糊不可同时传入。

## 输入

请求字段以 [topo.proto `TopoHostListReq`](../../../proto/backend/api/v3/topo.proto) 为准：

| 字段                       | Required | 含义                                                     |
| -------------------------- | -------- | -------------------------------------------------------- |
| `page`                     | no       | `Page.offset` / `Page.limit`；不传时按平台默认（待确认） |
| `only_count`               | no       | 仅返回 `total`，不返回 `items`；用于分页前先取总数       |
| `exact_include_conditions` | no       | 精确包含；见下表                                         |
| `fuzzy_include_conditions` | no       | 模糊包含；见下表                                         |

`TopoHostExactConditions` 字段（精确包含）：

| 字段                 | 类型     | 含义                                                                                    |
| -------------------- | -------- | --------------------------------------------------------------------------------------- |
| `bk_host_id`         | []int64  | 主机 ID                                                                                 |
| `bk_biz_id`          | []int64  | 业务 ID                                                                                 |
| `bk_networkarea_id`  | []int64  | 网络区域 ID                                                                             |
| `bk_networkunit_id`  | []int64  | 网络单元 ID                                                                             |
| `bk_set_id`          | []int64  | 集群 ID；由 `get_business_inst_topo` 的 `topo_obj_id=set` 节点的 `topo_inst_id` 传入    |
| `bk_module_id`       | []int64  | 模块 ID；由 `get_business_inst_topo` 的 `topo_obj_id=module` 节点的 `topo_inst_id` 传入 |
| `os_type`            | []string | 操作系统类型（`linux` / `windows` / `aix` / `darwin` 等）                               |
| `arch`               | []string | CPU 架构（`x86_64` / `aarch64` 等）                                                     |
| `node_role`          | []string | 节点角色（`proxy` / `agent` 等）                                                        |
| `node_status`        | []string | 节点状态（`running` / `unknown` 等）                                                    |
| `node_version`       | []string | 节点版本                                                                                |
| `bk_agent_id`        | []string | agent ID                                                                                |
| `node_generation`    | []int64  | 节点 generation                                                                         |
| `proxy_tags`         | []string | proxy 标签                                                                              |
| `bk_host_innerip`    | []string | 主机内网 IPv4（精确匹配）；与 `fuzzy_include_conditions.bk_host_innerip` 不可同时使用   |
| `bk_host_innerip_v6` | []string | 主机内网 IPv6（精确匹配）                                                               |

`TopoHostFuzzyConditions` 字段（模糊包含）：

| 字段                 | 类型     | 含义                   |
| -------------------- | -------- | ---------------------- |
| `bk_host_name`       | []string | 主机名模糊匹配         |
| `dept_name`          | []string | 所属部门模糊匹配       |
| `bk_host_innerip`    | []string | 主机内网 IPv4 模糊匹配 |
| `bk_host_innerip_v6` | []string | 主机内网 IPv6 模糊匹配 |
| `bk_host_outerip`    | []string | 主机外网 IPv4 模糊匹配 |
| `bk_host_outerip_v6` | []string | 主机外网 IPv6 模糊匹配 |

`exact_include_conditions` 与 `fuzzy_include_conditions` 可同时使用；同一字段（如 `bk_host_innerip`）的精确与模糊不可同时传入。

## 最小 payload 和 curl template

需要 `curl` 和 `jq`。先设置当前部署的 API base URL 与目标过滤条件。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_BIZ_ID=2
export BK_SET_ID=100
export BK_MODULE_ID=200

HOST_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/topo/host/list" \
  -H "Content-Type: application/json" \
  -d "{
    \"page\": {\"offset\": 0, \"limit\": 50},
    \"exact_include_conditions\": {
      \"bk_biz_id\": [${BK_BIZ_ID}],
      \"bk_set_id\": [${BK_SET_ID}],
      \"bk_module_id\": [${BK_MODULE_ID}]
    }
  }")"

TOTAL="$(printf '%s' "${HOST_RESPONSE}" | jq -r '.data.total')"
FIRST_HOST_INNERIP="$(printf '%s' "${HOST_RESPONSE}" | jq -r '.data.items[0].info.bk_host_innerip_list[0]')"
FIRST_HOST_ID="$(printf '%s' "${HOST_RESPONSE}" | jq -r '.data.items[0].bk_host_id')"
```

按业务 + 模糊搜索主机名 / IP：

```bash
HOST_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/topo/host/list" \
  -H "Content-Type: application/json" \
  -d "{
    \"page\": {\"offset\": 0, \"limit\": 50},
    \"exact_include_conditions\": {
      \"bk_biz_id\": [${BK_BIZ_ID}]
    },
    \"fuzzy_include_conditions\": {
      \"bk_host_name\": [\"web\"],
      \"bk_host_innerip\": [\"10.0.\"]
    }
  }")"
```

仅取总数：

```bash
COUNT_ONLY_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/topo/host/list" \
  -H "Content-Type: application/json" \
  -d "{
    \"only_count\": true,
    \"exact_include_conditions\": {
      \"bk_biz_id\": [${BK_BIZ_ID}]
    }
  }")"

HOST_TOTAL="$(printf '%s' "${COUNT_ONLY_RESPONSE}" | jq -r '.data.total')"
```

## 系统解释

平台把该请求解释为按 `exact_include_conditions` 与 `fuzzy_include_conditions` 过滤后分页查询主机：

- `exact_include_conditions` 中同一字段的多个值是"或"关系（in 匹配）；不同字段之间是"与"关系。
- `fuzzy_include_conditions` 中同一字段的多个值是"或"关系；不同字段之间是"与"关系。
- 精确与模糊之间是"与"关系。
- `page.offset` / `page.limit` 决定分页窗口；`total` 表示满足条件的主机总数。
- 返回的 `items` 中每项 `Host` 含 `tenant_id` / `bk_host_id` / `info`（`HostInfo`）/ `state`（`HostState`）/ `create_at` / `updated_at`（来自 [common.proto `Host`](../../../proto/backend/api/v3/common.proto)）。

## 即时输出

`data` 字段以 [topo.swagger.json `v3TopoHostListResp`](../../api/swagger/backend/api/v3/topo.swagger.json) 为准：

| 字段           | 类型          | 含义               |
| -------------- | ------------- | ------------------ |
| `data.total`   | int64         | 满足条件的主机总数 |
| `data.items[]` | array\<Host\> | 主机数组           |

`Host` 字段（[common.proto `Host`](../../../proto/backend/api/v3/common.proto)）：

| 字段         | 类型      | 含义                                    |
| ------------ | --------- | --------------------------------------- |
| `tenant_id`  | string    | 租户 ID                                 |
| `bk_host_id` | int64     | 主机 ID                                 |
| `info`       | HostInfo  | 主机静态信息（IP / 名称 / OS / 登录等） |
| `state`      | HostState | 主机 agent 状态（角色 / 状态 / 版本等） |
| `create_at`  | uint64    | 创建时间（Unix 秒）                     |
| `updated_at` | uint64    | 更新时间（Unix 秒）                     |

`HostInfo` 字段（节选自 [common.proto](../../../proto/backend/api/v3/common.proto)）：

| 字段                      | 类型     | 含义               |
| ------------------------- | -------- | ------------------ |
| `bk_biz_id`               | int64    | 业务 ID            |
| `bk_networkarea_id`       | int64    | 网络区域 ID        |
| `bk_networkunit_id`       | int64    | 网络单元 ID        |
| `bk_host_name`            | string   | 主机名             |
| `dept_name`               | string   | 所属部门           |
| `bk_host_innerip_list`    | []string | 主机内网 IPv4 列表 |
| `bk_host_innerip_v6_list` | []string | 主机内网 IPv6 列表 |
| `bk_host_outerip_list`    | []string | 主机外网 IPv4 列表 |
| `bk_host_outerip_v6_list` | []string | 主机外网 IPv6 列表 |
| `os_type`                 | string   | 操作系统类型       |
| `cpu_arch`                | string   | CPU 架构           |
| `login_ip`                | string   | 登录 IP            |
| `login_port`              | int64    | 登录端口           |
| `login_user`              | string   | 登录用户           |
| `login_mode`              | string   | 登录模式           |
| `login_credit_valid`      | bool     | 登录凭据是否有效   |
| `export_ip`               | string   | 出口 IPv4          |
| `export_ip_v6`            | string   | 出口 IPv6          |
| `advertise_ip`            | string   | 对外宣告 IPv4      |
| `advertise_ip_v6`         | string   | 对外宣告 IPv6      |
| `relay_callback_port`     | int64    | relay 回调端口     |
| `relay_download_port`     | int64    | relay 下载端口     |
| `bk_addressing`           | string   | 地址类型           |

`HostState` 字段（节选自 [common.proto](../../../proto/backend/api/v3/common.proto)）：

| 字段              | 类型     | 含义            |
| ----------------- | -------- | --------------- |
| `node_role`       | string   | 节点角色        |
| `node_status`     | string   | 节点状态        |
| `node_version`    | string   | 节点版本        |
| `bk_agent_id`     | string   | agent ID        |
| `node_generation` | int64    | 节点 generation |
| `proxy_tags`      | []string | proxy 标签      |

## 最终或机器侧可见产物

`list_host` 是只读查询；不修改后端数据。

| 产物     | 来源                                                                                                 |
| -------- | ---------------------------------------------------------------------------------------------------- |
| 主机数组 | `data.items[]`；每项 `bk_host_id` / `info.bk_host_name` / `info.bk_host_innerip_list` / `state.*` 等 |
| 总数     | `data.total`                                                                                         |

`list_host` 的 `data.total` 与 `get_business_host_count` 的 `host_count` 含义不同：前者按 `exact_include_conditions` + `fuzzy_include_conditions` 过滤后聚合，后者按"该业务下所有主机"聚合。

## 重复行为

读取接口是无副作用的查询；重复调用不改变后端数据。

公开 contract 不定义：

- 推荐轮询间隔
- 同一组过滤条件的多次调用在并发场景下的取值合并行为
- `page.limit` 的上限（待确认）
- 主机的 `info` / `state` 字段的更新频率

## 失败情况与限制

- 请求校验失败（`page.limit` 非法等）时返回 API 错误。
- 当前账号对某个业务无访问权限时，平台返回权限错误；具体错误码以平台为准（待确认）。
- 主机的 `bk_host_innerip_list` 可能为空（主机没有内网 IP）；`bk_host_id` 一定非空。
- 同一字段（如 `bk_host_innerip`）的精确与模糊不可同时传入；同时传入时具体行为以平台返回为准（待确认）。
- `exact_include_conditions` 中传 `bk_host_id` 表示按主机 ID 精确查询；与 `bk_biz_id` / `bk_set_id` 等其它字段可同时使用，是"与"关系。
- `data.items` 为空数组但 `data.total > 0` 表示分页越界（`page.offset` 超出总数）；是合法响应，调用方应调整 `page.offset` 重新查询。
- `only_count = true` 时只返回 `data.total`，`data.items` 行为以平台为准（待确认）。
- `info.login_credit_valid` 的取值在本文档 contract 中未定义（待确认）。

## Contract 参考

- [接入总览](README.md)
- [拉取业务列表](list_business.md)
- [批量取业务主机数](get_business_host_count.md)
- [取业务实例拓扑](get_business_inst_topo.md)
- [Swagger contract](../../api/swagger/backend/api/v3/topo.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/topo.proto)
- [共享类型定义](../../../proto/backend/api/v3/common.proto)
