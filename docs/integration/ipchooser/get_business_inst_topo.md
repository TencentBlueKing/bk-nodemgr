# get_business_inst_topo

## 目的与适用场景

当调用方需要按 `bk_biz_id` 拉取该业务下的完整实例拓扑（`biz -> set -> module` + 自定义拓扑分组），并要求每个拓扑节点带预聚合的主机数时，使用 `get_business_inst_topo`。

## 输入

请求字段以 [topo.proto `TopoBusinessInstTopoGetReq`](../../../proto/backend/api/v3/topo.proto) 为准：

| 字段        | Required | 含义                 |
| ----------- | -------- | -------------------- |
| `bk_biz_id` | yes      | 目标业务 ID（int64） |

## 最小 payload 和 curl template

需要 `curl` 和 `jq`。先设置当前部署的 API base URL 与目标业务 ID。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_BIZ_ID=2

TOPO_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/topo/business/inst_topo/get" \
  -H "Content-Type: application/json" \
  -d "{\"bk_biz_id\": ${BK_BIZ_ID}}")"

ROOT_HOST_COUNT="$(printf '%s' "${TOPO_RESPONSE}" | jq -r '.data.items.host_count')"
FIRST_LEVEL_SETS="$(printf '%s' "${TOPO_RESPONSE}" | jq -r '.data.items.children[] | "\(.topo_inst_name) (\(.host_count))"')"
```

## 系统解释

平台把该请求解释为按 `bk_biz_id` 拉取该业务下的完整实例拓扑：

- 返回的 `data.items` 是单个 `TopoNodeInfo` 根节点；不是数组。
- 每个 `TopoNodeInfo` 含 5 个字段（来自 [topo.proto `TopoNodeInfo`](../../../proto/backend/api/v3/topo.proto)）：

| 字段             | 类型                  | 含义                                                                                          |
| ---------------- | --------------------- | --------------------------------------------------------------------------------------------- |
| `topo_obj_id`    | string                | 拓扑对象类型标识；用于在 `list_host` 中映射到对应的 `exact_include_conditions` 字段（见下表） |
| `topo_inst_id`   | int64                 | 该类型下的实例 ID                                                                             |
| `topo_inst_name` | string                | 实例名                                                                                        |
| `host_count`     | int64                 | 该节点下的主机总数（含子节点）                                                                |
| `children`       | array\<TopoNodeInfo\> | 子节点；递归展开形成完整拓扑                                                                  |

- `topo_obj_id` 与 `list_host` 字段的映射关系：

| `topo_obj_id` 取值    | 含义       | `list_host` 字段                     |
| --------------------- | ---------- | ------------------------------------ |
| `biz`                 | 业务       | `bk_biz_id`                          |
| `set`                 | 集群       | `bk_set_id`                          |
| `module`              | 模块       | `bk_module_id`                       |
| 其它自定义拓扑对象 ID | 自定义分组 | 公开 contract 中无直接映射（待确认） |

- 平台在响应前会按业务 / 集群 / 模块 / 自定义拓扑分组预聚合每个节点的 `host_count`；调用方拿到响应后不需要再发额外请求计算。
- 业务无访问权限时，平台返回权限错误。
- 业务对应 cmdb 拓扑根节点结构异常（多于 1 个根）时，平台返回内部错误；具体错误码以平台为准（待确认）。
- 业务对应 cmdb 拓扑为空时，平台返回 `data.items` 为空（`host_count` / `children` 可能不存在或为空），是合法响应。

## 即时输出

`data` 字段以 [topo.swagger.json `v3TopoBusinessInstTopoGetResp`](../../api/swagger/backend/api/v3/topo.swagger.json) 为准：

| 字段         | 类型           | 含义                                            |
| ------------ | -------------- | ----------------------------------------------- |
| `data.items` | `TopoNodeInfo` | 单个根节点；`host_count` 与 `children` 递归展开 |

`TopoNodeInfo` 字段同 [系统解释](#系统解释) 节中的表。

## 最终或机器侧可见产物

`get_business_inst_topo` 是只读查询；不修改后端数据。

| 产物             | 来源                                                                                                                                                                           |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 实例拓扑         | `data.items` 的递归 `TopoNodeInfo` 树；每个节点自带 `host_count`                                                                                                               |
| 节点 -> 主机过滤 | `topo_obj_id` / `topo_inst_id`；调用方按上表映射到 `list_host` 字段                                                                                                            |
| 业务下聚合主机数 | `data.items.host_count`（根节点）；与 `get_business_host_count` 的 host_count 含义不同：前者按 cmdb 拓扑节点聚合（含子节点递归、含空闲池与自定义拓扑分组规则），后者按业务聚合 |

## 重复行为

读取接口是无副作用的查询；重复调用不改变后端数据。

公开 contract 不定义：

- 推荐轮询间隔
- 同一 `bk_biz_id` 的多次调用在并发场景下的取值合并行为
- `TopoNodeInfo` 树的最大深度（理论上受 cmdb 自定义拓扑层级限制）

## 失败情况与限制

- 请求校验失败（`bk_biz_id` 字段缺失或非法）时返回 API 错误。
- 当前账号对业务无访问权限时，平台返回权限错误；具体错误码以平台为准（待确认）。
- 业务对应 cmdb 拓扑根节点多于 1 个时，平台返回内部错误（来自 cmdb 的根节点结构异常）；具体错误码以平台为准（待确认）。
- 业务对应 cmdb 拓扑为空时返回 `data.items` 为空对象；是合法响应。
- 自定义拓扑对象 ID（`topo_obj_id` 非 `biz` / `set` / `module`）的 `topo_inst_id` 与 `list_host` 的字段映射在本文档 contract 中未定义（待确认）。
- `host_count` 是预聚合结果；调用方传入的任何过滤 / 搜索条件不会反映在该数字上。

## Contract 参考

- [接入总览](README.md)
- [拉取业务列表](list_business.md)
- [批量取业务主机数](get_business_host_count.md)
- [分页查询主机](list_host.md)
- [Swagger contract](../../api/swagger/backend/api/v3/topo.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/topo.proto)
