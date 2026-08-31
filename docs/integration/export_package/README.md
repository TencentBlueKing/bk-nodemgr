# export_package 接入指南

本文面向需要通过 bk-nodemgr backend API 导出插件包的第三方平台开发者，说明如何启动一次导出、读取 workflow 整体结果、获取最终下载 URL，以及组合查询具体执行日志。

导出启动与结果接口的请求和响应 envelope 以 [pkg_workflow.swagger.json](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json) 为准；字段定义以 [pkg_workflow.proto](../../../proto/backend/api/v3/pkg_workflow.proto) 为准。通用 workflow 日志接口以 [plugin_workflow.swagger.json](../../api/swagger/backend/api/v3/plugin_workflow.swagger.json) 和 [plugin_workflow.proto](../../../proto/backend/api/v3/plugin_workflow.proto) 为准。

## 重要边界

`POST /api/v3/package/workflow/export/plugin` 的成功响应只表示导出 workflow 已创建，不表示插件包已经准备好下载。

`POST /api/v3/package/workflow/export_result` 只用于读取 workflow 整体状态，以及在导出成功完成后读取 `download_url`。它**不能直接得到具体 action / instance 日志**。如需读取具体流程日志，必须按 [read_export_logs](read_export_logs.md) 中的四步组合链路调用通用接口：

1. `POST /api/v3/plugin/workflow/list`：按 `exact_include_conditions.workflow_id` 查询 workflow。
2. `POST /api/v3/plugin/workflow/operation/list`：按 `workflow_id` 查询 operation。
3. `POST /api/v3/plugin/workflow/operation/instance/list`：按 `operation_id` 数组查询 operation instance。
4. `POST /api/v3/plugin/workflow/operation/instance/log/get`：按 `oper_inst_id` 查询具体日志。

不要根据 `export_result` 的响应直接拼接日志地址，也不要调用 file service 的内部 endpoint。

## 接入结果

一次完整的导出接入应使调用方能够：

- 以插件包名称和版本启动导出 workflow。
- 保存启动响应中的 `data.workflow_id`，作为本次导出的查询标识。
- 通过 `export_result` 读取 `data.status` 和 `data.is_finish`。
- 在成功完成时读取可选的 `data.download_url`，并关注可选的 `data.download_url_expired_at`。
- 在需要排查具体执行步骤时，通过四步通用接口组合读取 operation、instance 和日志。

## API-first 快速接入

| 动作                   | 接口或文档                              |
| ---------------------- | --------------------------------------- |
| 启动插件包导出         | [export_plugin](export_plugin.md)       |
| 读取导出状态和下载 URL | [export_result](export_result.md)       |
| 读取具体导出流程日志   | [read_export_logs](read_export_logs.md) |

完整集成流程：

1. 准备要导出的 `plugin_pkg_name` 和 `plugin_pkg_version`。
2. 调用 `POST /api/v3/package/workflow/export/plugin`，从成功响应保存 `data.workflow_id`。
3. 使用该 `workflow_id` 调用 `POST /api/v3/package/workflow/export_result`，读取 workflow 整体状态。
4. 仅在 workflow 成功完成且响应提供 `data.download_url` 时，使用该 URL 下载插件包；同时处理 `data.download_url_expired_at`。
5. 若需要具体 action / instance 日志，按上文四步链路调用通用 `plugin_workflow` 接口。

## 公开接口总览

| 步骤 | Method and path                                           | 关键输入                                    | 主要输出                                        |
| ---- | --------------------------------------------------------- | ------------------------------------------- | ----------------------------------------------- |
| ①    | `POST /api/v3/package/workflow/export/plugin`             | `plugin_pkg_name` / `plugin_pkg_version`    | `data.workflow_id`                              |
| ②    | `POST /api/v3/package/workflow/export_result`             | `workflow_id`                               | `data.status` / `data.is_finish` / 可选下载字段 |
| ③    | `POST /api/v3/plugin/workflow/list`                       | `exact_include_conditions.workflow_id` 列表 | workflow 查询结果                               |
| ④    | `POST /api/v3/plugin/workflow/operation/list`             | `workflow_id`                               | operation 查询结果                              |
| ⑤    | `POST /api/v3/plugin/workflow/operation/instance/list`    | `operation_id` 列表                         | operation instance 查询结果                     |
| ⑥    | `POST /api/v3/plugin/workflow/operation/instance/log/get` | `oper_inst_id`                              | 具体 operation instance 日志                    |

curl template 中的 `BK_NODEMGR_API_BASE` 由调用方提供，表示当前部署的 API base URL。本指南不定义统一 gateway、认证 header、轮询间隔或下载 URL 的确切有效期。

## 重复行为与限制

- 导出启动接口是创建 workflow 的调用；公开 contract 不定义重复提交相同插件名称和版本时的合并、替换、去重或幂等行为。
- `export_result` 与通用 workflow 日志接口用于读取结果，不应被当作导出启动接口重复调用的替代品。
- 公开 contract 不定义推荐轮询频率、最大等待时长、重试窗口、回滚行为或 `download_url_expired_at` 的时间单位与具体有效期。
- 导出 workflow 成功完成前，不应把缺少 `download_url` 解释为可下载地址；成功完成后的下载地址及其过期信息以接口响应为准。

## Contract 参考

- [启动导出插件包](export_plugin.md)
- [读取导出结果](export_result.md)
- [读取具体导出日志](read_export_logs.md)
- [Package workflow Swagger](../../api/swagger/backend/api/v3/pkg_workflow.swagger.json)
- [Package workflow Proto](../../../proto/backend/api/v3/pkg_workflow.proto)
- [Plugin workflow Swagger](../../api/swagger/backend/api/v3/plugin_workflow.swagger.json)
- [Plugin workflow Proto](../../../proto/backend/api/v3/plugin_workflow.proto)
