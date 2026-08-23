# read_debug_logs

## 目的与适用场景

当第三方平台希望读取一个调试会话的整体状态、每主机执行进度与流式输出时，使用通用 `plugin_workflow` 接口组合。

调试会话的输出读取路径与节点管理前端"任务历史"日志查看方式一致；只凭 `start_debug` 返回的 `workflow_id` 即可串联出所有主机的调试日志。

## 输入

读取路径上的输入只有 `workflow_id`（来自 `start_debug` 响应）。后续每一步的请求字段以 [plugin_workflow.swagger.json](../../api/swagger/backend/api/v3/plugin_workflow.swagger.json) 为准：

| 步骤 | 接口                                                      | 关键请求字段                                |
| ---- | --------------------------------------------------------- | ------------------------------------------- |
| ①    | `POST /api/v3/plugin/workflow/list`                       | `exact_include_conditions.workflow_id` 列表 |
| ②    | `POST /api/v3/plugin/workflow/operation/list`             | 顶层 `workflow_id`                          |
| ③    | `POST /api/v3/plugin/workflow/operation/instance/list`    | `operation_id` 列表（可批量）               |
| ④    | `POST /api/v3/plugin/workflow/operation/instance/log/get` | `oper_inst_id`（逐 instance）               |

## 最小 payload 和 curl template

需要 `curl` 和 `jq`。先设置当前部署的 API base URL 与 `workflow_id`。

### 步骤 ① 整体概览

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export WORKFLOW_ID="${WORKFLOW_ID}"  # 来自 start_debug 响应

WORKFLOW_LIST="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/plugin/workflow/list" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
  {
    "exact_include_conditions": {
      "workflow_id": ["${WORKFLOW_ID}"]
    }
  }
EOF
)"

# 整体状态位于 items[0].status；多主机调试会带回 bk_host_id 列表
WORKFLOW_STATUS="$(printf '%s' "${WORKFLOW_LIST}" | jq -r '.data.items[0].status')"
```

### 步骤 ② 每主机执行单位

```bash
OPERATION_LIST="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/plugin/workflow/operation/list" \
  -H "Content-Type: application/json" \
  -d "{\"workflow_id\": \"${WORKFLOW_ID}\"}")"

# 每个 operation 对应一台目标主机；取最新 instance 时使用 instance_ids 最后一个元素
OPERATION_IDS="$(printf '%s' "${OPERATION_LIST}" | jq -r '.data.operations[].operation_id' | paste -sd, -)"
```

### 步骤 ③ 每 instance 进度

```bash
OPERATION_INSTANCE_LIST="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/plugin/workflow/operation/instance/list" \
  -H "Content-Type: application/json" \
  -d "{\"operation_id\": [$(printf '%s' "${OPERATION_IDS}" | sed 's/,/","/g; s/^/"/; s/$/"/')]}")"

# 调试流式输出对应的 instance id（按 operation 逐个取）
OPER_INST_IDS="$(printf '%s' "${OPERATION_INSTANCE_LIST}" | jq -r '.data.oper_inst_data[].oper_inst_id' | paste -sd, -)"
```

### 步骤 ④ 逐 instance 取日志

```bash
for OPER_INST_ID in $(printf '%s' "${OPER_INST_IDS}" | tr ',' ' '); do
  curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/plugin/workflow/operation/instance/log/get" \
    -H "Content-Type: application/json" \
    -d "{\"oper_inst_id\": \"${OPER_INST_ID}\"}"
done
```

## 系统解释

平台以四层结构返回调试会话的执行轨迹：

- `workflow`：一次调试会话；`status` 表示整体进度。
- `operation`：每个解析出的目标主机一个；承载该主机的执行轨迹。
- `operation_instance`：每次执行尝试一个；轮询时取 `instance_ids` 最后一个为最新实例。
- `operation_instance_log`：该 instance 的日志；调试过程的流式输出位于其中。

调用链调用顺序与节点管理前端"任务历史"日志查看方式一致：先按 `workflow_id` 收敛到会话，再展开到每主机、每次执行、最终日志。

## 即时输出

每一步的 `data` 字段以 [plugin_workflow.swagger.json](../../api/swagger/backend/api/v3/plugin_workflow.swagger.json) 为准。

调试过程的实时输出位于步骤 ④ 的 `data.oper_inst_logs` 中，每条日志含 `time`（Unix 毫秒）、`level` 与 `text_zh` / `text_en` 字段。

这些输出是状态查询结果，不证明调试会话已经达到期望的最终状态。

## 最终或机器侧可见产物

`workflow/list` 返回的 `status` 表示整体终态。公开 contract 文档化的终态取值（来自响应字段的对外可观察取值）：

| 终态值           | 含义                        |
| ---------------- | --------------------------- |
| `success`        | 全部 execution 单位成功结束 |
| `failed`         | 全部 execution 单位失败结束 |
| `partial_failed` | 部分主机成功部分失败        |

多主机时，步骤 ② 的每个 `operation` 对应一台主机；各 operation 独立读日志、独立判定。

具体每 instance 的状态字段语义以响应 schema 为准；本文档不展开 `life_cycle.state` 的内部枚举。

## 重复行为

读取接口是无副作用的查询；重复调用不改变会话状态。

公开 contract 不定义：

- 推荐轮询间隔
- 增量读取协议（按 `time` 游标去重由调用方自行实现）
- 终态进入后的日志保留时长

## 失败情况与限制

- 请求结构错误或 unsupported fields 是 API contract 问题，检查 [plugin_workflow.swagger.json](../../api/swagger/backend/api/v3/plugin_workflow.swagger.json)。
- 步骤 ① 找不到 `workflow_id` 对应的 workflow：响应 `data.items` 为空。
- 步骤 ② 在不存在的 `workflow_id` 上调用：响应 `data.operations` 为空。
- 步骤 ③ / ④ 的 `operation_id` / `oper_inst_id` 非法：API 错误。
- 整体 `status` 是即时查询结果，不证明目标主机已收敛到期望状态；机器侧验收应结合会话日志与会话结束后的清理结果综合判定。

## Contract 参考

- [接入总览](README.md)
- [启动调试](start_debug.md)
- [停止调试](stop_debug.md)
- [Workflow read Swagger](../../api/swagger/backend/api/v3/plugin_workflow.swagger.json)
- [术语表](../../concepts/glossary.md)
