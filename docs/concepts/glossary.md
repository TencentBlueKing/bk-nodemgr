# Glossary

本文定义 bk-nodemgr 中跨文档、API、Proto、frontend type 和代码共享的 canonical terms。

Glossary 的目标不是翻译所有词，而是减少同一概念多种叫法、相近概念混用、历史名称继续扩散的问题。概念完整模型仍放在对应 `docs/concepts/**` 文档中；本文只锁定术语、边界和现有映射。

## How To Use

- 新增 domain concept 前，先检查是否已有对应 term。
- 修改 API、Proto、frontend type 字段前，确认名称是否与 canonical term 一致。
- 写文档时优先使用 `Preferred forms`，避免继续扩散 `Avoid forms`。
- 遇到历史名称、通俗叫法或边界未定的词，先标记为 `Proposed`，不要借 glossary 推动代码改名。

## Entry Status

| Status       | 语义                                                         |
| ------------ | ------------------------------------------------------------ |
| `Proposed`   | 建议术语或建议边界，尚未经过充分使用验证。                   |
| `Accepted`   | 已由现有概念文档、`pkg/types` contract 或稳定 API 明确定义。 |
| `Deprecated` | 历史术语，不建议新增使用。                                   |
| `Replaced`   | 已被其他 canonical term 替代。                               |

## Inclusion Rules

只有以下变化需要检查或更新 glossary：

- 新增跨 service 共享的 domain concept。
- 新增或修改 API、Proto、frontend type 字段。
- 新增状态枚举、权限 action、IAM resource。
- 同一概念在代码或文档中出现两个以上名称。
- 术语边界被新的需求改变。

不要求每个 internal helper、DAO 字段、临时 UI 文案都进入 glossary。

## Terms

术语先按功能分块阅读，再在块内处理相近概念边界。功能块只影响阅读路径，不改变 term status。

| Functional block      | Covers                                                                | When to read                                                  |
| --------------------- | --------------------------------------------------------------------- | ------------------------------------------------------------- |
| Host Context          | 主机领域模型、CMDB 主机身份和节点运行态属性                           | 讨论主机身份、Host.Static、Host.Dynamic 或主机网络参数时。    |
| Node and Connectivity | Node、Network Area、Network Unit、Access Point、节点角色和 Proxy 边界 | 讨论节点与管控单元/区域、跨网络管理、Proxy 语义或拓扑入口时。 |
| Access Control        | `Scope`、授权范围、权限和资源                                         | 讨论 IAM、权限 action、resource 或请求可见范围时。            |
| Workflow              | workflow、operation 和 operation instance                             | 讨论底层任务编排、执行步骤和运行记录机制时。                  |
| Operation Lifecycle   | 安装、升级、重载和更新动作                                            | 讨论节点或插件生命周期 action 与更新动词边界时。              |
| Deploy Policy         | 部署状态、部署策略和部署目标 spec                                     | 讨论部署范围、目标规格、策略和部署结果语义时。                |
| Plugin Runtime        | 插件包、插件和进程                                                    | 讨论 plugin package、tenant plugin 或 process 语义时。        |
| Ownership             | 租户和业务                                                            | 讨论资源归属、租户边界或 CMDB 业务维度时。                    |
| Host Credentials      | 主机凭证和 historical `credit`                                        | 讨论安装凭证、host credential vault 或历史命名时。            |

### Host Context

#### Host

| Status     | Term | Definition                                                                                   | Preferred forms | Avoid forms                 | Distinguish from                                                                   | Code/API mapping                                           | Source                                            |
| ---------- | ---- | -------------------------------------------------------------------------------------------- | --------------- | --------------------------- | ---------------------------------------------------------------------------------- | ---------------------------------------------------------- | ------------------------------------------------- |
| `Accepted` | Host | 节点管理中的主机领域模型，承载 CMDB 主机属性、节点管理运行态属性，以及节点操作所需网络参数。 | `Host`、主机    | 把所有 Node 语义都写成 Host | `Node` 是产品语境下更宽的管理对象；`Host` 强调 CMDB 主机身份和节点管理运行态模型。 | `pkg/types.Host`、`Host.Static`、`Host.Dynamic`、`host_id` | `docs/concepts/node/host.md`、`pkg/types/host.go` |

