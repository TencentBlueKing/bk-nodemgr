# specify_plugin_sub_config

## 目的与适用场景

当第三方平台希望声明已安装插件的配置时，使用 `specify_plugin_sub_config`。

现有概念文档对该 mode 的定义是：更新已安装插件的配置文件内容。它只更新配置，不安装或升级插件版本。

## 输入

该 mode 的 desired-state 字段：

| Field | Required | 含义 |
| --- | --- | --- |
| `plugin_name` | yes | 被声明配置的已安装插件 |
| `config_files_detail` | no | 要更新的配置文件详情；声明文件内容时提供 |
| `custom_config_context` | no | 以结构化对象传入的自定义值 |

每个 `config_files_detail` item 遵循 proto 文档化结构：

| Field | 含义 |
| --- | --- |
| `name` | 配置文件名称 |
| `content` | 配置文件内容 |
| `is_main_config` | 该项是否为主配置 |

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
              "is_main_config": true
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

create 返回 `data.deploy_policy_id`，但当前版本不支持执行包含该 spec 的策略。保留该 ID 可用于策略查询或更新；不要把它发送到 execute endpoint 用于配置下发。

## 系统解释

平台会把该 spec 解释为 `plugin_name` 的 config-only desired state。

它不声明插件安装、插件包选择或插件版本升级。如果插件可能不存在，应先使用 `specify_plugin` 或其他文档化插件安装流程。

## 即时输出

create 返回 `data.deploy_policy_id`，用于标识已创建策略。

当前版本没有该 mode 的 supported execution result。如果仍然调用 execute endpoint，它可能先返回 `data.trigger_id`，随后任务在处理 unsupported spec 时失败。创建策略或收到该 ID，都不证明配置已下发到目标、写入插件、完成 reload 或生效。

## 最终或机器侧可见产物

当前 `deploy_policy` 执行流程不会把该声明物化为机器侧配置产物。

deploy-policy contract 也不定义文件路径、merge-vs-replace 行为、reload 行为、health check 或时序保证。

## 重复行为

概念文档只定义该 mode 为已安装插件的配置更新。

公开 contract 不定义重复声明会 merge、replace、patch、no-op，是否 retry safe，或是否 rollback 配置。

## 失败情况与限制

- 请求校验要求提供 `plugin_name`。
- 请求校验不要求 `config_files_detail`，但声明文件内容时至少需要一个有意义的 item。
- 目标插件必须已安装；该 mode 不安装或升级插件。
- 当前版本不支持通过 `deploy_policy` 执行该 spec。

## Contract 参考

- [接入总览](README.md)
- [Spec 概念](../../concepts/deploy_policy/spec.md)
- [Scope 概念](../../concepts/deploy_policy/scope.md)
- [Swagger contract](../../api/swagger/backend/api/v3/deploy_policy.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/deploy_policy.proto)
