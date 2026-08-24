# import_result

## 目的与适用场景

当第三方平台希望读取一次 `import_v3_plugin` / `import_v2_plugin` / `import_v2_external_plugin` 的 workflow 执行结果时，使用 `import_result`。

`import_package` 不带 host scope（与 `install` / `upgrade` 不同），因此通用 `plugin/workflow/*` 列表查询无法直接命中；本接口为 import workflow 单独提供结果读取入口，并通过 `data.operations[]` 提供各 operation 的源信息（`operation_id` / `last_instance_id` / `oper_inst_logs` / `extra_execution_logs`）。

> import 变体与 `data.operations[].operation_id` 的对应关系：

| 触发导入的端点              | 对应 operation_id                   |
| --------------------------- | ----------------------------------- |
| `import_v3_plugin`          | `package_plugin_v3_import`          |
| `import_v2_plugin`          | `package_plugin_v2_import`          |
| `import_v2_external_plugin` | `package_external_plugin_v2_import` |

## 输入

请求字段以 [pkg_workflow.proto `PackageImportResultReq`](../../../proto/backend/api/v3/pkg_workflow.proto) 为准：

| 字段          | Required | 含义                                                                                                |
| ------------- | -------- | --------------------------------------------------------------------------------------------------- |
| `workflow_id` | yes      | `import_v3_plugin` / `import_v2_plugin` / `import_v2_external_plugin` 任一返回的 `data.workflow_id` |

## 最小 payload 和 curl template

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export WORKFLOW_ID="${WORKFLOW_ID}"  # 来自 import 响应（任一变体）

RESULT_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/package/workflow/import_result" \
  -H "Content-Type: application/json" \
  -d "{\"workflow_id\": \"${WORKFLOW_ID}\"}")"

# 整体状态
WORKFLOW_STATUS="$(printf '%s' "${RESULT_RESPONSE}" | jq -r '.data.status')"
IS_FINISH="$(printf '%s' "${RESULT_RESPONSE}" | jq -r '.data.is_finish')"

