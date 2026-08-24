# import_package 接入指南

本文面向需要通过 bk-nodemgr backend API 导入插件包到 package 服务的第三方平台开发者，说明接入流程、调用方需要提交的输入、API 即时输出，以及导入期望产生的可观察效果。

`import_*` 与 `import_result` 的请求与响应 envelope 以 [pkg_workflow.swagger.json](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json) 为准；字段以 [pkg_workflow.proto](../../../proto/backend/api/v3/pkg_workflow.proto) 为准。operation / action / log 类型见 [workflow.proto](../../../proto/backend/api/v3/workflow.proto)。领域术语见 [glossary](../../concepts/glossary.md) 与 [plugin/README](../../concepts/plugin/README.md)。

## 接入结果

`import_package` 在 bk-nodemgr package 服务中注册一个由外部下载链接指向的插件包。`import_package` 当前提供 3 个变体，分别对应不同的插件包类型：

| 变体                        | 插件包类型    | 适用场景                 |
| --------------------------- | ------------- | ------------------------ |
| `import_v3_plugin`          | v3 插件包     | 通用 v3 插件包导入       |
| `import_v2_plugin`          | v2 插件包     | 官方 v2 插件包导入       |
| `import_v2_external_plugin` | v2 外部插件包 | 第三方 v2 外部插件包导入 |

三个变体的请求字段一致：

- `filename`：插件包文件名。
- `download_url`：插件包可下载 URL。
- `md5`：插件包 MD5 校验值。

API 响应只确认导入 workflow 已创建；响应本身不证明插件包已经下载、已经校验、已经写入 release 记录，也不证明插件包已经在 package 服务注册为可发布条目。导入是否最终落地，需要通过 `import_result` 接口读取 workflow 执行结果来判定。

> 与 `install` / `upgrade` 等按 host 范围的 workflow 不同，`import_package` 不带 host scope，因此通用 `plugin/workflow/*` 列表查询无法直接命中。本次接入不引导使用通用 workflow 列表查询。

## API-first 快速接入

curl template 按动作拆分：

| 动作                       | 阅读                                                      |
| -------------------------- | --------------------------------------------------------- |
| 触发一次 v3 插件包导入     | [import_v3_plugin](import_v3_plugin.md)                   |
| 触发一次 v2 插件包导入     | [import_v2_plugin](import_v2_plugin.md)                   |
| 触发一次 v2 外部插件包导入 | [import_v2_external_plugin](import_v2_external_plugin.md) |
| 读取导入 workflow 执行结果 | [import_result](import_result.md)                         |

完整集成流程：

1. 准备好要导入的插件包的可下载 URL 与 MD5。
2. 选择对应的 import 变体（`import_v3_plugin` / `import_v2_plugin` / `import_v2_external_plugin`）调用 `POST /api/v3/package/workflow/import/{variant}`，提交 `filename` / `download_url` / `md5`。
3. 从 import 响应保存 `data.workflow_id`，作为本次导入的标识。
4. 调用 `POST /api/v3/package/workflow/import_result`，按 `workflow_id` 读取 workflow 状态与每 operation 的源信息。
5. 根据 `import_result` 的 `data.status` 与 `data.is_finish` 判定导入是否完成；按 `data.operations[]` 的 `oper_inst_logs` 定位具体 action 日志。

## 接入流程

### 1. 确认 API 上下文

公开的 import_package endpoints：

| 步骤                   | Method and path                                           | 即时输出                                               |
| ---------------------- | --------------------------------------------------------- | ------------------------------------------------------ |
| 触发 v3 插件包导入     | `POST /api/v3/package/workflow/import/v3/plugin`          | `data.workflow_id`                                     |
| 触发 v2 插件包导入     | `POST /api/v3/package/workflow/import/v2/plugin`          | `data.workflow_id`                                     |
| 触发 v2 外部插件包导入 | `POST /api/v3/package/workflow/import/v2/external_plugin` | `data.workflow_id`                                     |
| 读取导入结果           | `POST /api/v3/package/workflow/import_result`             | `data.status` / `data.is_finish` / `data.operations[]` |

curl template 中的 `BK_NODEMGR_API_BASE` 由调用方提供，表示当前部署的 API base URL；import_package contract 不定义统一 gateway 或认证 header。

### 2. 准备插件包来源

请求字段直接描述要导入的插件包在外部存储中的位置与校验值：

| 字段           | Required | 含义                                                    |
| -------------- | -------- | ------------------------------------------------------- |
| `filename`     | yes      | 插件包文件名                                            |
| `download_url` | yes      | 插件包可下载 URL；导入过程中 bk-nodemgr 会从该 URL 拉取 |
| `md5`          | yes      | 插件包 MD5 校验值；用于校验下载内容                     |

调用方必须保证 `download_url` 在导入过程中对 bk-nodemgr 可达；`md5` 应当与所下载文件内容一致。

