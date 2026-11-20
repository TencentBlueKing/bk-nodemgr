# sync_v2_alive_process_info

## 目的与适用场景

`sync_v2_alive_process_info` 用于触发指定主机的 V2 兼容插件存活进程信息同步。

该能力由 `POST /api/v3/sync/gse/plugin/process_v2/info/alive` 提供。

该接口只处理 V2 兼容插件进程记录，不处理 V3 插件进程。

该接口所需权限：无。

## 输入

请求字段以 [sync.proto `SyncAlivePluginProcessV2InfoReq`](../../../proto/backend/api/v3/sync.proto) 为准：

| 字段       | Required | 含义                                      |
| ---------- | -------- | ----------------------------------------- |
| `host_ids` | yes      | 待同步的主机 ID 列表，至少包含一个主机 ID |

## 最小 payload 和 curl template

需要 `curl` 和 `jq`。先设置当前部署的 API base URL 和待同步主机。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_HOST_IDS_JSON="[1001,1002]"

SYNC_ALIVE_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/sync/gse/plugin/process_v2/info/alive" \
  -H "Content-Type: application/json" \
  -d "{\"host_ids\": ${BK_HOST_IDS_JSON}}")"

SYNC_ALIVE_TRIGGER_ID="$(printf '%s' "${SYNC_ALIVE_RESPONSE}" | jq -r '.data.trigger_id')"
```

## 系统解释

平台把请求解释为一次 V2 存活进程同步：

- 只处理 `process_v2` 记录中状态为 `running` 的进程，以及状态为 `unknown` 且 `auto_start=true` 的进程。
- 已记录为 `stopped` 的进程不在本接口范围内。
- Agent 在线时，向 GSE 的 `nodeman` 命名空间查询真实进程信息，并回写本地记录。
- Agent 离线时，相关进程被标记为 `unknown`。
- 查询失败或 GSE 未上报的进程保留本地记录状态，不会被标记为 `unknown`。

## 即时输出

响应包含 `data.trigger_id`，表示同步工作流已创建并激活。

`trigger_id` 不表示同步已经完成，也不表示进程状态已经更新。

## 最终或机器侧可见产物

同步期望产生的可观察效果：

- 目标主机的 V2 兼容进程记录被 GSE 实际状态更新。
- Agent 离线主机的相关进程记录被标记为 `unknown`。
- 已停止记录不因本接口改变。

同步结果以进程记录查询结果为准。

## 重复行为

公开 contract 不定义重复调用的合并、替换或去重行为，也不定义 idempotency key 或 retry window。

每次调用都会创建新的同步工作流。

## 失败情况与限制

- `host_ids` 至少包含一个主机 ID；空列表无法创建同步任务。
- 接口成功只表示工作流已创建并激活，不表示同步已经完成。
- 本接口只处理 V2 兼容进程记录。
- 已记录为 `stopped` 的进程不在此接口范围内。
- 公开 contract 不定义同步最大等待时长、失败重试窗口或推荐轮询间隔。

## Contract 参考

- [Sync Swagger](../../api/swagger/backend/api/v3/sync.swagger.json)
- [Sync Proto](../../../proto/backend/api/v3/sync.proto)
- [Sync Alive Plugin Process V2 Info API](../../../apigw/apidocs/zh/Sync_SyncAlivePluginProcessV2Info.md)