### Node and Connectivity

#### Node / Network Area / Network Unit / Access Point

| Status     | Term         | Definition                                                                                               | Preferred forms          | Avoid forms                                   | Distinguish from                                                 | Code/API mapping                                                       | Source                                                                                                                         |
| ---------- | ------------ | -------------------------------------------------------------------------------------------------------- | ------------------------ | --------------------------------------------- | ---------------------------------------------------------------- | ---------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `Proposed` | Node         | 由 bk-nodemgr 管理的节点概念，通常围绕 Agent 或 Proxy 的安装、升级、重载、卸载等生命周期操作展开。       | `Node`、节点             | 在需要表达 CMDB 主机身份时使用 Node 替代 Host | `Host` 是主机领域模型；`Agent` 和 `Proxy` 是节点角色或运行组件。 | `node_agent.proto`、`node_proxy.proto`、`docs/concepts/node/README.md` | `docs/concepts/node/control_mode.md`、`proto/application/api/v3/node_agent.proto`、`proto/application/api/v3/node_proxy.proto` |
| `Proposed` | Network Area | GSE 网络区域维度，用于区分主机所属的 GSE 网络范围。                                                      | `Network Area`、管控区域 | 在同一文档中混用网络区域、管控区域            | `Network Unit` 是节点管理维护的代理管控基本单元。                | `bk_networkarea_id`、`network_area`                                    | `docs/concepts/node/host.md`、`proto/application/api/v3/topo.proto`                                                            |
| `Accepted` | Network Unit | 节点管理的代理管控基本单元；同一单元内节点经 GSE Proxy 与上游通信，由该单元内 Proxy 承载跨网络管理能力。 | `Network Unit`、管控单元 | 在同一文档中混用网络单元、管控单元            | `Network Area` 是 GSE 网络区域；`Access Point` 是访问入口概念。  | `bk_networkunit_id`、`NetworkUnit`                                     | `docs/concepts/topo/networkunit.md`、`docs/concepts/node/host.md`                                                              |
| `Proposed` | Access Point | 拓扑中的接入点概念，用于表达节点或单元访问管控面能力的入口。                                             | `Access Point`、接入点   | Network Unit                                  | `Network Unit` 是管控单元；Access Point 是访问入口。             | `access_point`、topo API/type                                          | `proto/application/api/v3/topo.proto`、`front/src/@types/topo.d.ts`                                                            |

#### Agent / Proxy

| Status     | Term  | Definition                                                                                                                | Preferred forms     | Avoid forms                            | Distinguish from                                                               | Code/API mapping                                                         | Source                                                             |
| ---------- | ----- | ------------------------------------------------------------------------------------------------------------------------- | ------------------- | -------------------------------------- | ------------------------------------------------------------------------------ | ------------------------------------------------------------------------ | ------------------------------------------------------------------ |
| `Proposed` | Agent | 普通节点角色，承载目标主机上的 GSE Agent 能力，并参与 install、upgrade、restart、reconfig、uninstall 等节点生命周期操作。 | `Agent`、Agent 节点 | Host、Proxy                            | `Host` 是主机模型；`Proxy` 是跨网络管理能力的节点角色。                        | `node_role`、`node_agent.proto`、`pkg/types/node_agent.go`               | `docs/concepts/node/host.md`、`docs/concepts/node/control_mode.md` |
| `Proposed` | Proxy | 跨网络场景下承载 Agent 控制、文件通道、数据通道、Relay 服务等能力的节点角色。                                             | `Proxy`、Proxy 节点 | 只写成 proxy 导致与 proxy service 混淆 | `Agent` 是普通节点角色；`proxy service` 是服务入口；`gse-proxy` 是外部组件名。 | `node_role`、`proxy_tags`、`node_proxy.proto`、`pkg/types/node_proxy.go` | `docs/concepts/node/host.md`、`docs/concepts/topo/networkunit.md`  |

