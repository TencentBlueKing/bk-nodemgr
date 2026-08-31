# deploy_policy 接入指南

本文面向需要通过 bk-nodemgr backend API 创建和执行 `deploy_policy` 的第三方平台开发者，说明接入流程、调用方需要提交的输入、API 即时输出，以及策略期望产生的机器侧可见效果。

endpoint 和响应 envelope 以 [deploy_policy.swagger.json](../../api/swagger/backend/api/v3/deploy_policy.swagger.json) 为准；各 mode 的 `param` 与 `scope` 字段以 [deploy_policy.proto](../../../proto/backend/api/v3/deploy_policy.proto) 为准。领域语义见 [scope](../../concepts/deploy_policy/scope.md) 与 [spec](../../concepts/deploy_policy/spec.md)。

## 接入结果

`deploy_policy` 用来描述一组目标的 desired state。第三方平台提交：

- `scope`：哪些机器或 service instance 受影响。
- `spec`：这些目标最终应达到的插件状态。
- `enabled`：该策略是否允许执行。

API 响应只确认策略已创建，或执行任务已发起；响应本身不证明所有目标已经达到 desired state。

## API-first 快速接入

curl template 按 mode 拆分：

| 目标                                               | 阅读                                                                                  |
| -------------------------------------------------- | ------------------------------------------------------------------------------------- |
| 确保指定名称和版本的插件存在                       | [specify_plugin](specify_plugin.md)                                                   |
| 基于插件包生成并确保插件实例存在                   | [specify_plugin_pkg](specify_plugin_pkg.md)                                           |
| 声明已安装插件的配置内容                           | [specify_plugin_sub_config](specify_plugin_sub_config.md)                             |
| 基于 scope source targets 生成已安装插件的配置文件 | [project_plugin_config_template_to_hosts](project_plugin_config_template_to_hosts.md) |

已文档化 mode 的创建流程一致：

1. 根据目标选择构造 `scopes`。
2. 根据 desired state 构造 `specs`。
3. 调用 `POST /api/v3/deploy_policy/create`。
4. 从 create 响应保存 `data.deploy_policy_id`。
5. 用该 `deploy_policy_id` 调用 `POST /api/v3/deploy_policy/execute`，让目标收敛到声明的插件或配置状态。
6. 从 execute 响应保存 `data.trigger_id`，作为执行任务标识。

## 接入流程

### 1. 确认 API 上下文

公开的 deploy-policy endpoints：

| 步骤     | Method and path                      | 即时输出                   |
| -------- | ------------------------------------ | -------------------------- |
| 创建策略 | `POST /api/v3/deploy_policy/create`  | `data.deploy_policy_id`    |
| 执行策略 | `POST /api/v3/deploy_policy/execute` | `data.trigger_id`          |
| 查询策略 | `POST /api/v3/deploy_policy/list`    | `data.total`, `data.items` |
| 更新策略 | `POST /api/v3/deploy_policy/update`  | 响应结构见 Swagger         |

curl template 中的 `BK_NODEMGR_API_BASE` 由调用方提供，表示当前部署的 API base URL；deploy-policy contract 不定义统一 gateway 或认证 header。

### 2. 把目标转换成 scope

`scope` 告诉 bk-nodemgr 如何定位目标机器或 service instance。支持的目标结果类型见 [scope](../../concepts/deploy_policy/scope.md)：

| Scope type         | 可产生 host 目标 | 可产生 service_instance 目标 |
| ------------------ | ---------------- | ---------------------------- |
| `topo`             | yes              | yes                          |
| `set_template`     | yes              | yes                          |
| `service_template` | yes              | yes                          |
| `instance`         | yes              | yes                          |
| `dynamic_group`    | yes              | no                           |

当策略直接作用于机器时使用 `host` granularity；当策略必须区分模块或 service instance 位置时使用 `service_instance` granularity。

### 3. 声明 desired state

`spec` 定义每个 `scope` 选中目标的 desired final state。公开概念文档把 `spec` 定义为目标应达到的期望状态。

#### specify_plugin

当平台希望目标节点拥有指定 `plugin_name` 与 `version` 的插件时，使用 `specify_plugin`。文档化行为是：插件不存在则安装，版本不匹配则升级。

详情和 curl template：[specify_plugin](specify_plugin.md)。

#### specify_plugin_pkg

当平台提供插件包名称和版本时，使用 `specify_plugin_pkg`。文档化行为是：目标节点应拥有基于指定插件包和版本安装出的插件。插件名称由 deploy policy ID 和 module ID 生成。

详情和 curl template：[specify_plugin_pkg](specify_plugin_pkg.md)。

#### specify_plugin_sub_config

`specify_plugin_sub_config` 只用于声明已安装插件的配置。文档化行为是 config-only：更新插件配置文件内容，不安装或升级插件版本。

执行策略后，系统会渲染并下发声明的非主配置文件，使目标插件最终拥有对应 config。目标插件必须已经安装。

详情和 curl template：[specify_plugin_sub_config](specify_plugin_sub_config.md)。

#### project_plugin_config_template_to_hosts