### 3. 触发导入

按插件包类型选择对应端点：

| 插件包类型    | 端点                                                      | 详情                                                      |
| ------------- | --------------------------------------------------------- | --------------------------------------------------------- |
| v3 插件包     | `POST /api/v3/package/workflow/import/v3/plugin`          | [import_v3_plugin](import_v3_plugin.md)                   |
| v2 插件包     | `POST /api/v3/package/workflow/import/v2/plugin`          | [import_v2_plugin](import_v2_plugin.md)                   |
| v2 外部插件包 | `POST /api/v3/package/workflow/import/v2/external_plugin` | [import_v2_external_plugin](import_v2_external_plugin.md) |

提交 `filename` / `download_url` / `md5` 三个字段。

import 响应包含 `data.workflow_id`。把它保存为本次导入的 identity，用于后续 `import_result` 读取。

### 4. 读取导入结果

调用 `POST /api/v3/package/workflow/import_result`，按 `workflow_id` 读取 workflow 状态与每 operation 的源信息：

- `data.status`：workflow 整体状态。
- `data.is_finish`：workflow 是否进入终态。
- `data.operations[]`：本次导入的 operation 源信息数组；每个元素提供 `operation_id` / `last_instance_id` / `oper_inst_logs`（按 action 名索引的 `WorkflowActionData`）/ `extra_execution_logs`（额外的 `WorkflowActionMessage` 日志）。

`operations[]` 的字段即为 "operation 源信息"：第三方平台可以凭 `operation_id` 串联到具体执行，按 `action_name` 读取 action 日志与生命周期。

`data.operations[].operation_id` 与 import 变体的对应关系：

| 触发导入的端点              | 对应 operation_id                   |
| --------------------------- | ----------------------------------- |
| `import_v3_plugin`          | `package_plugin_v3_import`          |
| `import_v2_plugin`          | `package_plugin_v2_import`          |
| `import_v2_external_plugin` | `package_external_plugin_v2_import` |

详情和 curl template：[import_result](import_result.md)。

### 5. 解读即时输出

任一 import 变端点（`import_v3_plugin` / `import_v2_plugin` / `import_v2_external_plugin`）返回 `data.workflow_id`，表示平台已接受导入请求。

`workflow_id` 不证明插件包已经下载、已经校验、已经写入 release 记录。

`POST /api/v3/package/workflow/import_result` 返回 `data.status` / `data.is_finish` / `data.operations[]`；导入是否最终完成，应以 `data.is_finish = true` 且 `data.status` 为终态成功取值、并结合 `data.operations[]` 中各 operation 的 `oper_inst_logs` 判定。

### 6. 观察最终产物

导入期望产生的可观察效果：

- workflow 进入终态：`data.is_finish = true`。
- 至少一个 operation 的 `oper_inst_logs` 中各 action `life_cycle.state` 反映成功（具体状态取值见 [workflow.proto](../../../proto/backend/api/v3/workflow.proto)，不在本文档重复枚举）。
- 插件包作为可发布条目出现在 `package/release/plugin/list` 接口返回中（独立验证项；见下文「导入后验证」）。

公开 contract 不定义：

- 导入的最大等待时长上限（待确认）
- 重复 import 同一 `filename` + `md5` 组合的合并、替换或去重行为
- 导入过程中下载失败或 MD5 不匹配的具体错误返回

### 7. 导入后验证（独立验证项）

`import_result` 反映的是 import workflow 本身的执行结果。第三方平台若需要确认插件包已经作为可发布条目注册到 package 服务，应**独立**调用 `package/release/plugin/list`，按 `exact_include_conditions.name` / `exact_include_conditions.version` 命中查询；命中即视为可发布条目已存在。

这是与 `import_result` 不同的关注点：前者读 workflow 执行结果，后者读 release 注册状态。

### 8. 处理失败与边界

以下内容属于接入边界：

- 请求结构错误或 unsupported fields 是 API contract 问题，检查 [pkg_workflow.swagger.json](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json)。
- 任一 import 变体的成功响应都是即时 API 结果，不是最终注册成功的证明。
- `import_result` 的 `data.is_finish = true` 表示 workflow 进入终态，不单独证明 release 列表已命中。
- 公开 contract 不定义通用 idempotency key、retry window、polling cadence、rollback behavior。

## Contract 参考

- [import_v3_plugin](import_v3_plugin.md)
- [import_v2_plugin](import_v2_plugin.md)
- [import_v2_external_plugin](import_v2_external_plugin.md)
- [import_result](import_result.md)
- [Swagger contract](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/pkg_workflow.proto)
- [Workflow 类型定义](../../../proto/backend/api/v3/workflow.proto)
- [Release list Swagger](../../api/swagger/backend/api/v3/pkg.swagger.json)
- [术语表](../../concepts/glossary.md)
- [Plugin 概念](../../concepts/plugin/README.md)