#### Proxy node / Proxy process / proxy service / gse-proxy

| Status     | Term          | Definition                                                              | Preferred forms             | Avoid forms       | Distinguish from                                                                  | Code/API mapping                    | Source                                                                                        |
| ---------- | ------------- | ----------------------------------------------------------------------- | --------------------------- | ----------------- | --------------------------------------------------------------------------------- | ----------------------------------- | --------------------------------------------------------------------------------------------- |
| `Proposed` | Proxy node    | `node_role` 为 Proxy 的节点，承担所在 Network Unit 内的跨网络管理能力。 | `Proxy node`、Proxy 节点    | proxy、gse-proxy  | `Proxy process` 是运行进程；`proxy service` 是服务入口；`gse-proxy` 是 GSE 组件。 | `node_role`、`proxy_tags`           | `docs/concepts/node/host.md`、`docs/concepts/topo/networkunit.md`                             |
| `Proposed` | Proxy process | Proxy 节点上被托管或执行的具体进程。                                    | `Proxy process`、Proxy 进程 | Proxy node        | `Proxy node` 是节点角色；process 是运行实例。                                     | `Process`、GSE trusteeship          | `docs/concepts/plugin/README.md`、`docs/concepts/plugin/v2_v3_compatibility_and_migration.md` |
| `Proposed` | proxy service | backend 面向 GSE cluster 或相关通信链路提供的 proxy 服务入口。          | `proxy service`             | Proxy、Proxy node | `Proxy node` 是被管控节点角色；`proxy service` 是服务端能力。                     | `cmd/backend`、service startup docs | `docs/operation/architecture.md`                                                              |
| `Proposed` | gse-proxy     | GSE 体系中的具体外部组件名，不等同于 bk-nodemgr 的 Proxy 节点领域概念。 | `gse-proxy`                 | Proxy             | `Proxy node` 是节点管理角色；`gse-proxy` 是外部组件。                             | GSE external projection             | `docs/concepts/plugin/v2_v3_compatibility_and_migration.md`、`pkg/types/gse.go`               |

### Access Control

#### Scope / Authorized Scope / Permission / Resource

| Status     | Term             | Definition                                                            | Preferred forms              | Avoid forms              | Distinguish from                                                   | Code/API mapping                          | Source                                                                |
| ---------- | ---------------- | --------------------------------------------------------------------- | ---------------------------- | ------------------------ | ------------------------------------------------------------------ | ----------------------------------------- | --------------------------------------------------------------------- |
| `Accepted` | Scope            | `pkg/types` 中的资源范围 union type；同一时间只能有一个 branch 生效。 | `Scope`                      | 范围、作用域、授权域混用 | `Authorized Scope` 是权限结果；`Resource` 是被保护对象。           | `pkg/types.Scope`、`NewScopeWith*`        | `pkg/types/README.md`、`pkg/types/scope.go`                           |
| `Proposed` | Authorized Scope | IAM 或权限判断后允许访问的资源范围。                                  | `Authorized Scope`、授权范围 | Scope                    | `Scope` 是 domain value object；Authorized Scope 是授权计算结果。  | IAM projection、permission helper         | `pkg/types/iam.go`、`docs/developer/iam_provider_guide.md`            |
| `Proposed` | Permission       | 对特定 action 和 resource 的访问许可。                                | `Permission`、权限           | Scope                    | `Resource` 是被保护对象；Permission 是能否执行 action 的判断结果。 | IAM action、`front/src/constants/auth.ts` | `docs/developer/iam_provider_guide.md`、`front/src/constants/auth.ts` |
| `Proposed` | Resource         | IAM 权限模型中的受保护对象类型或实例。                                | `Resource`、资源             | Instance                 | `Resource Instance` 是具体资源实例；Permission 是访问许可。        | IAM resource、resource instance           | `docs/developer/iam_provider_guide.md`、`pkg/types/iam.go`            |

