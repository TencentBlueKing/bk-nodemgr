# IP / host chooser 接入指南

本文档描述 bk-nodemgr 提供的 4 个公开 topo endpoint 如何被组合调用，以便第三方调用者一次性获得"按业务 / 拓扑节点 / 主机过滤维度"查询主机的能力。每个 endpoint 的请求、响应字段和字段语义以 [topo.swagger.json](../../api/swagger/backend/api/v3/topo.swagger.json) 与 [topo.proto](../../../proto/backend/api/v3/topo.proto) 为准；共享的 `Business` / `Host` / `HostInfo` / `HostState` / `Page` 类型以 [common.proto](../../../proto/backend/api/v3/common.proto) 为准。

## 接入结果

`ip_selector` 是把 4 个公开 topo endpoint 组合起来，让调用方按"业务 → 实例拓扑 → 主机"层层定位到具体主机的能力。4 个 endpoint 之间的关系：

1. `list_business` 返回业务列表，作为后续 3 个 endpoint 的入口维度（按 `bk_biz_id` 缩窄）。
2. `get_business_host_count` 给定若干 `bk_biz_id`，返回每业务的主机总数。
3. `get_business_inst_topo` 给定 `bk_biz_id`，返回该业务下的实例拓扑（`biz -> set -> module` + 自定义拓扑），每个节点预聚合 `host_count`。
4. `list_host` 给定一组过滤条件（业务 / 集群 / 模块 / 节点类型 / IP / 主机名等），分页返回满足条件的主机。

四者组合可以一次性拿到"业务 → 拓扑节点 → 该节点下主机"的查询链路：调用方先选业务，再选拓扑节点，最后用节点对应的过滤字段拉主机。

4 个 endpoint 的响应只表示"按当前请求条件返回的数据"；不证明调用方已经把数据用于任何下游动作。

## API-first 快速接入

curl template 按 endpoint 拆分：

| 步骤             | 阅读                                                  |
| ---------------- | ----------------------------------------------------- |
| 拉取业务列表     | [list_business](list_business.md)                     |
| 批量取业务主机数 | [get_business_host_count](get_business_host_count.md) |
| 取业务实例拓扑   | [get_business_inst_topo](get_business_inst_topo.md)   |
| 分页查询主机     | [list_host](list_host.md)                             |

完整调用链：

1. 调用 `list_business` 拿到业务列表，得到 `bk_biz_id` 集合。
2. 把这些 `bk_biz_id` 传给 `get_business_host_count` 拿到每业务的主机数。
3. 调用 `get_business_inst_topo(bk_biz_id=...)` 拿到该业务下的实例拓扑与每个节点的预聚合 `host_count`。
4. 把拓扑节点的 `topo_obj_id` / `topo_inst_id` 映射到 `list_host` 的 `exact_include_conditions` 字段（业务 / 集群 / 模块），分页拉取该节点下的主机。

## 接入流程

### 1. 确认 API 上下文

公开的 4 个 topo endpoints（均在 `bk-nodemgr` backend service 下）：

| 步骤             | Method and path                             | 即时输出                                                          |
| ---------------- | ------------------------------------------- | ----------------------------------------------------------------- |
| 拉取业务列表     | `POST /api/v3/topo/business/list`           | `data.total`, `data.items[]`（业务数组）                          |
| 批量取业务主机数 | `POST /api/v3/topo/business/host_count/get` | `data.items[]`（每业务的 host count）                             |
| 取业务实例拓扑   | `POST /api/v3/topo/business/inst_topo/get`  | `data.items`（单个 `TopoNodeInfo` 根，含 children 与 host count） |
| 分页查询主机     | `POST /api/v3/topo/host/list`               | `data.total`, `data.items[]`（主机数组）                          |

curl template 中的 `BK_NODEMGR_API_BASE` 由调用方提供，表示当前部署的 API base URL；topo contract 不定义统一 gateway 或认证 header。

### 2. 准备 shell 变量

