# project_plugin_config_template_to_hosts

## 目的与适用场景

当第三方平台希望把一批 scope source targets 投射为某个已安装插件的一组配置文件时，使用 `project_plugin_config_template_to_hosts`。

该 mode 是 config-only desired state：它只声明插件子配置文件，不安装插件、不升级插件，也不选择插件包或插件版本。配置承载主机由 `plugin_name` 反向定位到的已有插件进程决定。

## 输入

该 mode 的 desired-state 字段：

| Field                   | Required | 含义                                                            |
| ----------------------- | -------- | --------------------------------------------------------------- |
| `plugin_name`           | yes      | 被声明配置的已安装插件；同时用于反向定位配置承载主机            |
| `config_files_detail`   | yes      | 配置模板详情；第一版只允许一个 item，且必须指定 `template_name` |
| `custom_config_context` | no       | 以结构化对象传入的自定义值                                      |

每个 `config_files_detail` item 遵循 proto 文档化结构：

| Field            | 含义                                                   |
| ---------------- | ------------------------------------------------------ |
| `template_name`  | 配置模板名称；该 mode 使用它作为生成配置文件的模板来源 |
| `content`        | 可选配置内容；具体解释以插件配置模板 contract 为准     |
| `is_main_config` | 该项是否为主配置；该 mode 用于生成 sub config          |

该 mode 不接收 `placement_host_ids`。配置文件会声明到 `plugin_name` 反向定位出的插件所在主机上。

## 最小 payload 和 curl template

示例使用 `instance` scope 和 `service_instance` granularity。需要 `curl` 和 `jq`。先设置当前部署的 API base URL 与 source target 标识。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_BIZ_ID=2
export BK_SERVICE_INSTANCE_ID=30001

CREATE_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/create" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
  {
    "name": "project-plugin-config-template-example",
    "description": "Project source targets into plugin sub config files",
    "enabled": true,
    "scopes": [
      {
        "type": "instance",
        "scope": {
          "granularity": "service_instance",
          "bk_biz_id": ${BK_BIZ_ID},
          "instance_ids": [${BK_SERVICE_INSTANCE_ID}]
        }
      }
    ],
    "specs": [
      {
        "type": "project_plugin_config_template_to_hosts",
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

平台会把该 spec 解释为 project config template desired state：每个 scope 解析出的 source target 都应投射成指定 `plugin_name` 下、同一个配置模板派生出的一个 sub config file。

`plugin_name` 有两层含义：

| 含义     | 说明                                         |
| -------- | -------------------------------------------- |
| 配置归属 | 生成的 sub config file 归属于该插件          |
| 承载定位 | 系统通过该插件名反向定位已有插件进程所在主机 |

配置文件名称由系统根据配置模板名和 source target identity 生成：

```text
<config_name> = <base_name>_deploy_<deploy_policy_id>_<source_module_id>_<source_host_id><ext>
```

当 source target 是 host granularity 时，`source_module_id` 使用 `0`。

## 即时输出

create 返回 `data.deploy_policy_id`，用于标识已创建策略。

execute 返回 `data.trigger_id`，用于标识已发起的配置收敛任务。

`trigger_id` 不证明配置文件已经写入、插件已经 reload 或最终状态已经收敛。机器侧验收应检查目标配置文件及插件状态。

## 最终或机器侧可见产物

目标插件的目录公式为：

```text
<plugin_home> = <base_deploy_dir>/<deploy_env>/plugin/<plugin_group>/<plugin_name>
```

对每个 source target，系统生成一个非主配置文件。配置文件路径由目标插件的 sub config 路径规则和生成的 `<config_name>` 决定。

## 重复行为

该 mode 的语义基于 desired state：source target 集合或匹配 `plugin_name` 的插件进程集合变化后，系统应重新计算期望配置集合。

不再属于期望集合的 project config file 应删除。

## 失败情况与限制

- 请求校验要求提供 `plugin_name`。
- 请求校验要求 `config_files_detail` 只包含一个 item。
- 请求校验要求该 item 提供 `template_name`。
- 该 mode 不接收 `placement_host_ids`；不要把承载主机作为请求字段发送。
- 该 mode 不安装或升级插件；如果没有主机存在指定 `plugin_name`，本轮不会产生配置投射。
- source target 是 host granularity 时，生成名称中的 `source_module_id` 为 `0`。
- 成功的 `execute` 响应只表示执行任务已发起。

## Contract 参考

- [接入总览](README.md)
- [Spec 概念](../../concepts/deploy_policy/spec.md)
- [Scope 概念](../../concepts/deploy_policy/scope.md)
- [Swagger contract](../../api/swagger/backend/api/v3/deploy_policy.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/deploy_policy.proto)
