# stop_debug

## 目的与适用场景

当第三方平台希望请求停止一个运行中的调试会话时，使用 `stop_debug`。

`stop_debug` 是非阻塞：接口只请求终止运行中的调试命令，会话是否已经终止、是否已经完成清理，需要通过 `plugin_workflow` 接口后续读取确认。

## 输入

请求字段以 [plugin.proto `PluginStopDebugReq`](../../../proto/backend/api/v3/plugin.proto) 为准：

| 字段          | Required | 含义                                    |
| ------------- | -------- | --------------------------------------- |
| `workflow_id` | yes      | `start_debug` 返回的 `data.workflow_id` |

## 最小 payload 和 curl template

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export WORKFLOW_ID="${WORKFLOW_ID}"  # 来自 start_debug 响应

curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/plugin/stop_debug" \
  -H "Content-Type: application/json" \
  -d "{\"workflow_id\": \"${WORKFLOW_ID}\"}"
```

## 系统解释

平台会请求本次调试会话停止运行中的调试命令。`stop_debug` 的语义是写信号，不是同步终止；运行中的调试进程由平台侧处理信号后自行终止，随后会话进入终态并清理调试现场。

公开 contract 不定义：

- 停止信号从写入到运行进程生效的最大时延
- 已经被自动清理的会话再次 `stop_debug` 的最终行为（除返回成功外的更多语义未在 contract 中定义）

## 即时输出

响应 `data` 为空对象；HTTP 成功响应只表示停止信号已写入。

成功的 `stop_debug` 响应不证明调试命令已终止、不证明会话已进入终态、不证明目标主机已清理。

## 最终或机器侧可见产物

调用 `stop_debug` 后，平台期望产生的可观察效果：

| 产物         | 说明                                       |
| ------------ | ------------------------------------------ |
| 调试进程终止 | 运行中的调试命令被请求终止                 |
| 调试现场清理 | 会话进入终态后，平台清理本次调试的隔离现场 |

最终状态与终态判定以 `plugin_workflow` 接口为准，详见 [read_debug_logs](read_debug_logs.md)。

## 重复行为

公开 contract 不定义重复 `stop_debug` 调用的合并或幂等行为。已处于终态的会话再次 `stop_debug` 的具体返回语义未在 contract 中定义。

## 失败情况与限制

平台对 `stop_debug` 的前置校验：

| 条件                                              | 行为     |
| ------------------------------------------------- | -------- |
| `workflow_id` 对应的 workflow 不存在              | API 错误 |
| `workflow_id` 对应的 workflow 不是 debug 类型     | API 错误 |
| `workflow_id` 对应的 workflow 已处于终态          | API 错误 |
| `workflow_id` 对应的 workflow 下无 execution 单位 | API 错误 |

`stop_debug` 是非阻塞；成功响应不证明调试命令已经终止或会话已经清理。

## Contract 参考

- [接入总览](README.md)
- [启动调试](start_debug.md)
- [读取调试会话](read_debug_logs.md)
- [Swagger contract](../../api/swagger/backend/api/v3/plugin.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/plugin.proto)