下面的章节都会复用以下变量。`BK_BIZ_ID` 表示当前要查询的业务；4 个 endpoint 本身都支持多业务查询，所以可以是数组。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_BIZ_ID=2
```

### 3. 调用顺序

调用方按下面的顺序组合调用这 4 个 endpoint（并发 / 串行由调用方自己决定；本文档描述的是数据依赖而不是时序约束）：

1. **并行（可选）**：`list_business` 与 `get_business_host_count` 没有相互数据依赖，可并行调用把业务列表与每业务主机数一次性拿到。
2. **按需（每业务）**：`get_business_inst_topo(bk_biz_id=...)` 需要来自 `list_business` 的 `bk_biz_id`，所以必须在拿到业务 ID 后调用。同一个 `bk_biz_id` 一次响应即可拿到完整拓扑；调用方在客户端缓存后不需重复调用。
3. **按需（每节点）**：`list_host` 需要来自 `get_business_inst_topo` 的 `topo_obj_id` / `topo_inst_id` 才能映射到过滤字段。同一个拓扑节点下，按 `page` 与 `fuzzy_include_conditions` 分页调用。
4. **按需（搜索 / 过滤变化）**：`list_host` 可被同一调用方在过滤条件变化时重复调用；服务端不维护调用方状态。

### 4. 解读即时输出

| endpoint                  | 关键响应字段                                                                                                   | 用途                                                                                        |
| ------------------------- | -------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| `list_business`           | `data.items[].bk_biz_id` / `bk_biz_name` / `tenant_id`                                                         | 作为后续 3 个 endpoint 的入参维度                                                           |
| `get_business_host_count` | `data.items[].bk_biz_id` / `host_count`                                                                        | 与 `list_business` 的 `items` 按 `bk_biz_id` 对齐                                           |
| `get_business_inst_topo`  | `data.items`（单个 `TopoNodeInfo`）+ 递归 `children`                                                           | 每个节点自带 `topo_obj_id` / `topo_inst_id` / `host_count`，可直接作为 `list_host` 过滤入参 |
| `list_host`               | `data.total` + `data.items[].bk_host_id` / `info.bk_host_innerip_list` / `info.os_type` / `state.node_role` 等 | 直接给出当前过滤条件下分页的主机列表                                                        |

### 5. 观察最终产物

4 个 endpoint 都是只读查询；不修改后端数据。

| 产物           | 来源                                                          |
| -------------- | ------------------------------------------------------------- |
| 业务列表       | `list_business.data.items[]`                                  |
| 业务主机数     | `get_business_host_count.data.items[]`（按 `bk_biz_id` 对齐） |
| 实例拓扑       | `get_business_inst_topo.data.items`（递归 `TopoNodeInfo`）    |
| 某节点下的主机 | `list_host.data.items[]`（按节点对应字段过滤）                |
| 总主机数       | `list_host.data.total`                                        |

组合调用不直接证明调用方已把数据用于任何下游动作；4 个 endpoint 都只是"读"。

### 6. 重复行为

4 个 endpoint 都是无副作用的查询；重复调用不改变后端数据。

公开 contract 不定义：

- 推荐轮询间隔
- 同一组过滤条件的多次调用在并发场景下的取值合并行为
- 同一 `bk_biz_id` 在多次 `get_business_inst_topo` 调用之间的结果合并行为
- `page.limit` 的上限（待确认）

调用方应自行管理：

- 同一 `bk_biz_id` 下 `get_business_inst_topo` 结果的客户端缓存（避免重复调用）。
- `list_host` 的去抖（同一组过滤条件在短时间内的多次请求）。
- 调用方拿到的 `bk_host_id` 列表的去重 / 同步。

### 7. 处理失败与边界

以下内容属于接入边界：

- 请求结构错误或 unsupported fields 是 API contract 问题，检查 [topo.swagger.json](../../api/swagger/backend/api/v3/topo.swagger.json)。
- `get_business_inst_topo` 在 `bk_biz_id` 对应的 cmdb 拓扑根节点多于 1 个时返回平台内部错误；具体返回以平台为准（待确认）。
- `get_business_inst_topo` 在 cmdb 拓扑为空时返回 `data.items` 为空的合法响应，不算错误。
- `list_business` 的 `fuzzy_include_conditions.bk_biz_name` 是模糊匹配；具体匹配规则（大小写、前后缀、通配符）以平台返回为准（待确认）。
- `list_host` 的 `bk_set_id` / `bk_module_id` 等拓扑字段在组合调用中来自 `get_business_inst_topo` 返回的 `TopoNodeInfo`；映射规则见 `get_business_inst_topo` 的"系统解释"小节。
- 权限不足时 `get_business_host_count` / `get_business_inst_topo` / `list_host` 会拒绝访问不属于当前调用者的业务；具体返回的 `code` / `message` 以平台返回为准（待确认）。
- 4 个 endpoint 都没有"idempotency key" / "retry token" 概念；调用方负责自己的请求去重与重试策略。

### 8. 重复或修订选择器结果

4 个 endpoint 都是无副作用的查询组合；调用方不需要管理"撤销 / 重做"。如果调用方在已经按某种条件拉过数据后希望换成另一种条件，只需重新发起对应 endpoint 的调用，服务端会按当前请求条件重新返回。

## Contract 参考

- [list_business](list_business.md)
- [get_business_host_count](get_business_host_count.md)
- [get_business_inst_topo](get_business_inst_topo.md)
- [list_host](list_host.md)
- [Swagger contract](../../api/swagger/backend/api/v3/topo.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/topo.proto)
- [共享类型定义](../../../proto/backend/api/v3/common.proto)