### Workflow

#### Workflow / Operation / Operation Instance

| Status     | Term               | Definition                                                                     | Preferred forms                | Avoid forms                                             | Distinguish from                                                            | Code/API mapping                                    | Source                                                                |
| ---------- | ------------------ | ------------------------------------------------------------------------------ | ------------------------------ | ------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------- | --------------------------------------------------------------------- |
| `Proposed` | Workflow           | 节点或插件操作的任务过程模型，组织一次 install、upgrade、reconfig 等业务过程。 | `Workflow`、工作流             | 把 workflow、operation、operation instance 全部写成任务 | `Operation` 是 workflow 内的操作语义；`Operation Instance` 是实际执行记录。 | `pkg/types/node_workflow.go`、`node_workflow.proto` | `pkg/types/README.md`、`proto/application/api/v3/node_workflow.proto` |
| `Proposed` | Operation          | workflow 中的操作语义，表示一个可执行或可追踪的步骤。                          | `Operation`、操作              | Workflow                                                | `Workflow` 是整体过程；`Operation Instance` 是某次执行记录。                | operation fields/types                              | `pkg/types/README.md`、`proto/application/api/v3/node_workflow.proto` |
| `Proposed` | Operation Instance | 某个 operation 在一次 workflow 中的实际执行记录。                              | `Operation Instance`、操作实例 | Operation                                               | `Operation` 是抽象步骤；Operation Instance 是运行记录。                     | operation instance fields/types                     | `pkg/types/README.md`、`proto/application/api/v3/node_workflow.proto` |

### Operation Lifecycle

#### Install / Upgrade / Reconfig / Update

| Status     | Term     | Definition                                                                                        | Preferred forms        | Avoid forms           | Distinguish from                                                                          | Code/API mapping                                                                                      | Source                                                                                                  |
| ---------- | -------- | ------------------------------------------------------------------------------------------------- | ---------------------- | --------------------- | ----------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| `Proposed` | Install  | 将 Agent、Proxy 或 plugin 首次部署到目标节点的生命周期动作。                                      | `Install`、安装        | Update                | `Upgrade` 通常涉及已有对象版本升级；`Reconfig` 触发配置重新生成和状态回写。               | `/install`、`LaunchInstall*`                                                                          | `docs/concepts/node/control_mode.md`、`pkg/types/node_agent.go`、`pkg/types/node_proxy.go`              |
| `Proposed` | Upgrade  | 对已有 Agent、Proxy 或 plugin 执行版本升级的生命周期动作。                                        | `Upgrade`、升级        | Update                | `Update` 是字段、记录或 policy 的变更动词，不默认表示版本升级。                           | `/upgrade`、`LaunchUpgrade*`                                                                          | `docs/concepts/node/control_mode.md`、`pkg/types/node_agent.go`、`pkg/types/node_proxy.go`              |
| `Proposed` | Reconfig | 触发 installer 重新渲染配置并回写执行状态的节点重载动作。                                         | `Reconfig`、重载、重配 | Upgrade、Update       | `Upgrade` 改变版本；`Update` 修改字段、记录或 policy，不一定触发 installer 生命周期动作。 | `/reconfig`、`reconfig_options`                                                                       | `docs/concepts/node/control_mode.md`、`pkg/types/node_agent.go`、`pkg/types/node_proxy.go`              |
| `Proposed` | Update   | 对已有记录、字段、metadata、ops fields 或 policy 做变更的通用动词；不作为节点版本升级的推荐术语。 | `Update`、更新         | 把版本升级写成 Update | `Upgrade` 是 lifecycle version transition；`Reconfig` 是 installer 配置重载。             | `update_ops_fields`、`UpdateHostOpsFields`、`UpdateActionInstanceLifecycle`、`DeployPolicySvc.Update` | `pkg/thirdparty/cmdb/handler.go`、`pkg/workflow/storage.go`、`proto/backend/api/v3/deploy_policy.proto` |

