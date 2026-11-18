# stop_v2

## 目的与适用场景

`stop_v2` 用于批量停止指定主机上的 V2 插件进程。

该能力由 `POST /api/v3/plugin/stop_v2` 提供。接口异步执行停止任务并返回 `workflow_id`，可通过插件 workflow 接口查询执行结果。

## 输入

请求字段以 [plugin.proto `PluginStopV2Req`](../../../proto/backend/api/v3/plugin.proto) 为准：

| 字段                   | Required | 含义                                   |
| ---------------------- | -------- | -------------------------------------- |
| `plugin`               | yes      | 待停止的 V2 插件列表，至少包含一个元素 |
| `plugin[].bk_host_id`  | yes      | 目标主机 ID，必须为正整数              |
| `plugin[].plugin_name` | yes      | V2 插件名称，不能为空                  |

## 最小 payload 和 curl template

需要 `curl` 和 `jq`。先设置当前部署的 API base URL、目标主机和插件名称。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_HOST_ID=1001
export PLUGIN_NAME="bkunifylogbeat"

STOP_V2_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/plugin/stop_v2" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
  {
    "plugin": [
      {
        "bk_host_id": ${BK_HOST_ID},
        "plugin_name": "${PLUGIN_NAME}"
      }
    ]
  }
EOF
)"

STOP_V2_WORKFLOW_ID="$(printf '%s' "${STOP_V2_RESPONSE}" | jq -r '.data.workflow_id')"
```

多主机停止时在 `plugin` 数组中追加元素：

```json
{
  "plugin": [
    { "bk_host_id": 1001, "plugin_name": "bkunifylogbeat" },
    { "bk_host_id": 1002, "plugin_name": "bkunifylogbeat" }
  ]
}
```

## 系统解释

平台把请求解释为一次 V2 插件停止：

- 按 `bk_host_id` 和 `plugin_name` 定位目标主机上的 V2 插件进程。
- 停止 V2 进程并解除其 GSE 托管。

## 即时输出

响应包含 `data.workflow_id`，表示停止 workflow 已创建。

`workflow_id` 不证明 V2 进程已经停止；必须等待 workflow 进入成功终态。

## 最终或机器侧可见产物

停止期望产生的可观察效果：

| 产物        | 说明                                                         |
| ----------- | ------------------------------------------------------------ |
| V2 进程停止 | V2 进程记录状态为 `stopped`                                  |
| V2 托管解除 | V2 进程不再处于 GSE 托管运行状态                             |
| 停止结果    | 通过 `plugin/workflow/*` 读取停止 workflow，确认进入成功终态 |

## 重复行为

公开 contract 不定义重复 `stop_v2` 调用的合并、替换或去重行为，也不定义 idempotency key 或 retry window。

重复调用会创建新的停止 workflow。V2 进程已经停止时，停止步骤可以提前结束；但这不等同于公开 contract 承诺了完整幂等语义。

## 失败情况与限制

- `plugin` 至少包含一个元素；`bk_host_id` 必须为正整数；`plugin_name` 不能为空。
- `stop_v2` 是异步 workflow；成功响应只表示停止任务已提交。
- 如果 V2 仍为 `running` 或 `unknown`，不能视为停止成功；应查看 workflow 日志或重试。
- `stop_v2` 只停止 V2 进程，不卸载 V2 二进制、配置或运行目录。
- 公开 contract 不定义单批主机数量、并发上限、失败重试窗口或推荐轮询间隔。

## Contract 参考

- [Plugin Swagger](../../api/swagger/backend/api/v3/plugin.swagger.json)
- [Plugin Proto](../../../proto/backend/api/v3/plugin.proto)
- [Plugin Stop V2 API](../../../apigw/apidocs/zh/Plugin_StopPluginV2.md)
