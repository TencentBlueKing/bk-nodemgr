# specify_plugin_sub_config_template

## 目的与适用场景

当第三方平台希望按配置模板为已安装插件声明 deploy-policy 管理的 sub config file 时，使用 `specify_plugin_sub_config_template`。

该 mode 是 config-only desired state：它只声明插件子配置文件，不安装插件、不升级插件，也不选择插件包或插件版本。目标插件必须已经安装并处于可匹配状态。

## 输入

该 mode 的 desired-state 字段：

| Field                   | Required | 含义                             |
| ----------------------- | -------- | -------------------------------- |
| `plugin_name`           | yes      | 被声明配置的已安装插件           |
| `config_files_detail`   | no       | 配置模板详情；声明文件内容时提供 |
| `custom_config_context` | no       | 以结构化对象传入的自定义值       |

每个 `config_files_detail` item 遵循 proto 文档化结构：

| Field            | 含义                                                      |
| ---------------- | --------------------------------------------------------- |
| `template_name`  | 配置模板名称；优先用于生成 deploy-policy 管理的配置文件名 |
| `name`           | 配置文件名称；当 `template_name` 为空时作为生成名称来源   |
| `content`        | 可选配置内容；具体解释以插件配置模板 contract 为准        |
| `is_main_config` | 该项是否为主配置；该 mode 用于生成 sub config             |

有效的 template item 至少需要提供 `template_name` 或 `name`。如果两者都为空，该 item 不会产生配置声明。

## 最小 payload 和 curl template

示例使用 `instance` scope 和 `host` granularity。需要 `curl` 和 `jq`。先设置当前部署的 API base URL 与目标主机。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_BIZ_ID=2
export BK_HOST_ID=10001

CREATE_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/create" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
  {
    "name": "apply-plugin-config-template-example",
    "description": "Declare generated sub config for an installed plugin",
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
        "type": "specify_plugin_sub_config_template",
        "param": {
          "plugin_name": "example_plugin",
          "config_files_detail": [
            {
              "template_name": "target.conf",
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

使用返回的 ID 执行策略，并提取 `data.trigger_id`：

```bash
EXECUTE_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/execute" \
  -H "Content-Type: application/json" \
  -d "{\"deploy_policy_id\": ${DEPLOY_POLICY_ID}}")"

TRIGGER_ID="$(printf '%s' "${EXECUTE_RESPONSE}" | jq -r '.data.trigger_id')"
```

## 系统解释

平台会把该 spec 解释为 `plugin_name` 的 template-based sub config desired state。

系统先根据 `template_name` 或 `name` 生成 deploy-policy 管理的配置文件名称：

```text
<config_name> = <base_name>_deploy_<deploy_policy_id><ext>
```

然后只对已存在并可匹配到 `plugin_name` 的目标插件声明缺失的 sub config file。该 mode 不声明插件安装、插件包选择或插件版本升级。

## 即时输出

create 返回 `data.deploy_policy_id`，用于标识已创建策略。

execute 返回 `data.trigger_id`，用于标识已发起的配置收敛任务。

`trigger_id` 不证明配置文件已经写入、插件已经 reload 或最终状态已经收敛。机器侧验收应检查目标配置文件及插件状态。

## 最终或机器侧可见产物

目标插件的目录公式为：

```text
<plugin_home> = <base_deploy_dir>/<deploy_env>/plugin/<plugin_group>/<plugin_name>
```

对每个有效 template item，系统生成一个非主配置文件。配置文件路径由目标插件的 sub config 路径规则和生成的 `<config_name>` 决定。

本页示例最终生成的配置名称为：

```text
target_deploy_<deploy_policy_id>.conf
```

Windows 使用 `\` 作为路径分隔符。配置文件写入后，系统 reload 目标插件，使声明的 config 生效。

## 重复行为

该 mode 的语义基于 desired state：如果 deploy-policy 管理的目标 sub config file 已经存在，则本轮不重复创建同名配置。

公开 contract 不定义 idempotency key、retry window、rollback rule，也不定义重复 API 调用的时序保证。

## 失败情况与限制

- 请求校验要求提供 `plugin_name`。
- 请求校验不要求 `config_files_detail`，但声明文件内容时至少需要一个有意义的 item。
- 有效 item 至少需要 `template_name` 或 `name`；两者都为空时不会生成配置声明。
- 目标插件必须已安装；该 mode 不安装或升级插件。
- 该 mode 与 `specify_plugin_sub_config` 互斥，不应在同一 deploy policy 中同时声明同一类配置 desired state。
- 成功的 `execute` 响应只表示配置收敛任务已发起。

## Contract 参考

- [接入总览](README.md)
- [Spec 概念](../../concepts/deploy_policy/spec.md)
- [Scope 概念](../../concepts/deploy_policy/scope.md)
- [Swagger contract](../../api/swagger/backend/api/v3/deploy_policy.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/deploy_policy.proto)