#### Manual Install / Offline Install

| Status     | Term            | Definition                                                                     | Preferred forms             | Avoid forms     | Distinguish from                                                    | Code/API mapping               | Source                               |
| ---------- | --------------- | ------------------------------------------------------------------------------ | --------------------------- | --------------- | ------------------------------------------------------------------- | ------------------------------ | ------------------------------------ |
| `Proposed` | Manual Install  | Backend 生成 bootstrap 命令，由管理员手动执行；重点是触发方式由人执行。        | `Manual Install`、手动安装  | Offline Install | `Offline Install` 关注目标环境无法访问管控面网络。                  | `is_manual`、bootstrap command | `docs/concepts/node/control_mode.md` |
| `Proposed` | Offline Install | 目标环境无法访问管控面网络时，管理员手动下载离线包、执行并回写状态的安装形态。 | `Offline Install`、离线安装 | Manual Install  | Manual Install 关注执行方式；Offline Install 关注网络条件和离线包。 | `is_offline`、Offline mode     | `docs/concepts/node/control_mode.md` |

### Deploy Policy

#### Deployment / Deploy Policy / Deploy Spec

| Status     | Term          | Definition                                                            | Preferred forms                            | Avoid forms              | Distinguish from                                         | Code/API mapping                             | Source                                                                              |
| ---------- | ------------- | --------------------------------------------------------------------- | ------------------------------------------ | ------------------------ | -------------------------------------------------------- | -------------------------------------------- | ----------------------------------------------------------------------------------- |
| `Proposed` | Deployment    | 节点或插件部署状态模型，记录部署对象、目标、版本和执行结果相关语义。  | `Deployment`、部署                         | Deploy Policy            | `Deploy Policy` 是策略；`Deploy Spec` 是目标规格。       | `node_deployment.go`、`plugin_deployment.go` | `pkg/types/README.md`                                                               |
| `Proposed` | Deploy Policy | 决定部署范围、目标和行为的策略模型。                                  | `Deploy Policy`、部署策略                  | Deployment               | Deployment 是执行或状态结果；Deploy Policy 是策略。      | `deploy_policy.go`、`deploy_policy.proto`    | `docs/concepts/deploy_policy/README.md`、`proto/backend/api/v3/deploy_policy.proto` |
| `Accepted` | Deploy Spec   | `pkg/types` 中的部署目标 union type；同一时间只能有一个 branch 生效。 | `Deploy Spec`、`DeploySpec`、部署目标 spec | 多个目标 branch 同时生效 | `Deploy Policy` 是策略；Deploy Spec 是策略内的目标规格。 | `DeploySpec`、`NewDeploySpecWith*`           | `pkg/types/README.md`、`docs/concepts/deploy_policy/spec.md`                        |

### Plugin Runtime

#### Plugin / Plugin Package / Process

| Status     | Term           | Definition                                                                                      | Preferred forms                   | Avoid forms             | Distinguish from                                            | Code/API mapping                      | Source                                                                                        |
| ---------- | -------------- | ----------------------------------------------------------------------------------------------- | --------------------------------- | ----------------------- | ----------------------------------------------------------- | ------------------------------------- | --------------------------------------------------------------------------------------------- |
| `Accepted` | Plugin Package | 全局共享的静态插件资源集合，可有多个版本，本身没有运行生命周期。                                | `Plugin Package`、插件包、Package | Plugin                  | `Plugin` 是租户内业务对象；`Process` 是主机上的运行实例。   | release/package model                 | `docs/concepts/plugin/README.md`、`pkg/types/release.go`                                      |
| `Accepted` | Plugin         | 租户范围内的插件业务对象，本身不携带版本，也没有运行生命周期。                                  | `Plugin`、插件                    | Plugin Package、Process | Plugin Package 是静态资源；Process 是每台主机上的运行实例。 | `pkg/types.Plugin`、plugin APIs       | `docs/concepts/plugin/README.md`、`pkg/types/plugin.go`                                       |
| `Accepted` | Process        | 主机上某个 plugin 的运行实例，有版本和生命周期；每台主机上同一插件只能有一个 process instance。 | `Process`、进程                   | Plugin                  | Plugin 是业务对象；Process 是运行实例。                     | process fields/types、GSE trusteeship | `docs/concepts/plugin/README.md`、`docs/concepts/plugin/v2_v3_compatibility_and_migration.md` |

