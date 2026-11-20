# list_v2_processes

## 目的与适用场景

`list_v2_processes` 用于查询 V2 兼容插件进程列表。

该能力由 `POST /api/v3/process/list_v2` 提供，分页返回 V2 兼容进程记录，支持按业务、主机、插件名称、进程状态等条件过滤。

## 输入

请求字段以 [process.proto `ProcessV2ListReq`](../../../proto/backend/api/v3/process.proto) 为准：

| 字段                       | Required | 含义                                              |
| -------------------------- | -------- | ------------------------------------------------- |
| `page`                     | no       | 分页参数；`only_count=true` 时可省略              |
| `page.offset`              | no       | 起始偏移量；小于 `0` 时按 `0` 处理                |
| `page.limit`               | yes      | 返回条数上限；查询列表时取值范围为 `1`–`500`      |
| `only_count`               | no       | 是否只返回匹配总数；为 `true` 时 `items` 为空数组 |
| `exact_include_conditions` | no       | 精确匹配条件                                      |
| `fuzzy_include_conditions` | no       | 模糊匹配条件                                      |

常用的 `exact_include_conditions` 字段：

| 字段              | 含义                                                                |
| ----------------- | ------------------------------------------------------------------- |
| `bk_biz_id`       | 业务 ID 列表                                                        |
| `bk_host_id`      | 主机 ID 列表                                                        |
| `plugin_name`     | 插件名称列表                                                        |
| `status`          | 进程状态列表：`init`、`running`、`stopped`、`unregister`、`unknown` |
| `version`         | 进程版本列表                                                        |
| `plugin_pkg_name` | 插件包名称列表                                                      |
| `plugin_group`    | 插件分组列表                                                        |
| `platform_os`     | 操作系统类型列表                                                    |
| `platform_arch`   | CPU 架构列表                                                        |
| `agent_id`        | Agent ID 列表                                                       |
| `generation`      | 进程代际列表                                                        |

`fuzzy_include_conditions` 支持 `name`（进程二进制名称）和 `plugin_pkg_name`（插件包名称）。

## 最小 payload 和 curl template

需要 `curl` 和 `jq`。先设置当前部署的 API base URL、业务 ID 和插件名称。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_BIZ_ID=2
export PLUGIN_NAME="bkunifylogbeat"

V2_LIST_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/process/list_v2" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
  {
    "page": {"offset": 0, "limit": 100},
    "exact_include_conditions": {
      "bk_biz_id": [${BK_BIZ_ID}],
      "plugin_name": ["${PLUGIN_NAME}"],
      "status": ["running"]
    }
  }
EOF
)"

RUNNING_HOST_IDS="$(printf '%s' "${V2_LIST_RESPONSE}" \
  | jq -r '.data.items[].bk_host_id' \
  | paste -sd, -)"
```

数据量大时按 `data.total` 和 `page.offset` 逐页拉取。

## 系统解释

平台把请求解释为一次 V2 兼容进程查询：

- `bk_biz_id`、`bk_host_id`、`plugin_name` 和 `status` 共同收敛目标进程。
- 未传 `bk_biz_id` 时，查询范围会收敛到调用方有 `plugin_view` 权限的业务。
- 目标业务或主机没有对应 V2 兼容记录时返回 `total=0` 与空 `items`。

`process_info.status` 的公开取值：

| 状态         | 含义                                 |
| ------------ | ------------------------------------ |
| `running`    | V2 插件进程正在运行                  |
| `stopped`    | V2 插件进程已停止                    |
| `unknown`    | V2 插件进程状态无法判定              |
| `init`       | 进程记录已创建，但尚未同步到实际状态 |
| `unregister` | 主机上没有该插件的 V2 进程记录       |

## 即时输出

响应包含：

- `data.total`：匹配的 V2 兼容进程总数。
- `data.items[]`：V2 兼容进程列表。
- `data.items[].bk_host_id` / `bk_biz_id` / `plugin_name` / `plugin_pkg_name` / `plugin_group`。
- `data.items[].process_info`：`pid` / `version` / `agent_id` / `auto_start` / `status` / `last_sync_at`。
- `data.items[].process_identity` / `process_controller` / `process_resource` / `process_monitor_policy`。

返回结果是查询时可见的进程记录状态。

## 最终或机器侧可见产物

本接口没有后续 workflow，也不产生机器侧变更。它的可观察产物是一次 V2 兼容进程状态快照：

- `running`：该主机当前存在运行中的 V2 进程。
- `stopped` / `unregister`：该主机当前没有运行中的 V2 进程。
- `unknown`：状态无法判定，不应直接视为已停止。

## 重复行为

读取接口无副作用；重复调用只返回当时的查询结果。

公开 contract 不定义：

- 状态同步延迟上限
- 查询结果的排序规则

## 失败情况与限制

- `only_count=false` 时必须传入有效分页，`page.limit` 取值 `1`–`500`。
- 接口检查调用方的 `plugin_view` 权限，并按授权业务范围收敛结果。
- 目标业务或主机没有 V2 兼容记录时返回空列表，不代表调用失败。
- `unknown` 表示无法判定，不能作为“V2 已停止”的证据。
- 本接口不触发安装、停止、启动或卸载。

## Contract 参考

- [Process Swagger](../../api/swagger/backend/api/v3/process.swagger.json)
- [Process Proto](../../../proto/backend/api/v3/process.proto)
- [V2 进程列表 API](../../../apigw/apidocs/zh/Process_ListV2Processes.md)
