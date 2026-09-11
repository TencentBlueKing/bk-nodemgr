# specify_plugin

See [Workflow usage](workflow.md) for execution result polling.

## 目的与适用场景

当第三方平台希望目标节点拥有指定 `plugin_name` 和 `version` 的插件时，使用 `specify_plugin`。

现有概念文档对该 mode 的定义是：确保目标节点有指定插件名称和版本；如果插件不存在，则安装；如果版本不匹配，则升级。

## 输入

该 mode 的 desired-state 字段：

| Field                   | Required | 含义                             |
| ----------------------- | -------- | -------------------------------- |
| `plugin_name`           | yes      | 目标节点上需要确保存在的插件名称 |
| `version`               | yes      | 需要确保的插件版本               |
| `custom_config_context` | no       | 以结构化对象传入的自定义值       |

策略还需要 `scopes`，以便 bk-nodemgr 解析目标节点。支持的 scope 形式见 [scope](../../concepts/deploy_policy/scope.md)。

## 最小 payload 和 curl template

示例使用 `instance` scope 和 `host` granularity。需要 `curl` 和 `jq`。先设置当前部署的 API base URL 与目标标识。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_BIZ_ID=2
export BK_HOST_ID=10001

CREATE_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/create" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
  {
    "name": "ensure-gse-plugin-example",
    "description": "Ensure a named plugin version on selected hosts",
    "enabled": true,
    "scopes": [
      {
        "type": "instance",
        "scope": {
          "granularity": "host",
          "bk_biz_id": ${BK_BIZ_ID},
          "instance_ids": [${BK_HOST_ID}]
        }
      }
    ],
    "specs": [
      {
        "type": "specify_plugin",
        "param": {
          "plugin_name": "example_plugin",
          "version": "1.0.0",
          "custom_config_context": {
            "env": "prod"
          }
        }
      }
    ]
  }
EOF
)"

DEPLOY_POLICY_ID="$(printf '%s' "${CREATE_RESPONSE}" | jq -r '.data.deploy_policy_id')"
```

使用返回的 ID 执行策略，并提取 `data.workflow_id`：

```bash
EXECUTE_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/execute" \
  -H "Content-Type: application/json" \
  -d "{\"deploy_policy_id\": ${DEPLOY_POLICY_ID}}")"

WORKFLOW_ID="$(printf '%s' "${EXECUTE_RESPONSE}" | jq -r '.data.workflow_id')"
```

## 系统解释

平台会把该 spec 解释为插件 desired state：`scopes` 选中的每个目标都应拥有 `plugin_name`，且版本为 `version`。

根据概念文档，插件缺失会触发安装，版本不匹配会触发升级。当前执行链会按 generation、OS、arch、`plugin_pkg_name` 和 `version` 选择 release package，渲染配置，传输 package，执行安装器并启动目标进程。

## 即时输出

create 返回 `data.deploy_policy_id`，用于标识已创建策略。

execute 返回 `data.workflow_id`，用于标识已发起的执行任务。

这些即时输出不是最终机器状态证明。

## 最终或机器侧可见产物

当前实现从插件注册信息取得 `plugin_group` 和 `plugin_pkg_name`，再按以下公式生成机器目录：

```text
<plugin_home> = <base_deploy_dir>/<deploy_env>/plugin/<plugin_group>/<plugin_name>
```

Windows 使用 `\` 作为路径分隔符。`<base_deploy_dir>` 来自目标节点的 plugin deployment 配置，可被 Network Unit 的 custom deploy config 覆盖；`<deploy_env>` 来自当前部署环境。

执行成功后可以在目标机核对：

| 产物                    | 路径                                                               |
| ----------------------- | ------------------------------------------------------------------ |
| 插件 package 内容       | `<plugin_home>/`；archive 内部相对路径保持不变                     |
| 配置根目录              | `<plugin_home>/etc/`                                               |
| 主配置文件              | `<plugin_home>/<config_template.file_path>/<config_template.name>` |
| 运行目录                | `<plugin_home>/run/`                                               |
| PID 文件                | `<plugin_home>/run/<plugin_pkg_name>.pid`                          |
| 数据目录                | `<plugin_home>/data/`                                              |
| 默认日志目录（Unix）    | `/var/log/<deploy_env>/plugin/`                                    |
| 默认日志目录（Windows） | `C:\<deploy_env>\logs\plugin\`                                     |

进程名来自 `plugin_pkg_name`；Windows 会追加 `.exe`。安装器会把 release package 解压到 `<plugin_home>/`，并把渲染后的配置复制到 `<plugin_home>/etc/`。

## 重复行为

文档化的 mode 语义基于 desired state：如果插件已经以请求版本存在，则 desired state 已满足。

公开 contract 不定义 idempotency key、retry window、rollback rule，也不定义重复 API 调用的时序保证。

## 失败情况与限制

- 请求校验要求提供 `plugin_name`。
- 请求校验要求提供 `version`。
- 无效 `scope` 会阻止目标解析。
- 成功的 `execute` 响应只表示执行任务已发起。
- 机器侧验收应检查上述目录、请求版本和目标进程状态，不能只检查 `workflow_id`。

## Contract 参考

- [接入总览](README.md)
- [Spec 概念](../../concepts/deploy_policy/spec.md)
- [Scope 概念](../../concepts/deploy_policy/scope.md)
- [Swagger contract](../../api/swagger/backend/api/v3/deploy_policy.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/deploy_policy.proto)
