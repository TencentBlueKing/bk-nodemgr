# project_plugin_pkg_to_hosts

See [Workflow usage](workflow.md) for execution result polling.

## 目的与适用场景

当第三方平台希望基于插件包为一组 remote targets 生成插件实例，并把这些插件实例部署到指定承载主机时，使用 `project_plugin_pkg_to_hosts`。

该 mode 把 `scope` 解析出的目标作为 remote target，用于派生插件 identity；插件进程实际运行在 `placement_host_ids` 指定的承载主机上。

## 输入

该 mode 的 desired-state 字段：

| Field                   | Required | 含义                               |
| ----------------------- | -------- | ---------------------------------- |
| `plugin_pkg_name`       | yes      | 作为安装来源的插件包名称           |
| `version`               | yes      | 需要确保的插件包版本               |
| `placement_host_ids`    | yes      | 插件进程实际运行的承载主机 ID 列表 |
| `custom_config_context` | no       | 以结构化对象传入的自定义值         |

`placement_host_ids` 是承载位置，不是第二套 `scope`。

## 最小 payload 和 curl template

示例使用 `instance` scope 和 `service_instance` granularity。需要 `curl` 和 `jq`。先设置当前部署的 API base URL、remote target 与承载主机。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_BIZ_ID=2
export BK_SERVICE_INSTANCE_ID=30001
export PLACEMENT_HOST_ID=10001

CREATE_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/create" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
  {
    "name": "project-plugin-pkg-to-hosts-example",
    "description": "Project generated plugin package instances to placement hosts",
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
        "type": "project_plugin_pkg_to_hosts",
        "param": {
          "plugin_pkg_name": "example_plugin_pkg",
          "version": "1.0.0",
          "placement_host_ids": [${PLACEMENT_HOST_ID}],
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

平台会把该 spec 解释为 `scope remote targets × placement_host_ids` 的插件包 desired state。

`scope` 解析出的 remote target 用于生成插件 identity；`placement_host_ids` 决定生成的插件进程放在哪些主机上运行。

插件名称按 remote target identity 派生：

```text
<plugin_name> = <plugin_pkg_name>_<deploy_policy_id>_<remote_module_id>_<remote_host_id>
```

当 remote target 是 host granularity 时，`remote_module_id` 使用 `0`。

## 即时输出

create 返回 `data.deploy_policy_id`，用于标识已创建策略。

execute 返回 `data.workflow_id`，用于标识已发起的执行任务。

这些即时输出不是承载主机上插件实例已经安装、升级或卸载完成的最终证明。

## 最终或机器侧可见产物

目标插件的目录公式为：

```text
<plugin_home> = <base_deploy_dir>/<deploy_env>/plugin/<plugin_group>/<plugin_name>
```

该 mode 的 `plugin_group` 为十进制 `<deploy_policy_id>`；`plugin_name` 使用 remote target identity 生成。机器侧应在每个 `placement_host_ids` 对应主机上核对生成插件实例。

Windows 使用 `\` 作为路径分隔符。`<base_deploy_dir>` 来自承载主机的 plugin deployment 配置，可被 Network Unit 的 custom deploy config 覆盖；`<deploy_env>` 来自当前部署环境。

## 重复行为

该 mode 的语义基于 desired state：期望集合由 `scope remote targets × placement_host_ids` 计算得到。

概念文档定义：期望存在但实际不存在时安装，版本不匹配时升级，实际存在但不再属于期望集合时卸载。

公开 contract 不定义 idempotency key、retry window、rollback rule，也不定义重复 API 调用的时序保证。

## 失败情况与限制

- 请求校验要求提供 `plugin_pkg_name`。
- 请求校验要求提供 `version`。
- 请求校验要求 `placement_host_ids` 非空，且每个 host ID 必须大于 `0`。
- `placement_host_ids` 只表示承载主机，不表达调度、分片、主备角色或按主机差异化配置。
- 无效 `scope` 会阻止 remote target 解析。
- 成功的 `execute` 响应只表示执行任务已发起。
- 如果某个部署版本的 execute 返回 unsupported spec type，表示该部署尚未启用该 mode 的执行收敛能力。

## Contract 参考

- [接入总览](README.md)
- [Spec 概念](../../concepts/deploy_policy/spec.md)
- [Scope 概念](../../concepts/deploy_policy/scope.md)
- [Swagger contract](../../api/swagger/backend/api/v3/deploy_policy.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/deploy_policy.proto)
