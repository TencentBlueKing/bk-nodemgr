# get_business_host_count

## 目的与适用场景

当调用方需要批量获取若干业务下的主机数量时，使用 `get_business_host_count`。

## 输入

请求字段以 [topo.proto `TopoBusinessHostCountGetReq`](../../../proto/backend/api/v3/topo.proto) 为准：

| 字段        | Required | 含义                                                    |
| ----------- | -------- | ------------------------------------------------------- |
| `bk_biz_id` | yes      | 待查询主机数的业务 ID 列表（[]int64）；可一次传多个业务 |

`bk_biz_id` 为空数组时，平台行为在本文档 contract 中未定义（待确认）。

## 最小 payload 和 curl template

需要 `curl` 和 `jq`。先设置当前部署的 API base URL 与目标业务 ID 列表。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_BIZ_ID=2

COUNT_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/topo/business/host_count/get" \
  -H "Content-Type: application/json" \
  -d "{\"bk_biz_id\": [${BK_BIZ_ID}]}")"

BIZ_HOST_COUNT="$(printf '%s' "${COUNT_RESPONSE}" | jq -r '.data.items[] | select(.bk_biz_id == '"${BK_BIZ_ID}"') | .host_count')"
```

一次查多个业务：

```bash
COUNT_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/topo/business/host_count/get" \
  -H "Content-Type: application/json" \
  -d '{
    "bk_biz_id": [2, 3, 5]
  }')"
```

## 系统解释

平台把该请求解释为按 `bk_biz_id` 列表批量查询每个业务下的主机数：

- 请求里的 `bk_biz_id` 是输入列表；返回的 `data.items[]` 顺序与请求列表未必一致，调用方应按 `bk_biz_id` 对齐。
- 业务无访问权限时，该业务不会出现在 `data.items[]` 中（不会因权限不足而失败）。
- `host_count` 是该业务下所有主机的总数（按 [topo.proto `BusinessHostCount`](../../../proto/backend/api/v3/topo.proto) 聚合），与 `list_host` 在该业务下分页查询的 `data.total` 含义不同：前者是"业务下所有主机"，后者是按 `exact_include_conditions` + `fuzzy_include_conditions` 过滤后聚合。
- `host_count` 与 `get_business_inst_topo` 的根节点 `host_count` 也不同：前者按业务聚合，后者按 cmdb 拓扑节点聚合（含子节点递归、含空闲池与自定义拓扑分组规则）。

## 即时输出

`data` 字段以 [topo.swagger.json `v3TopoBusinessHostCountGetResp`](../../api/swagger/backend/api/v3/topo.swagger.json) 为准：

| 字段           | 类型                       | 含义                              |
| -------------- | -------------------------- | --------------------------------- |
| `data.items[]` | array\<BusinessHostCount\> | 每项含 `bk_biz_id` / `host_count` |

`BusinessHostCount` 字段：

| 字段         | 类型  | 含义                            |
| ------------ | ----- | ------------------------------- |
| `bk_biz_id`  | int64 | 业务 ID；与请求列表中的元素对应 |
| `host_count` | int64 | 该业务下的主机总数              |

## 最终或机器侧可见产物

`get_business_host_count` 是只读查询；不修改后端数据。

| 产物       | 来源                                                             |
| ---------- | ---------------------------------------------------------------- |
| 业务主机数 | `data.items[]`；按 `bk_biz_id` 对齐到 `list_business` 的 `items` |

`host_count` 不反映任何过滤条件后的主机数；它只是"该业务下所有主机"的总数。

## 重复行为

读取接口是无副作用的查询；重复调用不改变后端数据。

公开 contract 不定义：

- 推荐轮询间隔
- 同一 `bk_biz_id` 列表的多次调用在并发场景下的取值合并行为
- `host_count` 随时间变化的更新频率

## 失败情况与限制

- 请求校验失败（`bk_biz_id` 字段缺失等）时返回 API 错误。
- 当前账号对某个业务无访问权限时，该业务不会出现在 `data.items[]` 中（不会因权限不足而失败）。
- 业务无主机时 `host_count = 0`；是合法响应。
- `bk_biz_id` 为空数组时的行为在本文档 contract 中未定义（待确认）。
- `host_count` 的口径差异：见 [系统解释](#系统解释)。

## Contract 参考

- [接入总览](README.md)
- [拉取业务列表](list_business.md)
- [取业务实例拓扑](get_business_inst_topo.md)
- [分页查询主机](list_host.md)
- [Swagger contract](../../api/swagger/backend/api/v3/topo.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/topo.proto)
