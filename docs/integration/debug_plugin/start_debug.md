# start_debug

## 目的与适用场景

当第三方平台希望在不改变生产进程的前提下，在一组目标主机上以隔离的调试部署方式运行一次插件的调试命令时，使用 `start_debug`。

`start_debug` 与 `install` / `upgrade` 的区别：

| 维度     | `install` / `upgrade` | `start_debug`                             |
| -------- | --------------------- | ----------------------------------------- |
| 目标对象 | 作用于目标 process    | 独立的隔离部署，不复用生产 process        |
| 运行方式 | 按托管配置启动        | 执行插件包 `control.debug` 声明的调试命令 |
| 生命周期 | 常驻                  | 一次性；会话结束自动清理调试现场          |

## 输入

请求字段以 [plugin.proto `PluginStartDebugReq`](../../../proto/backend/api/v3/plugin.proto) 为准：

| 字段                               | Required | 含义                                                    |
| ---------------------------------- | -------- | ------------------------------------------------------- |
| `debug_info.scope`                 | yes      | `ScopeInstance`；解析出的每个目标对应一次独立的执行单位 |
| `debug_info.plugin_name`           | yes      | 要调试的插件名                                          |
| `debug_info.version`               | yes      | 要调试的插件版本                                        |
| `debug_info.config_template_name`  | no       | 配置模板名列表                                          |
| `debug_info.custom_config_context` | no       | 自定义配置上下文                                        |

`ScopeInstance` 的形状：

| 字段           | Required | 含义                             |
| -------------- | -------- | -------------------------------- |
| `granularity`  | yes      | `host` 或 `service_instance`     |
| `bk_biz_id`    | yes      | 业务 ID                          |
| `instance_ids` | yes      | host 或 service instance ID 列表 |

## 最小 payload 和 curl template

示例使用 `instance` scope 和 `host` granularity。需要 `curl` 和 `jq`。先设置当前部署的 API base URL 与目标标识。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_BIZ_ID=2
export BK_HOST_ID=10001
export PLUGIN_NAME="example_plugin"
export PLUGIN_VERSION="1.0.0"

START_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/plugin/start_debug" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
  {
    "debug_info": {
      "scope": {
        "granularity": "host",
        "bk_biz_id": ${BK_BIZ_ID},
        "instance_ids": [${BK_HOST_ID}]
      },
      "plugin_name": "${PLUGIN_NAME}",
      "version": "${PLUGIN_VERSION}",
      "config_template_name": ["base"],
      "custom_config_context": {
        "env": "prod"
      }
    }
  }
EOF
)"

WORKFLOW_ID="$(printf '%s' "${START_RESPONSE}" | jq -r '.data.workflow_id')"
```

多主机时把 `instance_ids` 扩成多个 ID；`scope` 解析出的每个 target 各对应会话内一个独立的执行单位。

## 系统解释

平台会把该请求解释为一次独立的 debug workflow：

- `debug_info.plugin_name` + `debug_info.version` 决定调试用的插件包。
- `debug_info.config_template_name` / `custom_config_context` 与普通 install/upgrade 一致，控制渲染生成的配置。
- 调试命令本身来自插件包 `definition.yaml` 的 `control.debug`；API 请求无法指定运行命令。
- 多目标时，每个解析出的 target 各对应 workflow 内一个独立的执行单位；状态、日志、停止信号都按执行单位隔离。

## 即时输出

start 响应包含 `data.workflow_id`，用于标识本次调试会话。

`workflow_id` 不证明调试命令已经运行、不证明有输出、不证明目标主机已经清理。

## 最终或机器侧可见产物

调试会话期望产生的可观察效果：

| 产物         | 说明                                                                                 |
| ------------ | ------------------------------------------------------------------------------------ |
| 调试过程输出 | 通过通用 `plugin_workflow` 接口读取，调用链见 [read_debug_logs](read_debug_logs.md)  |
| 调试现场清理 | 会话结束后，平台清理本次调试的隔离部署与临时目录；目标主机不再保留本次会话的隔离现场 |

公开 contract 不定义：

- 调试命令的最大运行时长上限（待确认）
- 调试现场在主机上的具体路径
- 调试会话结束后的清理延迟
- 重复 start 同一目标的合并、替换或排队行为

## 重复行为

公开 contract 不定义重复 `start_debug` 调用的合并、替换或排队行为，也不定义 idempotency key、retry window。

## 失败情况与限制

- 请求校验要求 `debug_info` 非空。
- 请求校验要求 `debug_info.plugin_name` 非空。
- 请求校验要求 `debug_info.version` 非空。
- 请求校验要求 `debug_info.scope` 可解析；`scope` 解析失败时平台返回 API 错误。
- 成功的 start 响应只表示调试会话已创建，不证明调试命令已经执行。
- 调试命令本身来自插件包；运行命令由 `control.debug` 声明，请求方无法注入。
- 调试会话是否最终成功，需要通过 `plugin_workflow` 接口读取状态与日志判定。

## Contract 参考

- [接入总览](README.md)
- [读取调试会话](read_debug_logs.md)
- [停止调试](stop_debug.md)
- [Swagger contract](../../api/swagger/backend/api/v3/plugin.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/plugin.proto)
- [术语表](../../concepts/glossary.md)
