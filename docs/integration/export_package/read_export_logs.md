# read_export_logs

## 目的与适用场景

当第三方平台希望读取一次插件包导出 workflow 的具体 action / instance 日志时，使用通用 `plugin_workflow` 接口组合。

`export_result` 只能读取 workflow 整体状态和成功后的下载 URL，**不能直接得到具体流程日志**。从 `export_plugin` 响应得到 `workflow_id` 后，按以下四步收敛到具体日志：

1. `POST /api/v3/plugin/workflow/list`，使用 `exact_include_conditions.workflow_id`。
2. `POST /api/v3/plugin/workflow/operation/list`，使用 `workflow_id`。
3. `POST /api/v3/plugin/workflow/operation/instance/list`，使用 `operation_id` 数组。
4. `POST /api/v3/plugin/workflow/operation/instance/log/get`，使用 `oper_inst_id`。

## 输入

读取链路上的输入只有 `workflow_id`（来自 `export_plugin` 成功响应的 `data.workflow_id`）。后续每一步的请求字段以 [plugin_workflow.swagger.json](../../api/swagger/backend/api/v3/plugin_workflow.swagger.json) 和 [plugin_workflow.proto](../../../proto/backend/api/v3/plugin_workflow.proto) 为准：

| 步骤 | 接口                                                      | 关键请求字段                                |
| ---- | --------------------------------------------------------- | ------------------------------------------- |
| ①    | `POST /api/v3/plugin/workflow/list`                       | `exact_include_conditions.workflow_id` 列表 |
| ②    | `POST /api/v3/plugin/workflow/operation/list`             | 顶层 `workflow_id`                          |
| ③    | `POST /api/v3/plugin/workflow/operation/instance/list`    | `operation_id` 列表（可批量）               |
| ④    | `POST /api/v3/plugin/workflow/operation/instance/log/get` | `oper_inst_id`（逐 instance）               |

## 最小 payload 和 curl template

需要 `curl` 和 `jq`。先设置当前部署的 API base URL 与 `workflow_id`。

### 步骤 ① 按 workflow 查询整体记录

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export WORKFLOW_ID="${WORKFLOW_ID}"  # 来自 export_plugin 响应

WORKFLOW_LIST="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/plugin/workflow/list" \
  -H "Content-Type: application/json" \
  -d "$(jq -cn --arg workflow_id "${WORKFLOW_ID}" \
    '{exact_include_conditions: {workflow_id: [$workflow_id]}}')")"

# workflow 整体记录位于 data.items；字段以 plugin_workflow contract 为准
WORKFLOW_STATUS="$(printf '%s' "${WORKFLOW_LIST}" | jq -r '.data.items[0].status // empty')"
```

### 步骤 ② 查询 operation

```bash
OPERATION_LIST="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/plugin/workflow/operation/list" \
  -H "Content-Type: application/json" \
  -d "$(jq -cn --arg workflow_id "${WORKFLOW_ID}" \
    '{workflow_id: $workflow_id}')")"

# operation_id 可作为步骤 ③ 的批量输入
OPERATION_IDS_JSON="$(printf '%s' "${OPERATION_LIST}" | jq -c '[.data.operations[].operation_id]')"
```

### 步骤 ③ 查询 operation instance

```bash
OPERATION_INSTANCE_LIST="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/plugin/workflow/operation/instance/list" \
  -H "Content-Type: application/json" \
  -d "$(jq -cn --argjson operation_ids "${OPERATION_IDS_JSON}" \
    '{operation_id: $operation_ids}')")"

# 每个 instance 的日志使用步骤 ④ 单独读取
OPER_INST_IDS_JSON="$(printf '%s' "${OPERATION_INSTANCE_LIST}" | \
  jq -c '[.data.oper_inst_data[].oper_inst_id]')"
