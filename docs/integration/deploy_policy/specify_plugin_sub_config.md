# specify_plugin_sub_config

See [Workflow usage](workflow.md) for parent dispatch status and child workflow inspection.

## 目的与适用场景

当第三方平台希望声明已安装插件的配置时，使用 `specify_plugin_sub_config`。

现有概念文档对该 mode 的定义是：更新已安装插件的配置文件内容。它只更新配置，不安装或升级插件版本。

## 输入

该 mode 的 desired-state 字段：

| Field                   | Required | 含义                                     |
| ----------------------- | -------- | ---------------------------------------- |
| `plugin_name`           | yes      | 被声明配置的已安装插件                   |
| `config_files_detail`   | no       | 要更新的配置文件详情；声明文件内容时提供 |
| `custom_config_context` | no       | 以结构化对象传入的自定义值               |

每个 `config_files_detail` item 遵循 proto 文档化结构：

| Field            | 含义                                                   |
| ---------------- | ------------------------------------------------------ |
| `name`           | 配置文件名称                                           |
| `content`        | 配置文件内容                                           |
| `is_main_config` | 该项是否为主配置；该 mode 使用 `false` 声明 sub config |

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
    "name": "apply-plugin-config-example",
    "description": "Declare configuration for an installed plugin",
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
        "type": "specify_plugin_sub_config",
        "param": {
          "plugin_name": "example_plugin",
          "config_files_detail": [
            {
              "name": "example.conf",
              "content": "key=value",
              "is_main_config": false
            }
          ],
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

平台会把该 spec 解释为 `plugin_name` 的 config-only desired state。

它不声明插件安装、插件包选择或插件版本升级。如果插件可能不存在，应先使用 `specify_plugin` 或其他文档化插件安装流程。

## 即时输出

create 返回 `data.deploy_policy_id`，用于标识已创建策略。

execute 返回 `data.workflow_id`，用于标识已发起的配置收敛任务。

`workflow_id` 不证明配置文件已经写入、插件已经 reload 或最终状态已经收敛。机器侧验收应检查目标配置文件及插件状态。

## 最终或机器侧可见产物

目标插件的目录公式为：

```text
<plugin_home> = <base_deploy_dir>/<deploy_env>/plugin/<plugin_group>/<plugin_name>
```

对于 `config_files_detail` 中 `is_main_config=false` 的每个 item，系统保证最终存在对应配置文件：

```text
<plugin_home>/etc/<plugin_name>/<name>
```

本页示例最终生成：

```text
<plugin_home>/etc/example_plugin/example.conf
```

Windows 使用 `\` 作为路径分隔符。配置文件写入后，系统 reload 目标插件，使声明的 config 生效。

## 重复行为

概念文档只定义该 mode 为已安装插件的配置更新。

公开 contract 不定义重复声明会 merge、replace、patch、no-op，是否 retry safe，或是否 rollback 配置。

## 失败情况与限制

- 请求校验要求提供 `plugin_name`。
- 请求校验不要求 `config_files_detail`，但声明文件内容时至少需要一个有意义的 item。
- 目标插件必须已安装；该 mode 不安装或升级插件。
- 该 mode 只处理 `is_main_config=false` 的 sub config；主配置应由插件安装或其他配置流程管理。
- 成功的 `execute` 响应只表示配置收敛任务已发起。

## Contract 参考

- [接入总览](README.md)
- [Spec 概念](../../concepts/deploy_policy/spec.md)
- [Scope 概念](../../concepts/deploy_policy/scope.md)
- [Swagger contract](../../api/swagger/backend/api/v3/deploy_policy.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/deploy_policy.proto)
