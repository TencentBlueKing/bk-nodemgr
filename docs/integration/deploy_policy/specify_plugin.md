# specify_plugin

## 目的与适用场景

当第三方平台希望目标节点拥有指定 `plugin_name` 和 `version` 的插件时，使用 `specify_plugin`。

现有概念文档对该 mode 的定义是：确保目标节点有指定插件名称和版本；如果插件不存在，则安装；如果版本不匹配，则升级。

## 输入

该 mode 的 desired-state 字段：

| Field | Required | 含义 |
| --- | --- | --- |
| `plugin_name` | yes | 目标节点上需要确保存在的插件名称 |
| `version` | yes | 需要确保的插件版本 |
| `custom_config_context` | no | 以结构化对象传入的自定义值 |

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

使用返回的 ID 执行策略，并提取 `data.trigger_id`：

```bash
EXECUTE_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/execute" \
  -H "Content-Type: application/json" \
  -d "{\"deploy_policy_id\": ${DEPLOY_POLICY_ID}}")"

TRIGGER_ID="$(printf '%s' "${EXECUTE_RESPONSE}" | jq -r '.data.trigger_id')"
```

## 系统解释

平台会把该 spec 解释为插件 desired state：`scopes` 选中的每个目标都应拥有 `plugin_name`，且版本为 `version`。

根据概念文档，插件缺失会触发安装，版本不匹配会触发升级。公开 contract 不描述具体包查找、传输、重启或 health-check 步骤。

## 即时输出

create 返回 `data.deploy_policy_id`，用于标识已创建策略。

execute 返回 `data.trigger_id`，用于标识已发起的执行任务。

这些即时输出不是最终机器状态证明。

## 最终或机器侧可见产物

期望的最终产物是：目标节点拥有指定名称的插件，且版本为请求中的版本。

deploy-policy API contract 不定义公开的机器路径、进程名称、reload 行为或 health-check 命令。除非你的部署暴露了其他文档化 contract，否则不要让接入逻辑依赖这些细节。

## 重复行为

文档化的 mode 语义基于 desired state：如果插件已经以请求版本存在，则 desired state 已满足。

公开 contract 不定义 idempotency key、retry window、rollback rule，也不定义重复 API 调用的时序保证。

## 失败情况与限制

- 请求校验要求提供 `plugin_name`。
- 请求校验要求提供 `version`。
- 无效 `scope` 会阻止目标解析。
- 成功的 `execute` 响应只表示执行任务已发起。
- 包来源、目标可达性和运行时健康状态不属于 schema-level 响应 contract。

## Contract 参考

- [接入总览](README.md)
- [Spec 概念](../../concepts/deploy_policy/spec.md)
- [Scope 概念](../../concepts/deploy_policy/scope.md)
- [Swagger contract](../../api/swagger/backend/api/v3/deploy_policy.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/deploy_policy.proto)