### Ownership

#### Tenant / Business / Biz

| Status     | Term     | Definition                                                    | Preferred forms    | Avoid forms                          | Distinguish from                                   | Code/API mapping                | Source                                                               |
| ---------- | -------- | ------------------------------------------------------------- | ------------------ | ------------------------------------ | -------------------------------------------------- | ------------------------------- | -------------------------------------------------------------------- |
| `Proposed` | Tenant   | 多租户边界概念，用于区分不同租户下的资源、插件和权限语义。    | `Tenant`、租户     | Business                             | `Business` / `Biz` 是 CMDB 或 IAM 业务资源维度。   | `tenant_id`、`pkg/types.Tenant` | `pkg/types/README.md`、`pkg/types/tenant.go`、`pkg/tenant/README.md` |
| `Proposed` | Business | CMDB 业务拓扑中的业务资源。                                   | `Business`、业务   | Tenant                               | Tenant 是租户边界；Business 是 CMDB/IAM 资源维度。 | `biz_id`、`bk_biz_id`           | `docs/concepts/node/host.md`、`pkg/types/cmdb.go`                    |
| `Proposed` | Biz      | Business 的常用缩写，主要保留在代码、API 字段和 CMDB 语境中。 | `Biz`、`bk_biz_id` | 在自然语言中用 Biz 替代所有 Business | `Business` 是完整术语；Biz 是代码/API 常见缩写。   | `bk_biz_id`、`biz_id`           | `docs/concepts/node/host.md`、`front/src/constants/auth.ts`          |

### Host Credentials

#### Credential / historical credit

| Status     | Term              | Definition                                                                         | Preferred forms                     | Avoid forms                          | Distinguish from                                        | Code/API mapping                         | Source                           |
| ---------- | ----------------- | ---------------------------------------------------------------------------------- | ----------------------------------- | ------------------------------------ | ------------------------------------------------------- | ---------------------------------------- | -------------------------------- |
| `Proposed` | Credential        | 用于主机登录、安装或外部 vault 集成的凭证语义。                                    | `Credential`、凭证、host credential | credit（新增自然语言文档中避免使用） | `credit` 是历史代码命名时，不自动代表新的自然语言术语。 | host credential vault、credential config | `docs/operation/installation.md` |
| `Proposed` | historical credit | 仓库历史命名中出现的 `credit`，第一版只作为 historical name 记录，不提出代码改名。 | historical `credit`                 | 新增文档继续用 credit 表示凭证       | `Credential` 是推荐自然语言术语。                       | existing credit paths/names              | `docs/operation/installation.md` |

### Naming Boundary

同一个 domain term 在不同边界可以有不同 serialization form：

| Boundary                      | Naming rule source                      | Example                       |
| ----------------------------- | --------------------------------------- | ----------------------------- |
| Proto Message / Service / RPC | `proto/README.md`                       | `NodeAgentInstall`            |
| Proto field / JSON field      | `proto/README.md`                       | `bk_biz_id`、`network_unit`   |
| Go domain model               | `pkg/types/README.md`                   | `Host`、`Scope`、`DeploySpec` |
| Frontend API/type             | generated API and `front/src/@types/**` | `NodeAgent`、`NodeProxy`      |
| Concept docs                  | `docs/concepts/**`                      | `Network Unit（管控单元）`    |

这些格式差异不是术语冲突。Glossary 只约束概念边界和推荐自然语言表达，不要求把现有 API、Proto、Go type 或 frontend type 统一改名。