`project_plugin_config_template_to_hosts` 用于把 scope 解析出的 source targets 投射为某个已安装插件的一组配置文件。文档化行为是 config-only：每个 source target 生成一个 sub config file，不安装或升级插件。

该 mode 不接收 `placement_host_ids`。系统通过 `plugin_name` 反向定位已有插件进程所在主机，并把生成配置声明到这些插件上。

详情和 curl template：[project_plugin_config_template_to_hosts](project_plugin_config_template_to_hosts.md)。

### 4. 提交策略

`POST /api/v3/deploy_policy/create` 接收：

- `name`
- `description`
- `enabled`
- `specs`
- `scopes`

create 响应包含 `data.deploy_policy_id`。把它保存为策略 identity，用于后续执行或更新。

### 5. 解读即时输出

`POST /api/v3/deploy_policy/execute` 接收 `deploy_policy_id`，返回 `data.trigger_id`。

`trigger_id` 表示平台已发起执行任务。它不等于最终插件健康状态、最终进程状态，也不保证机器已经收敛。

### 6. 观察最终产物

当前实现按以下公式生成插件目录：

```text
<plugin_home> = <base_deploy_dir>/<deploy_env>/plugin/<plugin_group>/<plugin_name>
```

Windows 使用 `\` 作为路径分隔符。`<base_deploy_dir>` 来自目标节点的 plugin deployment 配置，可被 Network Unit 的 custom deploy config 覆盖；`<deploy_env>` 来自当前部署环境。

机器侧可见效果取决于 spec mode：

| Mode                                      | `plugin_group`         | `plugin_name`                                      | 机器侧产物                                                                                                                             |
| ----------------------------------------- | ---------------------- | -------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `specify_plugin`                          | 插件注册信息中的 group | 请求中的 `plugin_name`                             | package 内容解压到 `<plugin_home>/`；配置目录为 `<plugin_home>/etc/`；PID 文件为 `<plugin_home>/run/<plugin_pkg_name>.pid`             |
| `specify_plugin_pkg`                      | `<deploy_policy_id>`   | `<plugin_pkg_name>_<deploy_policy_id>_<module_id>` | package 内容解压到 `<base_deploy_dir>/<deploy_env>/plugin/<deploy_policy_id>/<plugin_name>/`；配置与 PID 文件位于该目录下              |
| `specify_plugin_sub_config`               | 已安装插件的 group     | 请求中的 `plugin_name`                             | `is_main_config=false` 的文件写入 `<plugin_home>/etc/<plugin_name>/<name>`，随后 reload 插件                                           |
| `project_plugin_config_template_to_hosts` | 已安装插件的 group     | 请求中的 `plugin_name`                             | 每个 source target 生成一个 sub config file；配置名为 `<base_name>_deploy_<deploy_policy_id>_<source_module_id>_<source_host_id><ext>` |

两个安装类 mode 的 package archive 都直接解压到对应 `<plugin_home>/`。archive 内部相对路径会被保留；主配置文件路径由插件包的 `config_template.file_path` 和 `config_template.name` 决定：

```text
<plugin_home>/<config_template.file_path>/<config_template.name>
```

### 7. 重复或修订声明

概念文档描述的是 desired final state，不描述请求去重或 retry key。除非你的部署暴露了单独 contract，否则不要假设通用 idempotency key、retry safety、replacement behavior 或 rollback behavior。

只有现有概念文档支撑时，mode 页面才说明该 mode 的 repeat behavior。

### 8. 处理失败与边界

以下内容属于接入边界：

- 请求结构错误或 unsupported fields 是 API contract 问题，检查 Swagger reference。
- Unsupported `scope` 和目标组合是概念问题，检查 [scope](../../concepts/deploy_policy/scope.md)。
- Unsupported `spec` 语义是 desired-state 问题，检查 [spec](../../concepts/deploy_policy/spec.md)。
- create 或 execute 成功响应是即时 API 结果，不是最终机器状态验证。
- `specify_plugin_sub_config` 是 config-only，要求插件已安装；它只保证声明的非主配置文件存在，不安装或升级插件。
- `project_plugin_config_template_to_hosts` 是 config-only，要求 `plugin_name` 能反向定位到已有插件；它不接收 `placement_host_ids`，也不安装或升级插件。
- `specify_plugin_pkg_sub_config` 出现在概念文档中，但当前 proto、Swagger 和 type contract 均没有该字段；不要把它作为 spec type 发送。

## Contract 参考

- [Scope 概念](../../concepts/deploy_policy/scope.md)
- [Spec 概念](../../concepts/deploy_policy/spec.md)
- [Swagger contract](../../api/swagger/backend/api/v3/deploy_policy.swagger.json)
- [Proto variant 定义](../../../proto/backend/api/v3/deploy_policy.proto)
- [specify_plugin](specify_plugin.md)
- [specify_plugin_pkg](specify_plugin_pkg.md)
- [specify_plugin_sub_config](specify_plugin_sub_config.md)
- [project_plugin_config_template_to_hosts](project_plugin_config_template_to_hosts.md)