```

### 步骤 ④ 按 oper_inst_id 读取具体日志

```bash
printf '%s' "${OPER_INST_IDS_JSON}" | jq -r '.[]' | while IFS= read -r OPER_INST_ID; do
  curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/plugin/workflow/operation/instance/log/get" \
    -H "Content-Type: application/json" \
    -d "$(jq -cn --arg oper_inst_id "${OPER_INST_ID}" \
      '{oper_inst_id: $oper_inst_id}')"
done
```

上面的模板将步骤 ③ 返回的所有 instance 都传递到步骤 ④。若调用方只需要某个 operation 的最新 instance，应根据步骤 ③ 返回的 instance 信息按自身业务规则选择后再调用步骤 ④；公开 contract 不在本指南中规定选择策略。

## 系统解释

平台以四层结构返回导出 workflow 的执行轨迹：

- `workflow`：一次插件包导出请求；步骤 ① 用 `workflow_id` 收敛到该记录。
- `operation`：该 workflow 下的执行单位；步骤 ② 返回其 `operation_id`。
- `operation_instance`：某个 operation 的执行实例；步骤 ③ 返回其 `oper_inst_id`。
- `operation_instance_log`：某个 instance 的具体日志；步骤 ④ 返回日志内容。

这条组合链路用于读取具体流程日志，与 `export_result` 的 workflow 整体状态和下载 URL 关注点不同。

## 即时输出

每一步的 `data` 字段以 [plugin_workflow.swagger.json](../../api/swagger/backend/api/v3/plugin_workflow.swagger.json) 为准：

- 步骤 ① 返回 workflow 查询结果，通常从 `data.items` 读取匹配记录。
- 步骤 ② 返回 operation 查询结果，从 `data.operations` 读取 `operation_id`。
- 步骤 ③ 接收 `operation_id` 数组并返回 instance 数据，从 `data.oper_inst_data` 读取 `oper_inst_id`。
- 步骤 ④ 返回指定 instance 的日志；具体字段、日志级别和多语言文本以 contract 为准。

这些输出是日志查询结果，不会替代 `export_result` 对导出完成状态和下载 URL 的判断。

## 最终或机器侧可见产物

步骤 ④ 的响应是具体 operation instance 的机器可读日志。调用方可以按 `workflow_id`、`operation_id` 和 `oper_inst_id` 建立导出执行轨迹。

插件包是否可下载仍以 [export_result](export_result.md) 的 `data.is_finish`、`data.status` 和可选 `data.download_url` 为准；日志查询成功不等同于下载 URL 已准备好。

## 重复行为

四个接口均用于读取 workflow、operation、instance 或日志；重复调用不启动新的导出 workflow。公开 contract 不定义推荐轮询间隔、增量日志游标、日志保留时长，以及并发读取时的结果合并行为。

## 失败情况与限制

- 请求结构错误或 unsupported fields 是 API contract 问题，检查 [plugin_workflow.swagger.json](../../api/swagger/backend/api/v3/plugin_workflow.swagger.json)。
- 步骤 ① 找不到 `workflow_id` 对应的记录时，按响应 contract 处理空查询结果。
- 步骤 ② 在不存在或不可查询的 `workflow_id` 上调用时，按响应 contract 处理空结果或 API 错误。
- 步骤 ③ 的 `operation_id` 数组为空时，不会得到可供步骤 ④ 查询的 instance；调用方应先检查查询结果。
- 步骤 ④ 的 `oper_inst_id` 非法或不存在时，按 API 错误处理。
- `export_result` 不返回具体 action / instance 日志；不要从其响应拼接日志地址，也不要调用 file service 内部 endpoint。
- 公开 contract 不定义轮询间隔、最大等待时长、重试窗口、增量读取协议或终态后的日志保留时长。

## Contract 参考

- [接入总览](README.md)
- [启动导出插件包](export_plugin.md)
- [读取导出结果](export_result.md)
- [Plugin workflow Swagger](../../api/swagger/backend/api/v3/plugin_workflow.swagger.json)
- [Plugin workflow Proto](../../../proto/backend/api/v3/plugin_workflow.proto)
