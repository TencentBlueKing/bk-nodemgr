# specify_plugin_pkg

## 目的与适用场景

当第三方平台希望目标节点拥有基于指定插件包和版本安装出的插件时，使用 `specify_plugin_pkg`。

现有概念文档对该 mode 的定义是：确保目标节点有来自指定插件包名称和版本的插件。插件名称由 deploy policy ID 和 module ID 生成。如果插件不存在，则安装；如果版本不匹配，则升级。

## 输入

该 mode 的 desired-state 字段：

| Field | Required | 含义 |
| --- | --- | --- |
| `plugin_pkg_name` | yes | 作为安装来源的插件包名称 |
| `version` | yes | 需要确保的插件包版本 |
| `custom_config_context` | no | 以结构化对象传入的自定义值 |

当你的接入需要生成的插件 identity 体现 service/module 位置时，使用 `service_instance` scope。概念文档说明生成插件名称依赖 deploy policy ID 和 module ID。

## 最小 payload 和 curl template

示例使用 `instance` scope 和 `service_instance` granularity。需要 `curl` 和 `jq`。先设置当前部署的 API base URL 与目标标识。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_BIZ_ID=2
export BK_SERVICE_INSTANCE_ID=30001

CREATE_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/create" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
  {
    "name": "ensure-plugin-package-example",
    "description": "Ensure a generated plugin instance from a plugin package",
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
        "type": "specify_plugin_pkg",
        "param": {
          "plugin_pkg_name": "example_plugin_pkg",
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

使用返回的 ID 执行策略，并提取 `data.trigger_id`：

```bash
EXECUTE_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/execute" \
  -H "Content-Type: application/json" \
  -d "{\"deploy_policy_id\": ${DEPLOY_POLICY_ID}}")"

TRIGGER_ID="$(printf '%s' "${EXECUTE_RESPONSE}" | jq -r '.data.trigger_id')"
```

## 系统解释

平台会把该 spec 解释为插件包 desired state：每个解析出的目标都应拥有从 `plugin_pkg_name` 生成、版本为 `version` 的插件实例。

概念文档说明插件名称由 deploy policy ID 和 module ID 生成，但未定义第三方平台应自行实现的公开公式。

## 即时输出

create 返回 `data.deploy_policy_id`，用于标识已创建策略。

execute 返回 `data.trigger_id`，用于标识已发起的执行任务。

这些即时输出不是生成插件实例已经存在于每个目标上的最终证明。

## 最终或机器侧可见产物

期望的最终产物是：目标节点上存在基于请求插件包名称和版本生成的插件实例。

公开 deploy-policy contract 不定义机器路径、包缓存路径、生成文件名、reload 行为或 health check。

## 重复行为

文档化的 mode 语义基于 desired state：如果生成插件实例已经以请求的插件包版本存在，则 desired state 已满足。

公开 contract 不定义 idempotency key、retry window、未来版本中的生成名称稳定性、rollback rule，也不定义重复 API 调用的时序保证。

## 失败情况与限制

- 请求校验要求提供 `plugin_pkg_name`。
- 请求校验要求提供 `version`。
- 无效 `scope` 会阻止目标解析。
- 生成插件名称是平台拥有的结果；调用方不应推导或依赖未文档化公式。
- 成功的 `execute` 响应只表示执行任务已发起。

## Contract 参考

- [接入总览](README.md)
- [Spec 概念](../../concepts/deploy_policy/spec.md)
- [Scope 概念](../../concepts/deploy_policy/scope.md)
- [Swagger contract](../../api/swagger/backend/api/v3/deploy_policy.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/deploy_policy.proto)
