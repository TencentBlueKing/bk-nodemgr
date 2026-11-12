# Node Manager 集成方案

Node Manager 以独立同步镜像交付 Agent Package，在部署流程中调用平台提供的 `bkai-init`。本文描述接入职责和集成边界；以下 AIDEV 镜像名为建议名称，不表示仓库已提供对应 Dockerfile、Make target 或 Helm values。

## 镜像与资源交付

沿用现有 APIGW 的“平台工具 + 项目资源 + 部署任务”分工，不将同步逻辑放入 Backend，也不自行实现资源客户端。

| 环节     | APIGW 当前项目入口                                                                                   | AIDEV 接入方式                                                               |
| -------- | ---------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| 基础镜像 | [Dockerfile](../../../../install/images/bk-nodemgr-apigw-sync/Dockerfile) 使用 `apigw-manager:4.2.4` | 使用平台提供、内置 `bkai-init==0.1.6` 的基础镜像，固定并核对实际工具版本     |
| 项目输入 | `apigw/` 的资源、定义及接口文档                                                                      | Node Manager 自行维护的 `bkai.yaml`、Agent 配置和实际使用的依赖文件          |
| 派生镜像 | `bk-nodemgr-apigw-sync` 打包同步输入、脚本和 `adminclient`                                           | 独立镜像，建议命名为 `bk-nodemgr-aidev-sync`，将 Package 放入 `/app/package` |
| 构建入口 | 根 [Makefile](../../../../Makefile) 的 `docker-prepare-apigw-sync` / `docker-build-apigw-sync`       | 按同样的组装与构建职责纳入项目构建，不依赖开发者运行上游 Demo                |
| 部署任务 | [APIGW Job](../../../../install/helm/bk-nodemgr/templates/apigw-sync-job.yaml) 在安装、升级后执行    | 独立 AIDEV Job，运行时注入配置后调用 `bkai-init`                             |

派生镜像只加入项目资源和必要启动适配。平台基础镜像已含目标版本时，不重复安装 Python 包，不复制工具源码；基础镜像地址由平台提供，不使用 Demo 的本地镜像名作为部署地址。初始化工具版本、Node Manager 模块镜像版本、Package 版本、Skill 版本和 Agent 发布版本分别管理。

资源目录只包含实际需要交付的内容。内联 Prompt 不要求独立角色；没有 Skill、知识库或子智能体需求时，不为复制完整 Demo 创建这些资源。目录与最小清单见[Package 约束](agent-package.md)，构建期与运行期的区别见[Demo 机制](demo.md)。

## 配置与凭据注入

| 内容                                                                    | 交付位置                     | 更新影响                                       |
| ----------------------------------------------------------------------- | ---------------------------- | ---------------------------------------------- |
| Prompt、模型、指令、依赖关联、Skill 文件、知识库文档                    | 项目 Package，随模块镜像交付 | 文件变化需重新构建模块镜像                     |
| 网关地址、目标租户、完整空间 ID、执行用户、部署变量、资源排除和发布参数 | 部署任务运行配置             | 配置变化不要求重建资源镜像，须重新执行相应任务 |
| 应用密钥及其它真实凭据                                                  | 部署现有凭据注入机制         | 不进入资源文件、镜像、Git 或日志               |

项目 APIGW/IAM 通过 [Chart helper](../../../../install/helm/bk-nodemgr/templates/_helpers.tpl) 的 `bkappResolveScript` 和 `bk-nodemgr-adminclient bkapp` 解析应用身份，支持现有明文来源及 KMS 路径。AIDEV 应沿用该凭据管理边界，而不是另建一套密钥配置。

接入适配需要将解析结果传给 `BKAI_APP_CODE` / `BKAI_APP_SECRET`。`bkai-init` 的应用变量名不同于 APIGW 的 `BK_APP_CODE` / `BK_APP_SECRET`，不能假设直接注入原变量就会生效。现有 helper 的消费方与配置生成条件只覆盖已有同步任务，不能仅添加 Job 就视为已支持 AIDEV 凭据解析。