# 第一个 operation 的第一个 action 日志示例
FIRST_OPERATION_ID="$(printf '%s' "${RESULT_RESPONSE}" | jq -r '.data.operations[0].operation_id')"
```

## 系统解释

平台会以"workflow 整体 + 各 operation 源信息"两层结构返回 import 的执行轨迹：

- `data.status`：workflow 整体状态（具体取值见 `workflow.proto` 中 `WorkflowLifeCycle.state`，本文档不重复枚举）。
- `data.is_finish`：是否进入终态的布尔标志。
- `data.operations[]`：本次 import 的 operation 源信息数组；每个元素提供以下四个字段（operation 源信息）：

| 字段                   | 含义                                                                                                                                                            |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `operation_id`         | 该 operation 的标识                                                                                                                                             |
| `last_instance_id`     | 最近一次 instance 的标识（用于串联到 instance 详情）                                                                                                            |
| `oper_inst_logs`       | 该 instance 下按 `action_name` 索引的 `WorkflowActionData` 字典；含 `life_cycle` / `message.logs` / `display_name_zh` / `display_name_en` / `sub_workflow_refs` |
| `extra_execution_logs` | 额外的 `WorkflowActionMessage`，存放主流程之外的执行日志                                                                                                        |

operation / action / log 字段语义以 [workflow.proto](../../../proto/backend/api/v3/workflow.proto) 为准。

按 import 变体区分，operation 内部的 `action_name` 集合如下：

| 触发导入的端点              | operation_id                        | action_name 集合                                                                                                                                                        |
| --------------------------- | ----------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `import_v3_plugin`          | `package_plugin_v3_import`          | `package_import_plugin_v3_pkg_fetch_and_upload` / `package_publish_plugin_v3_pkg` / `package_release_plugin_enable` / `package_release_plugin_hidden`                   |
| `import_v2_plugin`          | `package_plugin_v2_import`          | `package_import_plugin_v2_pkg_fetch_and_upload` / `package_publish_plugin_v2_pkg` / `package_release_plugin_enable` / `package_release_plugin_hidden`                   |
| `import_v2_external_plugin` | `package_external_plugin_v2_import` | `package_import_external_plugin_v2_pkg_fetch_and_upload` / `package_publish_external_plugin_v2_pkg` / `package_release_plugin_enable` / `package_release_plugin_hidden` |

## 即时输出

`data` 字段以 [pkg_workflow.swagger.json](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json) 为准：

| 字段                | 类型   | 含义                                |
| ------------------- | ------ | ----------------------------------- |
| `data.status`       | string | workflow 整体状态                   |
| `data.is_finish`    | bool   | 是否进入终态                        |
| `data.operations[]` | array  | 各 operation 源信息；元素字段见下表 |

`data.operations[]` 元素字段（operation 源信息）：

| 字段                   | 类型   | 含义                                                                                                                                                              |
| ---------------------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `operation_id`         | string | 该 operation 的标识                                                                                                                                               |
| `last_instance_id`     | string | 最近一次 instance 的标识                                                                                                                                          |
| `oper_inst_logs`       | object | 按 `action_name` 索引的 `WorkflowActionData` 字典；`WorkflowActionData` 含 `life_cycle` / `message` / `display_name_zh` / `display_name_en` / `sub_workflow_refs` |
| `extra_execution_logs` | object | 额外的 `WorkflowActionMessage`（含 `logs` 数组）                                                                                                                  |

`WorkflowActionData` / `WorkflowActionMessage` / `WorkflowLifeCycle` 的字段定义见 [workflow.proto](../../../proto/backend/api/v3/workflow.proto)，本文档不重复枚举。

## 最终或机器侧可见产物

`import_result` 反映的是 import workflow 本身的执行结果。workflow 完成的可观察效果：

| 产物           | 说明                                                                                                   |
| -------------- | ------------------------------------------------------------------------------------------------------ |
| workflow 终态  | `data.is_finish = true`；`data.status` 反映终态取值（具体取值以 `workflow.proto` 为准）                |
| operation 终态 | `data.operations[]` 中各 operation 的 `oper_inst_logs`，按 `action_name` 读取各 action 的 `life_cycle` |

`import_result` 不直接证明 release 列表已命中；如需确认插件包已作为可发布条目注册到 package 服务，应独立调用 `package/release/plugin/list` 按 `name` + `version` 命中查询（详见 [接入总览](README.md) 第 7 节）。

## 重复行为

读取接口是无副作用的查询；重复调用不改变 workflow 状态。

公开 contract 不定义：

- 推荐轮询间隔
- 导入完成到 release 列表可命中的最大延迟
- 同一 `workflow_id` 的多次 `import_result` 调用在并发场景下的取值合并行为

## 失败情况与限制

- 请求结构错误或 unsupported fields 是 API contract 问题，检查 [pkg_workflow.swagger.json](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json)。
- `workflow_id` 对应的 workflow 不存在时，平台返回 API 错误。
- `data.operations[]` 为空表示本次 import 未产出任何 operation（可能为异常状态）；具体含义以平台返回为准。
- `data.is_finish = true` 表示 workflow 已进入终态；`data.status` 仍可能为非成功取值，第三方平台应结合 `data.operations[]` 的 `oper_inst_logs` 判定是否成功。
- `import_result` 反映的是 import workflow 本身的执行结果；不直接证明 release 列表已命中。

## Contract 参考

- [接入总览](README.md)
- [触发 v3 插件包导入](import_v3_plugin.md)
- [触发 v2 插件包导入](import_v2_plugin.md)
- [触发 v2 外部插件包导入](import_v2_external_plugin.md)
- [Swagger contract](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/pkg_workflow.proto)
- [Workflow 类型定义](../../../proto/backend/api/v3/workflow.proto)
- [Release list Swagger](../../api/swagger/backend/api/v3/pkg.swagger.json)
