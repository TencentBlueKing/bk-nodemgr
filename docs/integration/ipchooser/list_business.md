# list_business

## 目的与适用场景

当调用方需要按条件分页查询当前账号可访问的业务列表时，使用 `list_business`。

## 输入

请求字段以 [topo.proto `TopoBusinessListReq`](../../../proto/backend/api/v3/topo.proto) 为准：

| 字段                       | Required | 含义                                                     |
| -------------------------- | -------- | -------------------------------------------------------- |
| `page`                     | no       | `Page.offset` / `Page.limit`；不传时按平台默认（待确认） |
| `only_count`               | no       | 仅返回 `total`，不返回 `items`；用于分页前先取总数       |
| `exact_include_conditions` | no       | 精确包含；当前支持 `bk_biz_id`（[]int64）                |
| `fuzzy_include_conditions` | no       | 模糊包含；当前支持 `bk_biz_name`（[]string）             |

`exact_include_conditions` 与 `fuzzy_include_conditions` 可同时使用；具体行为以平台返回为准（待确认）。

## 最小 payload 和 curl template

需要 `curl` 和 `jq`。先设置当前部署的 API base URL。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"

LIST_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/topo/business/list" \
  -H "Content-Type: application/json" \
  -d '{
    "page": {"offset": 0, "limit": 100},
    "exact_include_conditions": {
      "bk_biz_id": []
    }
  }')"

TOTAL="$(printf '%s' "${LIST_RESPONSE}" | jq -r '.data.total')"
FIRST_BIZ_ID="$(printf '%s' "${LIST_RESPONSE}" | jq -r '.data.items[0].bk_biz_id')"
```

按业务名模糊匹配：

```bash
SEARCH_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/topo/business/list" \
  -H "Content-Type: application/json" \
  -d '{
    "page": {"offset": 0, "limit": 100},
    "fuzzy_include_conditions": {
      "bk_biz_name": ["蓝鲸"]
    }
  }')"
```

## 系统解释

平台把该请求解释为按条件分页查询当前账号可访问的业务：

- `exact_include_conditions.bk_biz_id` 决定精确匹配的业务 ID 集合；空数组表示不限制。
- `fuzzy_include_conditions.bk_biz_name` 决定按业务名模糊匹配的关键字集合；具体匹配规则（大小写、前后缀、通配符）以平台返回为准（待确认）。
- `page.offset` / `page.limit` 决定分页窗口；`total` 表示满足条件的业务总数。
- 返回的 `items` 中每个 `Business` 含 `tenant_id` / `bk_biz_id` / `bk_biz_name`（来自 [common.proto `Business`](../../../proto/backend/api/v3/common.proto)）。
- `bk_biz_id` 是后续 3 个 endpoint（`get_business_host_count` / `get_business_inst_topo` / `list_host`）的主入参，调用方需要把它从 `items[].bk_biz_id` 中保存下来。

## 即时输出

`data` 字段以 [topo.swagger.json `v3TopoBusinessListResp`](../../api/swagger/backend/api/v3/topo.swagger.json) 为准：

| 字段           | 类型              | 含义                                                       |
| -------------- | ----------------- | ---------------------------------------------------------- |
| `data.total`   | int64             | 满足条件的业务总数                                         |
| `data.items[]` | array\<Business\> | 业务数组；每项含 `tenant_id` / `bk_biz_id` / `bk_biz_name` |

`Business` 字段（[common.proto](../../../proto/backend/api/v3/common.proto)）：

| 字段          | 类型   | 含义                            |
| ------------- | ------ | ------------------------------- |
| `tenant_id`   | string | 租户 ID                         |
| `bk_biz_id`   | int64  | 业务 ID；后续 endpoint 的主入参 |
| `bk_biz_name` | string | 业务名                          |

## 最终或机器侧可见产物

`list_business` 是只读查询；不修改后端数据。

| 产物     | 来源                                                |
| -------- | --------------------------------------------------- |
| 业务列表 | `data.items[]`；每项的 `bk_biz_id` 与 `bk_biz_name` |
| 业务总数 | `data.total`                                        |

## 重复行为

读取接口是无副作用的查询；重复调用不改变后端数据。

公开 contract 不定义：

- 推荐轮询间隔
- 同一 `fuzzy_include_conditions` 的多次调用在并发场景下的取值合并行为

## 失败情况与限制

- 请求校验失败（`page.limit` 非法等）时返回 API 错误。
- 当前账号对某个业务无访问权限时，该业务不会出现在 `data.items[]` 中（不会因权限不足而失败）。
- 业务总数为 0 时返回 `data.total = 0` 且 `data.items = []`；是合法响应。
- `fuzzy_include_conditions.bk_biz_name` 的具体匹配规则（大小写、前后缀、通配符）在本文档 contract 中未定义（待确认）。
- `page.limit` 的上限在本文档 contract 中未定义（待确认）。
- `only_count = true` 时只返回 `data.total`，`data.items` 行为以平台为准（待确认）。

## Contract 参考

- [接入总览](README.md)
- [批量取业务主机数](get_business_host_count.md)
- [取业务实例拓扑](get_business_inst_topo.md)
- [分页查询主机](list_host.md)
- [Proto variant 定义](../../../proto/backend/api/v3/topo.proto)
- [共享类型定义](../../../proto/backend/api/v3/common.proto)
