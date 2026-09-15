# specify_plugin_pkg

See [Workflow usage](workflow.md) for parent dispatch status and child workflow inspection.

## 目的与适用场景

当第三方平台希望目标节点拥有基于指定插件包和版本安装出的插件时，使用 `specify_plugin_pkg`。

现有概念文档对该 mode 的定义是：确保目标节点有来自指定插件包名称和版本的插件。插件名称由 deploy policy ID 和 module ID 生成。如果插件不存在，则安装；如果版本不匹配，则升级。

## 输入

该 mode 的 desired-state 字段：

| Field                   | Required | 含义                       |
| ----------------------- | -------- | -------------------------- |
| `plugin_pkg_name`       | yes      | 作为安装来源的插件包名称   |
| `version`               | yes      | 需要确保的插件包版本       |
| `custom_config_context` | no       | 以结构化对象传入的自定义值 |

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

使用返回的 ID 执行策略，并提取 `data.workflow_id`：

```bash
EXECUTE_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/execute" \
  -H "Content-Type: application/json" \
  -d "{\"deploy_policy_id\": ${DEPLOY_POLICY_ID}}")"

WORKFLOW_ID="$(printf '%s' "${EXECUTE_RESPONSE}" | jq -r '.data.workflow_id')"
```

## 系统解释

平台会把该 spec 解释为插件包 desired state：每个解析出的目标都应拥有从 `plugin_pkg_name` 生成、版本为 `version` 的插件实例。

当前实现使用以下公式生成插件名称：

```text
<plugin_name> = <plugin_pkg_name>_<deploy_policy_id>_<module_id>
```

插件 group 为十进制 `<deploy_policy_id>`。同一个 deploy policy 在不同 module 下会生成不同 `plugin_name`。

## 即时输出

create 返回 `data.deploy_policy_id`，用于标识已创建策略。

execute 返回 `data.workflow_id`，用于标识已发起的执行任务。

这些即时输出不是生成插件实例已经存在于每个目标上的最终证明。

## 最终或机器侧可见产物

将生成名称和 group 代入 plugin home 公式后，目标目录是：

```text
<base_deploy_dir>/<deploy_env>/plugin/<deploy_policy_id>/<plugin_pkg_name>_<deploy_policy_id>_<module_id>/
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

进程名为 `plugin_pkg_name`；Windows 会追加 `.exe`。安装器会把 release package 解压到 `<plugin_home>/`，并把渲染后的配置复制到 `<plugin_home>/etc/`。

## 重复行为

文档化的 mode 语义基于 desired state：如果生成插件实例已经以请求的插件包版本存在，则 desired state 已满足。

公开 contract 不定义 idempotency key、retry window、未来版本中的生成名称稳定性、rollback rule，也不定义重复 API 调用的时序保证。

## 失败情况与限制

- 请求校验要求提供 `plugin_pkg_name`。
- 请求校验要求提供 `version`。
- 无效 `scope` 会阻止目标解析。
- 生成 `plugin_name` 需要 `module_id`；因此该 mode 应使用能够产生 `service_instance` target 的 scope。
- 成功的 `execute` 响应只表示执行任务已发起。
- 机器侧验收应检查上述生成目录、请求版本和目标进程状态，不能只检查 `workflow_id`。

## Contract 参考

- [接入总览](README.md)
- [Spec 概念](../../concepts/deploy_policy/spec.md)
- [Scope 概念](../../concepts/deploy_policy/scope.md)
- [Swagger contract](../../api/swagger/backend/api/v3/deploy_policy.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/deploy_policy.proto)