多租户时向工具传入应用实际租户与目标完整空间 ID。不要照搬 APIGW Job 的租户推导逻辑，也不要将默认 `system`、`bkaidev` 当作目标环境配置。执行者和追加管理员使用实际 `bk_username`；身份转换及 CLI 参数见[运行参数](execution.md#连接与运行参数)。

## 与 APIGW/MCP 的依赖关系

执行顺序为：**Node Manager API 可用 → APIGW 资源与目标环境发布 → MCP 创建、发布及所需授权 → AIDEV Package 校验和预览 → 同步并按需提交 Agent 发布 → 后台结果与实际调用验收。**

项目已有 `topo` MCP 声明，定位是拓扑、网络、接入点与主机信息的只读查询，不据此扩大为创建、修改或删除能力。接入 Agent 时使用网关返回的完整 MCP Server code，不将声明中的短名 `topo` 当作完整编码，也不猜测工具名称。

当前入口及其边界：

- [同步脚本](../../../../install/images/bk-nodemgr-apigw-sync/support-files/bin/sync-apigateway.sh) 仅在 `ENABLE_SYNC_MCP_SERVERS=true` 时同步 MCP；[Chart 默认配置](../../../../install/helm/bk-nodemgr/values.yaml) 关闭该开关，且 `apigwSync.enabled` 默认关闭。
- [仓库定义](../../../../apigw/definition.yaml) 与 Chart 中的 MCP 声明均包含 `status: 0`，不能由“有声明”认定目标 MCP 已发布。APIGW 的环境发布还受 `RELEASE_STAGES` 控制，默认配置不保证已经发布。
- Helm 用 [ConfigMap](../../../../install/helm/bk-nodemgr/templates/apigw-sync-configmap.yaml) 渲染的定义覆盖镜像内 `/data/definition.yaml`。仓库定义与 Chart 的 MCP `resource_names` 当前不同，接入时应核对实际渲染输入及其与发布资源名称的对应，不能只检查 `apigw/definition.yaml`。
- 当前 `grant_permissions` 声明是网关资源授权，不是 MCP 主动授权配置。MCP 创建、发布、应用与目标 Agent 所需授权须分别落实；`bkai-init` 不代办审批或授权。完整引用规则见 [MCP](resources.md#apigw-mcp)。

首次初始化 Agent 尚不存在时，目标平台需支持按应用空间权限查询 MCP 的初始化规则；不要提前创建空壳 Agent 绕过校验。Agent 已存在时，其空间仍须与本次目标空间一致。

## 部署任务与发布

AIDEV Job 参考 APIGW 的 `post-install` / `post-upgrade` 生命周期，并在所需 APIGW/MCP 步骤成功后执行。现有 APIGW hook weight 为 `5`；AIDEV 的执行顺序应显式表达依赖，不与其无序并发，也不因 APIGW 任务被关闭而假定依赖已就绪。

Helm 对 hook Job 等待的是 Job 完成，不是外部平台异步处理完成，见 [Helm hooks](https://helm.sh/docs/topics/charts_hooks/)。`bkai-init` 提交 Agent 发布或知识库导入后即可继续，因此需区分部署任务退出和平台业务完成。

- 写入命令必须明确 `sync --confirm`、清单路径和目标空间。人工预览先执行 `validate`、`show`、`plan`，不以 `plan` 成功代替写权限校验。
- 仅同步草稿时不加 `--publish`；需要常规发布时显式传 `--publish --publish_config_only=0`。首次包含主子 Agent 时，整包发布按先子后主提交，单独同步主 Agent 则要求其它来源已有可用发布版本。
- 发布策略不能直接继承 Demo 的 `make sync` 默认值，也不能将 APIGW 的发布开关视为 Agent 发布开关。
- 不照搬 APIGW 的 Job 重试次数。工具不自动重试写入，已成功资源不回滚；Kubernetes 重跑整个 Job 仍会再次执行同步，重跑前需核对线上状态和覆盖范围。

## 升级保护与验收

仓库维护的资源应提供完整配置。平台手工维护的 Agent 应按整个资源排除；只排除其 Skill 或角色不能保护 Agent 中的关联、环境变量或角色快照。知识库有文档时以完整本地目录为准，会删除线上多余内容。细则见[更新覆盖保护](execution.md#更新覆盖保护)。

| 验收层级   | 核对内容                                                                                                          |
| ---------- | ----------------------------------------------------------------------------------------------------------------- |
| 交付与配置 | 镜像携带正确 Package 和工具版本；实际注入的地址、租户、空间、身份、选择/排除列表及发布参数正确，凭据未泄露        |
| 资源同步   | 核对资源处理结果及 `show` / `plan` 的可回读字段；无法回读的 Skill 文件、envs 值、角色标签和知识库文档不能判为一致 |
| 后台任务   | 按平台发布状态、知识库任务 ID 确认结果；失败、部分成功或未知不算完成。无查询入口时将任务 ID 提供给平台维护方核实  |
| 实际运行   | 检查对话与 MCP 查询，以及实际引入的 Skill、子 Agent 或知识库检索；验证授权拒绝与租户、空间边界未被绕过            |

同步失败保留资源、步骤、执行时间、工具版本和 Trace ID/服务端请求标识，先查线上已写入状态再决定重跑。已发布配置、权限和知识库文档不因镜像或 Helm 回退自动恢复；不能把 Helm rollback 当作 AIDEV 资源回滚。
