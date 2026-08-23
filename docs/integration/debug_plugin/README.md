# debug_plugin 接入指南

本文面向需要通过 bk-nodemgr backend API 在目标主机上以隔离方式运行一次插件调试命令的第三方平台开发者，说明接入流程、调用方需要提交的输入、API 即时输出，以及调试会话期望产生的可观察效果。

`start_debug` / `stop_debug` 的请求与响应 envelope 以 [plugin.swagger.json](../../api/swagger/backend/api/v3/plugin.swagger.json) 为准；`DebugInfo` 字段以 [plugin.proto](../../../proto/backend/api/v3/plugin.proto) 为准；调试输出读取复用通用 [plugin_workflow.swagger.json](../../api/swagger/backend/api/v3/plugin_workflow.swagger.json) 接口。领域术语见 [glossary](../../concepts/glossary.md) 与 [plugin/README](../../concepts/plugin/README.md)。

## 接入结果

`debug_plugin` 在一组目标主机上以一次性的调试会话方式运行插件包的调试命令，实时流式返回命令输出，会话结束后自动清理调试现场。

第三方平台提交：

- `scope`：哪些机器或 service instance 会运行调试命令。
- `plugin_name` + `version`：要调试的插件。
- （可选）`config_template_name`、`custom_config_context`：与普通 install/upgrade 行为一致。

API 响应只确认调试 workflow 已创建；响应本身不证明调试命令已经执行、已经产出有效输出，也不证明目标主机已经清理。

## API-first 快速接入

curl template 按动作拆分：

| 动作                     | 阅读                                  |
| ------------------------ | ------------------------------------- |
| 启动一次调试会话         | [start_debug](start_debug.md)         |
| 读取调试状态与流式输出   | [read_debug_logs](read_debug_logs.md) |
| 请求停止运行中的调试会话 | [stop_debug](stop_debug.md)           |

完整集成流程：

1. 根据目标构造 `debug_info.scope`。
2. 调用 `POST /api/v3/plugin/start_debug`，提交 `debug_info`。
3. 从 start 响应保存 `data.workflow_id`，作为本次调试会话的标识。
4. 用 `workflow_id` 轮询通用 `plugin_workflow` 接口，读取整体状态、每主机进度与流式输出。
5. （可选）调用 `POST /api/v3/plugin/stop_debug`，请求停止运行中的调试会话。
6. 等待 `workflow/list` 返回的 `status` 进入终态，按 `read_debug_logs` 给出的判定规则给出最终结论。

## 接入流程

### 1. 确认 API 上下文

公开的 debug_plugin endpoints：

| 步骤                 | Method and path                                           | 即时输出              |
| -------------------- | --------------------------------------------------------- | --------------------- |
| 启动调试             | `POST /api/v3/plugin/start_debug`                         | `data.workflow_id`    |
| 请求停止调试         | `POST /api/v3/plugin/stop_debug`                          | 空 `data`             |
| 读取调试会话         | `POST /api/v3/plugin/workflow/list`                       | `data.items`          |
| 读取每主机 operation | `POST /api/v3/plugin/workflow/operation/list`             | `data.operations`     |
| 读取每 instance 进度 | `POST /api/v3/plugin/workflow/operation/instance/list`    | `data.oper_inst_data` |
| 读取日志             | `POST /api/v3/plugin/workflow/operation/instance/log/get` | `data.oper_inst_logs` |

curl template 中的 `BK_NODEMGR_API_BASE` 由调用方提供，表示当前部署的 API base URL；debug_plugin contract 不定义统一 gateway 或认证 header。

### 2. 把目标转换成 scope

`debug_info.scope` 告诉 bk-nodemgr 哪些目标主机上应运行调试命令。`scope` 的 `ScopeInstance` 形状与 deploy_policy 相同：

| 字段           | Required | 含义                             |
| -------------- | -------- | -------------------------------- |
| `granularity`  | yes      | `host` 或 `service_instance`     |
| `bk_biz_id`    | yes      | 业务 ID                          |
| `instance_ids` | yes      | host 或 service instance ID 列表 |

scope 解析出的每个目标各对应调试会话内一台主机的独立执行单位。多台主机会产生多个 operation。

### 3. 启动调试

调用 `POST /api/v3/plugin/start_debug`，提交 `debug_info`：

| 字段                               | Required | 含义             |
| ---------------------------------- | -------- | ---------------- |
| `debug_info.scope`                 | yes      | 目标范围         |
| `debug_info.plugin_name`           | yes      | 插件名           |
| `debug_info.version`               | yes      | 插件版本         |
| `debug_info.config_template_name`  | no       | 配置模板名列表   |
| `debug_info.custom_config_context` | no       | 自定义配置上下文 |

start 响应包含 `data.workflow_id`。把它保存为本次调试会话 identity，用于后续读取或停止。

详情和 curl template：[start_debug](start_debug.md)。

### 4. 解读即时输出

`POST /api/v3/plugin/start_debug` 返回 `data.workflow_id`，表示平台已接受调试请求。

`workflow_id` 不证明调试命令已经运行、不证明有输出、不证明目标主机已经清理。

`POST /api/v3/plugin/stop_debug` 是非阻塞：响应表示停止信号已经写入；运行中的调试进程是否已终止、是否已完成清理，需要通过 `workflow/list` 后续状态确认。

### 5. 观察最终产物

调试会话是一次性 workflow：

- 调试命令的实际内容来自插件包 `control.debug` 声明，API 请求无法指定运行命令。
- 命令在目标主机的插件部署目录内执行（与 deploy_policy 生成的 `<plugin_home>` 一致）。
- 调试过程输出需要通过通用 `plugin_workflow` 接口读取；具体调用链与节点管理前端"任务历史"日志查看方式一致。
- 会话结束（命令自然退出、被 stop 信号终止或超时）后，系统清理本次调试的隔离现场。

调用链、轮询与终态判定的 curl template：[read_debug_logs](read_debug_logs.md)。

机器侧验收应在目标主机核对以下两点：

- 调试过程输出：参见 [read_debug_logs](read_debug_logs.md) 的最终判定规则。
- 调试现场清理：会话结束后，目标主机不再保留本次调试的隔离部署。

### 6. 处理失败与边界

以下内容属于接入边界：

- 请求结构错误或 unsupported fields 是 API contract 问题，检查 [plugin.swagger.json](../../api/swagger/backend/api/v3/plugin.swagger.json)。
- `scope` 解析失败是 scope 问题，检查 [deploy_policy/scope 概念](../../concepts/deploy_policy/scope.md)。
- `start_debug` 成功响应是即时 API 结果，不是最终调试成功。
- `stop_debug` 是非阻塞；信号生效时延由系统侧处理，不在公开 contract 中量化。
- 调试命令本身来自插件包 `control.debug`；请求方无法也无法通过 API 注入自定义命令。
- 公开 contract 不定义通用 idempotency key、retry window、polling cadence、session token 形式与 rollback behavior。

## Contract 参考

- [start_debug](start_debug.md)
- [stop_debug](stop_debug.md)
- [read_debug_logs](read_debug_logs.md)
- [Swagger contract](../../api/swagger/backend/api/v3/plugin.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/plugin.proto)
- [Workflow read Swagger](../../api/swagger/backend/api/v3/plugin_workflow.swagger.json)
- [术语表](../../concepts/glossary.md)
- [Plugin 概念](../../concepts/plugin/README.md)
