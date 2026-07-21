# 插件 V2 / V3 部署模式、兼容共存与迁移方案

> 本文整理 bk-nodemgr 当前插件方案的设计原理与考虑原因，区分 V2 与 V3 两种部署模式，说明同机双 `bkmonitorbeat` 共存的动机与实现，阐述如何保证仅一台 bkm 上报基础数据，并梳理从 V2 迁移到 V3（含子插件迁移）的路径。
>
> 文中所有结论均附 `file_path:line_number` 引用，便于核对实现。

## 1. 背景与目标

bk-nodemgr 的"插件"是运行在被管控主机上的采集/执行进程（典型代表是 `bkmonitorbeat`），由节点管理通过 GSE Agent 托管拉起。历史上插件方案经历过一次代次升级：

- **V2（旧版官方插件）**：前端 i18n 称"V2官方插件" (`front/locales/zh-CN.yml:943`)，对应 `ReleaseTypeOriginPluginV2` / `ReleaseTypeOriginExternalPluginV2` (`pkg/types/release.go:33-37`)。
- **V3（标准插件）**：前端 i18n 称"标准插件" (`front/locales/zh-CN.yml:942`)，对应 `ReleaseTypeOriginPluginV3` (`pkg/types/release.go:39-40`)，包结构规范见 `docs/developer/plugin_pkg_build_guide.md:1`。

由于 V2 在生产环境存量巨大、且承担基础数据上报职责，无法一次性切换到 V3。为此引入**兼容模式（compatibility mode）**，允许 V2 与 V3 在同一主机共存，再通过 `migrate_from_v2` 工作流平滑迁移。

设计目标：

1. V3 可独立部署、独立升级，不依赖 V2 退场；
2. 共存期间**不能产生双倍上报**——基础数据仍由 V2 上报，V3 静默；
3. 子插件（子配置）可从 V2 迁移到 V3，迁移动作先挂载到 V3、再下线 V2，避免数据断档；
4. 全程通过工作流 (workflow) + GSE 托管实现，无需 SSH 旁路。

## 2. 术语对齐

仓库中的术语与外部说法存在差异，先做映射以免歧义：

| 外部说法 / 通俗叫法 | 仓库内实际术语 | 出处 |
|---|---|---|
| V2 / V3 部署模式 | 插件格式代次（V2 / V3） | `pkg/types/release.go:33-40` |
| Agent V2 / V3 | `NodeGeneration`（1=Agent 1.x 已不支持，2=Agent 2.x） | `pkg/types/generation.go`；`docs/developer/plugin_pkg_build_guide.md:91` |
| 官方进程 / 托管进程 | GSE 托管进程（trusteeship），GSE Agent 充当 supervisor | `internal/backend/manager/workflowdef/plugin/action_trusteeship_process_to_gse.go` |
| 子插件 | 子配置（subconfig / sub_config） | `docs/developer/plugin_pkg_build_guide.md:36,40`；`pkg/deployconstant/plugin_deploy_constant.go:145` |
| 基础数据上报 | `dataid` 配置项（dataid=-1 即不上报） | `internal/backend/manager/workflowdef/plugin/action_overwrite_plugin_config_for_compatibility.go:32,158` |
| 共存 | 兼容模式（compatibility mode） | `pkg/globalsettings/definition.go:27` |

> 注意：仓库内**没有** "coexist / 共存 / sidecar / operator / basereport / 官方进程" 等字面术语。"共存"是兼容模式的运行时效果，"基础数据上报"由 `dataid` 配置项控制，"子插件"在仓库里叫"子配置"。

## 3. 设计原理与考虑

### 3.1 为什么需要 V3

V2 存在若干历史包袱：

- **目录与命名耦合**：V2 部署在 `plugins/` 目录，路径通过把 `plugin` 替换为 `plugins` 生成，规则隐式 (`internal/backend/manager/workflowdef/pluginv2/action_inject_plugin_base_runtime_v2.go:214-248`)。
- **子配置路径固定**：V2 提供固定变量 `subconfig_path`，子配置目录由系统拼接 (`internal/backend/manager/workflowdef/pluginv2/action_ensure_and_update_plugin_config_details_v2.go:538`)。
- **模板渲染**：V2 以 Jinja2 为主。
- **GSE namespace**：V2 使用 `nodeman` 命名空间 (`internal/backend/manager/workflowdef/pluginv2/pluginv2.go:48`)。

V3 重新设计：

- 目录改为 `plugin/` (`pkg/deployconstant/plugin_deploy_constant.go:26`)，单数。
- 移除 `subconfig_path` 变量，子配置路径由 `definition.yaml` 中各子配置的 `filePath` 显式声明 (`docs/developer/plugin_pkg_build_guide.md:36-42`)，更安全（可防路径遍历）也更灵活。
- 模板渲染推荐 `go-template`，渲染上下文使用 PascalCase (`internal/backend/manager/workflowdef/plugin/action_ensure_and_update_plugin_config_details.go:237-252`)。
- GSE namespace 改为 `bk-nodemgr` (`pkg/thirdparty/gse/handler_proc.go:67`)。

### 3.2 为什么需要兼容模式（不能直接切换）

直接切换 V2 → V3 有两类风险：

1. **基础数据断档**：V2 `bkmonitorbeat` 负责的系统指标上报一旦停止，监控会有空洞；而 V3 是否能立即接管取决于配置正确性，不能赌。
2. **存量主机多**：V2 在生产大量使用，灰度切换比一刀切安全。

因此设计选择：**共存期双部署 + 仅一方上报**。V2 继续上报基础数据，V3 进入"静默"状态（`dataid=-1`），待 V3 验证可用后通过迁移工作流下线 V2。

### 3.3 为什么用 dataid=-1 而不是端口隔离

代码中**没有**通过端口区分两个 bkm 的数据流。区分手段是**配置覆写**：兼容模式下把 V3 配置文件中所有匹配 `dataid` 的行值改为 `-1` (`action_overwrite_plugin_config_for_compatibility.go:32,158-160`)。`-1` 是 bkm 约定的"无效 dataid"，相当于禁用该数据流上报。

选择 dataid 而非端口的考虑：

- bkm 的数据上报链路依赖 dataid 路由到监控平台，dataid 是天然的数据流开关；
- 端口隔离需要双端口规划、冲突处理，复杂度高；
- dataid 覆写只改配置文件内容，回滚成本低（关掉兼容模式即可恢复）。

### 3.4 为什么 V2/V3 可以同名进程共存

GSE 托管进程的唯一 key 是 `agentID:namespace:procName` (`pkg/thirdparty/gse/handler_proc.go:145`)。V2 namespace = `nodeman`，V3 namespace = `bk-nodemgr`，**namespace 不同 → key 不同 → 两个 `bkmonitorbeat` 进程可以被 GSE 同时托管而不冲突**。这是同机双 bkm 共存的物理基础。

## 4. V2 插件方案

### 4.1 架构

- **Manager 入口**：`internal/backend/manager/plugin_v2_manager.go`，提供 `LaunchPluginEnsurePluginV2` / `LaunchUninstallPluginV2` / `LaunchStopPluginV2`。
- **接口**：`internal/backend/manager/iface/iface.go:89-99`（`iPluginManagerPluginV2` 子接口）。
- **工作流定义目录**：`internal/backend/manager/workflowdef/pluginv2/`（包注释 `pluginv2.go:11`）。
- **Action 注册**：`internal/backend/manager/register_defs.go:265-308`（`registerDefPluginV2`）。

### 4.2 部署目录与 namespace

- 部署根目录 `plugins/`：`action_inject_plugin_base_runtime_v2.go:214-219` (`pluginV2BaseDirName = "plugins"`)。
- 目录生成：把路径中的 `plugin` 替换为 `plugins`（`GeneratePluginV2HomeDir`，:222-248）。
- GSE namespace = `nodeman`：`pluginv2.go:48` (`procNameSpaceNodeMan`)。
- 进程托管：`action_trusteeship_process_to_gse_v2.go:32-38` 用 `gse.WithProcNameSpace(procNameSpaceNodeMan)` 创建 handler，调用 `TrusteeshipAndStartProcess` (`action_start_process_v2.go:36`)。

### 4.3 V2 Ensure 工作流

`oper_plugin_ensure_plugin_v2.go:46-59`，action 链：

```
InjectPluginBaseRuntimeV2 → FinishIfPluginProcessV2Alive → VerifyPluginAvailabilityV2 →
RenderPluginDeploymentV2 → EnsureAndUpdatePluginConfigDetailsV2 → RenderPluginConfigV2 →
TransferPluginPkgToNodeV2 → InstallPluginV2 → WaitPluginInstallerCompleteV2 → StartProcessV2
```

### 4.4 V2 子配置（子插件）管理

- 子配置路径变量 `subconfig_path`：`action_ensure_and_update_plugin_config_details_v2.go:429` (`keySubConfigPath`)。
- 注释明确 V2 仍需提供固定 subconfig_path：`:538`。
- 子配置目录生成：`action_inject_plugin_base_runtime_v2.go:246-248` (`GeneratePluginV2SubConfigDir`)。
- 独立的 apply_subconfig 工作流：`oper_apply_plugin_subconfig_v2.go`。

## 5. V3 插件方案

### 5.1 架构

- **Manager 入口**：`internal/backend/manager/plugin_manager.go`，提供 `LaunchInstallPlugin` / `LaunchUpgradePlugin` / `LaunchUninstallPlugin` / `LaunchApplyPluginSubConfig` / `LaunchRestartProcess` / `LaunchMigrateFromV2` / `LaunchStopProcess`。
- **接口**：`internal/backend/manager/iface/iface.go:65-87`（`iPluginManagerPlugin` 子接口）。
- **工作流定义目录**：`internal/backend/manager/workflowdef/plugin/`。
- **Action 注册**：`internal/backend/manager/register_defs.go:212-261`（`registerDefPlugin`）。

### 5.2 部署目录与 namespace

- 部署根目录 `plugin/`：`pkg/deployconstant/plugin_deploy_constant.go:26` (`PluginBaseDirName = "plugin"`)。
- GSE namespace = `bk-nodemgr`：`pkg/thirdparty/gse/handler_proc.go:67` (`procNameSpaceNodemgr`)。
- 进程托管：V3 的 action 创建 GSE handler 时**不传** `WithProcNameSpace`，使用默认值 `bk-nodemgr`（`action_trusteeship_process_to_gse.go:35-37`、`action_restart_process.go:36`）。

### 5.3 V3 Install 工作流

`oper_install_plugin.go:46-62`，action 链：

```
TryStopProcess → UpsertProcess → VerifyPluginAvailability → InjectPluginCustomDeployConfig →
RenderPluginDeployment → EnsureAndUpdatePluginConfigDetails → RenderPluginConfig →
OverwritePluginConfigForCompatibility → TransferPluginPkgToNode → InstallPlugin →
WaitPluginInstallerComplete → StartProcess → UpdateProcess
```

注意 `OverwritePluginConfigForCompatibility`（:55）是兼容模式的关键 hook，V2 工作流中没有对应 action。

### 5.4 V3 子配置（子插件）管理

- 子配置路径由 `definition.yaml` 中各子配置的 `filePath` 声明：`docs/developer/plugin_pkg_build_guide.md:39-42`。
- `FetchProcessSubConfigIntoDeployment`：查询 `IsMainConfig: false` 的 ProcessConfig 并注入 deployment（`action_fetch_process_config_into_deployment.go`）。
- ApplyPluginSubConfig 工作流：`oper_apply_plugin_subconfig.go`，含 `OverwritePluginConfigForCompatibility` + `ReloadProcess`。
- API 路由：`internal/backend/router/api-v3/plugin/apply_subconfig.go:23` (`ApplySubConfig` handler)。
- SubConfigDir 生成：`pkg/deployconstant/plugin_deploy_constant.go:145-147` (`GeneratePluginSubConfigDir`)。

## 6. V2 与 V3 异同对比

### 6.1 差异

| 维度 | V2 | V3 | 出处 |
|------|-----|-----|------|
| 部署目录 | `plugins/`（复数） | `plugin/`（单数） | `action_inject_plugin_base_runtime_v2.go:215` vs `pkg/deployconstant/plugin_deploy_constant.go:26` |
| GSE namespace | `nodeman` | `bk-nodemgr` | `pluginv2/pluginv2.go:48` vs `pkg/thirdparty/gse/handler_proc.go:67` |
| GSE 进程 key | `agentID:nodeman:procName` | `agentID:bk-nodemgr:procName` | `pkg/thirdparty/gse/handler_proc.go:145` |
| 子配置路径变量 | 提供 `subconfig_path`（固定） | 不提供，改用 `filePath` 显式声明 | `action_ensure_and_update_plugin_config_details_v2.go:538` vs `docs/developer/plugin_pkg_build_guide.md:36` |
| 模板渲染器 | Jinja2 为主 | go-template（推荐） | `docs/developer/plugin_pkg_build_guide.md:145-149` |
| 渲染上下文命名 | — | PascalCase | `action_ensure_and_update_plugin_config_details.go:237-252` |
| 安装入口 | `LaunchPluginEnsurePluginV2` | `LaunchInstallPlugin` | `plugin_v2_manager.go` vs `plugin_manager.go` |
| 兼容模式 hook | 无 | `OverwritePluginConfigForCompatibility` | `oper_install_plugin.go:55` |
| 工作流类型常量 | `PluginWorkflowTypePluginEnsureV2` 等 | `PluginWorkflowTypeInstall` 等 | `pkg/types/plugin_workflow.go` |

### 6.2 共同点（复用组件）

- **相同的 Storage 层**：V2 与 V3 共用 `pluginStg.IStorage` / `IDaoPluginDeployment` / `IDaoProcess` / `IDaoPluginWorkflow`，对比 `plugin/plugin.go:37-43` 与 `pluginv2/pluginv2.go:36-42` 字段完全一致。
- **相同的领域类型**：`PluginDeployment` / `PluginDeploymentInfo` / `PluginDeploymentPluginConf`（`pkg/types/plugin_deployment.go:22-26`）、`PluginWorkflow`（`pkg/types/plugin_workflow.go:21-33`）。
- **相同的 GSE Handler**：都通过 `capability.GSEHandler.NewHandlerProc(...)` 创建，仅 namespace 选项不同。
- **相同的 Release / CMDB / File Handler**：`release.IStorage` / `capability.CMDBHandler` / `capability.FileHandler`。
- **相同的进程托管模型**：均通过 GSE trusteeship 拉起，GSE Agent 充当 supervisor，进程异常退出自动拉起。

## 7. 同机双 bkmonitorbeat 场景

### 7.1 为什么会同时存在两个

为了灰度迁移。V3 是新方案，但 V2 仍在生产大量使用且承担基础数据上报。为平滑切换：

1. 先同时部署 V2 + V3 两个 `bkmonitorbeat`（namespace 不同，可共存，见 §3.4）；
2. V3 处于兼容模式，`dataid` 被覆写为 `-1`，**只跑子插件、不上报基础数据**（见 §8）；
3. V2 继续负责基础数据上报；
4. 最终通过 `MigrateFromV2` 工作流下线 V2（见 §9）。

### 7.2 双安装的触发点

代码中有两处会在安装 V3 的同时额外安装一份 V2：

#### 触发点一：Agent / Proxy 安装时预置插件

`internal/backend/manager/workflowdef/node/action_install_pre_ordered_plugins.go`

`Do` 方法中**并行**起两个 goroutine（:143-175）：

- `installPreOrderedPlugin`（:145-159）→ `LaunchInstallPlugin` 安装 V3；
- `installPreOrderedPluginV2`（:161-175）→ `LaunchPluginEnsurePluginV2` 安装 V2。

V2 安装的门控条件（:315-317）：

```go
// v2 plugin is special, it not controlled by InstallPreOrderedPlugins.
if !deployInfo.InstallOptions.EnableCompatibilityMode || !deployInfo.InstallOptions.InstallPreOrderedPlugins {
    return false, "", nil
}
```

即 `EnableCompatibilityMode && InstallPreOrderedPlugins` 同时为 true 才装 V2。

预置插件清单：

- V3（:303-308）：Agent → `[bkmonitorbeat]`，Proxy → `[bkmonitorbeat, bk-nodemgr-relay]`；
- V2（:402-407）：Agent → `[bkmonitorbeat]`，Proxy → `[bkmonitorbeat]`。

两者都按 `deployInfo.Host.Dynamic.NodeGeneration` 调 `GetReleasePluginDefaultVersion` 取默认版本（:242, :356），保证版本与 Agent 代次匹配。

#### 触发点二：插件 Install API 调用

`internal/backend/router/api-v3/plugin/install.go`

`Install` handler 中：

1. `installPlugin`（:53）→ `LaunchInstallPlugin` 安装 V3；
2. `ensurePluginV2`（:58）→ `LaunchPluginEnsurePluginV2` 安装 V2。

V2 安装过滤（:104-106）：只保留 `EnableCompatibilityMode == true` 的插件才装 V2：

```go
pluginDeploymentParam = slices.DeleteFunc(pluginDeploymentParam, func(item *types.PluginDeploymentParam) bool {
    return !item.EnableCompatibilityMode
})
```

### 7.3 兼容模式何时被判定

`EnableCompatibilityMode` 字段并非由客户端传入（proto 中无 `compatibility` 字段，grep 0 命中），而是由服务端根据全局策略解析：

- Agent 安装：`internal/backend/router/api-v3/node/agent/install.go:141` 读取策略，:219-224 对 `bkmonitorbeat` 调 `DecideCompatibilityMode` 设置 `EnableCompatibilityMode`。
- Proxy 安装：`internal/backend/router/api-v3/node/proxy/install.go:98,147,243-246`。
- 插件 Install / Upgrade / ApplySubConfig：`install.go:51,76,103,135-160` 中 `resolvePluginCompatibilityModePolicy` + `applyPluginCompatibilityModePolicy`。

## 8. 如何保证只有一台 bkm 上报基础数据

### 8.1 兼容模式策略（全局）

- **全局设置键**：`PluginCompatibilityModePolicy = "plugin_compatibility_mode_policy"` (`pkg/globalsettings/definition.go:27`)。
- **默认值**：`{"enabled_plugins":["bkmonitorbeat"],"disabled_biz":[]}`（:58-59），即**默认对 bkmonitorbeat 启用兼容模式**。
- **Policy 结构**：`internal/backend/manager/compatibility/policy.go:4-13`

  ```go
  type Policy struct {
      EnabledPlugins []string      `json:"enabled_plugins"`
      DisabledBiz    []DisabledBiz `json:"disabled_biz"`
  }
  type DisabledBiz struct {
      TenantID string `json:"tenant_id"`
      BKBizID  int64  `json:"bk_biz_id"`
  }
  ```

- **决策逻辑**：`internal/backend/manager/compatibility/decision.go:6-22` —— 插件名在 `enabled_plugins` 列表中、且 (tenant, biz) 不在 `disabled_biz` 中，则返回 true。
- **策略解析与回退**：`internal/backend/manager/compatibility/parse.go`，读取失败时回退到 `DefaultPolicy()`。

### 8.2 dataid 覆写机制（核心）

`internal/backend/manager/workflowdef/plugin/action_overwrite_plugin_config_for_compatibility.go`

- **正则**（:32）：`` `(?im)^([ \t]*[^#\r\n:]*dataid[^:\r\n]*:[ \t]*).*?([ \t]+#.*)?$` ``
  匹配配置文件中所有非注释、含 `dataid` 的键值行。
- **门控**（:107-114）：`EnableCompatibilityMode == false` 时跳过，日志 `"未启用兼容模式，跳过插件配置文件兼容性覆写操作"`。
- **覆写**（:124-128）：遍历所有 `ConfigFilesDetail`，对每个文件内容调用 `overwritePluginConfigContentForCompatibility`。
- **替换**（:158-160）：

  ```go
  func overwritePluginConfigContentForCompatibility(content string) string {
      return pluginConfigCompatibilityOverwriteRegex.ReplaceAllString(content, "${1}-1${2}")
  }
  ```

  把 `dataid:` 的值替换为 `-1`，保留行尾注释。
- **落库**（:130-148）：覆写后的内容写回 `PluginDeploymentPluginConf` 与 `ProcessConfig`，确保下发的配置就是 `-1`。

### 8.3 注册位置

该 action 注册在三条 V3 工作流，确保任意入口下 V3 配置都会被覆写：

- install：`oper_install_plugin.go:55`
- upgrade：`oper_upgrade_plugin.go:56`
- apply_subconfig：`oper_apply_plugin_subconfig.go:57`

### 8.4 效果

兼容模式开启时：

| 进程 | dataid | 上报行为 |
|------|--------|---------|
| V2 `bkmonitorbeat`（namespace=`nodeman`） | 原 dataid | 正常上报基础数据 |
| V3 `bkmonitorbeat`（namespace=`bk-nodemgr`） | `-1` | 静默，不上报基础数据 |

两个进程物理共存（GSE key 不同），但数据流只有 V2 一条活跃。V3 仍可承载子插件逻辑（子配置可正常下发），只是不上报到监控平台。

### 8.5 关闭某业务的兼容模式

通过 `disabled_biz` 可以按 (tenant, biz) 粒度关闭兼容模式，使其在该业务下不双装 V2、不覆写 dataid，从而让 V3 直接接管上报。这为灰度回退提供了业务级开关。

## 9. 子插件（子配置）迁移路径 V2 → V3

### 9.1 迁移 API

- **路由**：`POST /api/v3/plugin/migrate_from_v2`（`internal/backend/router/api-v3/plugin/plugin.go:64`）。
- **Handler**：`internal/backend/router/api-v3/plugin/migrate_from_v2.go:24-74` (`MigrateFromV2`)。
- **Proto 定义**：`proto/backend/api/v3/plugin.proto:195-206,277-284`（`PluginMigrateFromV2Req/Resp`、rpc `MigratePluginFromV2`）。
- **Manager 入口**：`internal/backend/manager/plugin_manager.go:476-519` (`LaunchMigrateFromV2`)，并发为每个 `PluginDeployment` 创建迁移操作实例（`createMigrateFromV2Oper`，:521）。
- **工作流类型**：`PluginWorkflowTypeMigrateV2 = "migrate_plugin_v2"`（`pkg/types/plugin_workflow.go:69-70`）。
- **Param 类型**：`MigrateFromV2Param`（`pkg/types/manager.go:128-135`）。
- **引入版本**：v3.0.1-alpha.32 (2026-05-28)，见 `support-files/changelog/zh/v3.0.1-alpha.32_2026-05-28.md:8,11`。

### 9.2 迁移工作流步骤

`internal/backend/manager/workflowdef/plugin/oper_migrate_from_plugin_process_v2.go:46-59`，action 链：

```
1. FetchPluginProcess              — 获取现有 V3 进程信息
2. InjectPluginCustomDeployConfig  — 注入自定义部署配置
3. VerifyPluginAvailability        — 验证插件可用性
4. RenderPluginDeployment          — 渲染部署信息
5. EnsureAndUpdatePluginConfigDetails — 确保并更新配置详情（含子配置）
6. RenderPluginConfig              — 渲染配置文件
7. PushPluginConfig                — 推送配置到节点
8. StopPluginV2Process             — 停止 V2 插件进程（关键步骤）
9. RestartProcess                  — 重启 V3 进程（关键步骤）
10. UpdateProcess                  — 更新进程信息
```

### 9.3 "先挂子插件到新 bkm，再下线旧 bkm"的实现

迁移工作流正是这个顺序：

1. **步骤 5–7**：先把新配置（含子配置）渲染并推送到 V3 `bkmonitorbeat`，此时 V3 已具备接管子插件的能力，但仍静默（dataid=-1 由 `OverwritePluginConfigForCompatibility` 处理）。
2. **步骤 8 `StopPluginV2Process`**：停止 V2 进程。这是一个**子工作流调用**：

   - `action_stop_plugin_v2_process.go:154-181` 中构造 `StopProcessParam{Type: PluginWorkflowTypeStopV2, ...}`，调用 `pluginMgrIface.LaunchStopPluginV2` 启动子工作流；
   - 子工作流定义：`pluginv2/oper_stop_plugin_v2.go:46-53`，链路 `InjectPluginBaseRuntimeV2 → FinishIfPluginProcessV2NotAlive → FetchPluginProcessV2 → StopProcessV2`；
   - 主工作流轮询等待子工作流完成（`action_stop_plugin_v2_process.go:199-239`）。
3. **步骤 9 `RestartProcess`**：重启 V3 进程，使其加载新配置并开始上报（此时 V2 已下线，V3 接管）。

> 注意：迁移工作流**不会重新传输插件包**。`migrate_from_v2.go:49` 使用 `PluginDeploymentTransferOptionsOnlyTransferInstaller()`（定义于 `pkg/types/plugin_deployment.go:221-226`），因为 V3 插件包在 install 阶段已经传输到位，迁移只需更新配置与进程状态。

### 9.4 迁移前置条件

- 主机上已存在 V3 `bkmonitorbeat`（由 install 或 Agent 安装时的预置插件流程部署）；
- 兼容模式曾开启（否则不会双装 V2）；
- V3 插件包已就位（迁移不传包）。

### 9.5 迁移后状态：V3 dataid 自动恢复

- V2 进程停止（namespace=`nodeman` 的 GSE 托管被清除）；
- V3 进程重启，加载迁移工作流**重新渲染**的配置；
- **V3 dataid 自动恢复为正常值，不是 -1**。验证依据：
  1. 迁移 handler `migrate_from_v2.go:24-74` **不调用** `resolvePluginCompatibilityModePolicy`（对比 `install.go:51` / `upgrade.go:48` / `apply_subconfig.go:48` 均调用），即迁移流程的 `EnableCompatibilityMode` 不被设置；
  2. 迁移工作流 `oper_migrate_from_plugin_process_v2.go:46-59` 的 action 链**不含** `OverwritePluginConfigForCompatibility`，不会把 dataid 改为 -1；
  3. 步骤 5 `EnsureAndUpdatePluginConfigDetails` 的 `fillConfigDetails` 会把 `ConfigFilesDetail[idx].Content = tpl.SourceContent`（`action_ensure_and_update_plugin_config_details.go:201`），即**重置为插件包模板原文**，覆盖掉 install 阶段被覆写的 `dataid=-1`；
  4. 步骤 6 `RenderPluginConfig`（`action_render_plugin_config.go:177`）对模板原文渲染，dataid 变为渲染后的正常值。

> 结论：`migrate_from_v2` 本身就是"恢复 dataid + 接管上报"的动作，迁移完成后 V3 即正常上报，**无需额外接口恢复 dataid**。

### 9.6 迁移不清理 V2 文件

迁移工作流只 `StopPluginV2Process`（停止 V2 进程、清除 GSE 托管），**不 uninstall V2**（不清理 V2 二进制、配置、`plugins/` 目录）。V2 残留文件需要额外手段清理，见 §10.1。

## 10. 运营手段：卸载 V2 bkm 与恢复 dataid

本节回答两个运维问题：① 迁移后或共存期如何卸载旧的 V2 `bkmonitorbeat`；② 如何让 V3 `bkmonitorbeat` 的 dataid 从 -1 恢复为正常值以接管上报。

### 10.1 卸载旧 V2 bkm：当前无独立 API

**现状：代码已实现 V2 uninstall 能力，但未暴露任何 API 入口。**

- `LaunchUninstallPluginV2`（`internal/backend/manager/plugin_v2_manager.go:128-129`，接口 `iface/iface.go:94-95`）已实现，但 grep 全仓**没有任何 router/handler 调用它**（仅定义与本文档引用）。
- `LaunchStopPluginV2`（`plugin_v2_manager.go:216-217`，`iface/iface.go:97-98`）**只被迁移工作流的子工作流调用**（`action_stop_plugin_v2_process.go:175`），同样无 API handler。
- `/uninstall`（`uninstall.go:54`）只调 V3 的 `LaunchUninstallPlugin`，不触及 V2。
- `/stop`（`plugin.go:65`）同样只作用于 V3。

**唯一能停 V2 的途径**：`POST /api/v3/plugin/migrate_from_v2`（`migrate_from_v2.go:55`），它通过 `StopPluginV2Process` 子工作流停掉 V2 进程并清除 GSE 托管（namespace=`nodeman` 的 key），但**不清理 V2 文件**（二进制、配置、`plugins/` 目录残留）。

**运营缺口**：若需在迁移后彻底卸载 V2 文件，当前无 API 可用。可选方案：
1. 后续为 `LaunchUninstallPluginV2` 补一个 API handler（实现已就绪，只缺路由暴露）；
2. 或手动通过 GSE / 脚本清理 `plugins/` 目录下 V2 残留。

### 10.2 恢复 V3 dataid：`POST /api/v3/plugin/upgrade`

若未走迁移路径，而是想在共存期直接让 V3 接管上报，需要把 V3 的 `dataid` 从 -1 恢复为正常值。恢复手段是**关闭兼容模式后触发一次会重新渲染配置的工作流**。

**接口**：`POST /api/v3/plugin/upgrade`（`internal/backend/router/api-v3/plugin/upgrade.go:23`）

**为什么是 upgrade**：
- upgrade handler（`upgrade.go:48`）调 `resolvePluginCompatibilityModePolicy` + `applyPluginCompatibilityModePolicy`，按全局策略设置 `EnableCompatibilityMode`；
- upgrade 工作流（`oper_upgrade_plugin.go:46-64`）含完整渲染链：`EnsureAndUpdatePluginConfigDetails`（:54，重载模板原文）→ `RenderPluginConfig`（:55，渲染）→ `OverwritePluginConfigForCompatibility`（:56，兼容模式关则跳过覆写）；
- 当 `EnableCompatibilityMode == false` 时，`OverwritePluginConfigForCompatibility` 在 `action_overwrite_plugin_config_for_compatibility.go:107-114` 直接 return，不把 dataid 改为 -1；
- 于是渲染后的正常 dataid 被保留并下发。

**同等可行的接口**：
- `POST /api/v3/plugin/apply_subconfig`（`apply_subconfig.go:48`）：同样走兼容策略 + 含 `OverwritePluginConfigForCompatibility`（`oper_apply_plugin_subconfig.go:56`），适合子配置变更场景顺带恢复 dataid。
- `POST /api/v3/plugin/install`（`install.go:53`）：重装即重新渲染，但语义偏重。

**不可行**：
- `POST /api/v3/plugin/restart`（`oper_restart_process.go:46-53`）：工作流只有 `FetchPluginProcess → RestartProcess → CheckPluginProcessAlive → UpdateProcess`，**不重新渲染配置**，无法改变 dataid。

### 10.3 推荐运营流程

**路径 A（推荐，走迁移）**：

1. 确认 V3 `bkmonitorbeat` 已通过 install 或 Agent 预置插件部署到位（共存期 dataid=-1，V2 上报）；
2. 调 `POST /api/v3/plugin/migrate_from_v2`：停 V2 进程 + 重新渲染 V3 配置 + 重启 V3。迁移完成即 V3 dataid 自动恢复、接管上报（见 §9.5）；
3. V2 文件残留按 §10.1 方案清理（当前需手动或补 API）；
4. 迁移稳定后，通过 Admin API（§10.4）把目标业务加入 `disabled_biz` 或从 `enabled_plugins` 移除 `bkmonitorbeat`，避免新装主机再双装 V2。

**路径 B（不走迁移，直接让 V3 接管）**：

1. 通过 Admin API（§10.4）关闭目标业务的兼容模式（加入 `disabled_biz`）；
2. 调 `POST /api/v3/plugin/upgrade` 对该业务主机触发 V3 升级工作流 → 重新渲染 + 跳过覆写 → V3 dataid 恢复正常、开始上报；
3. 此时 V2 与 V3 同时上报，需立即停 V2。但 §10.1 指出无独立 stop V2 的 API —— 只能调 `migrate_from_v2`（它会在停 V2 的同时再次渲染 V3，本次 dataid 仍会恢复，因为迁移工作流不含覆写 action）。**路径 B 实际会退化为路径 A**。

> 因此推荐路径 A：`migrate_from_v2` 一步到位（停 V2 + 恢复 V3 dataid），路径 B 因缺少独立停 V2 的 API 而不闭环。

### 10.4 兼容模式策略 Admin API

兼容模式策略可通过 Admin API 读写：

- **路由注册**：`internal/backend/router/admin/globalsettings/globalsettings.go:48-49`。
- **Handler**：`internal/backend/router/admin/globalsettings/plugin_compatibility_mode_policy.go:67,87`。
- **路径**：`/plugin_compatibility_mode_policy/get`、`/plugin_compatibility_mode_policy/upsert`。

典型用法：

- 全局关闭兼容模式：把 `enabled_plugins` 设为 `[]`，新装主机不再双装 V2。
- 按业务灰度：把目标业务加入 `disabled_biz`，该业务下 V3 直接接管上报。
- 迁移完成后：可逐步清空 `enabled_plugins` 或加入 `disabled_biz`，让 V3 恢复正常 dataid。

> 该 Admin API 暂无对外 apigw 文档（apigw/apidocs 下无对应文件），仅代码内定义。

## 11. 风险与注意事项

1. **迁移不清理 V2 文件**：`migrate_from_v2` 只 `StopPluginV2Process`（停进程 + 清 GSE 托管），不 uninstall V2 二进制/配置/`plugins/` 目录。当前无独立 API 卸载 V2（`LaunchUninstallPluginV2` 已实现但未暴露 handler，见 §10.1），迁移后 V2 文件残留需手动或补 API 清理。

2. **共存期 V3 不会自动恢复 dataid**：迁移之外的操作（install / upgrade / apply_subconfig）只要兼容模式仍开启，`OverwritePluginConfigForCompatibility` 就会把 dataid 改回 -1。若不走迁移而想让 V3 接管上报，必须先通过 Admin API（§10.4）关闭兼容模式，再触发 upgrade/apply_subconfig 重新渲染（见 §10.2）。`/restart` 不会重新渲染，无法恢复 dataid。

3. **dataid 覆写正则的覆盖范围**：正则匹配所有含 `dataid` 的非注释行（`action_overwrite_plugin_config_for_compatibility.go:32`）。若插件配置中有非上报用途的 `dataid` 字段（如某个子模块的 id），也会被改成 -1。需确认 bkm 配置模板中所有 `dataid` 行确实都代表上报开关。

4. **双安装的资源开销**：共存期主机上运行两个 bkm 进程，CPU/内存/PID 占用翻倍。Proxy 角色还会多装 `bk-nodemgr-relay`（V3，:307）。短期可接受，长期应尽快完成迁移。

5. **离线模式**：`installPreOrderedPluginV2` 同样支持离线安装（:341-346），离线环境下 V2/V3 双装仍会发生，需确保离线包内同时包含 V2 与 V3 插件包。

6. **GSE namespace 不可混用**：V2 必须用 `nodeman`，V3 必须用 `bk-nodemgr`。若手动运维时误传 namespace，会导致 GSE key 冲突或托管失效。

7. **Admin API 无对外文档**：`/plugin_compatibility_mode_policy/*` 仅代码定义，操作前需参考 `internal/backend/manager/compatibility/policy.go` 的结构。

## 12. 关键文件索引

按主题分组，便于精读：

### 兼容模式与 dataid 覆写
- `pkg/globalsettings/definition.go:27,58-59` — 全局策略键与默认值
- `internal/backend/manager/compatibility/policy.go:4-21` — Policy 结构与默认策略
- `internal/backend/manager/compatibility/decision.go:6-32` — 决策逻辑
- `internal/backend/manager/compatibility/parse.go` — 策略解析与回退
- `internal/backend/manager/workflowdef/plugin/action_overwrite_plugin_config_for_compatibility.go:32,107-160` — dataid 覆写核心

### 双安装触发点
- `internal/backend/manager/workflowdef/node/action_install_pre_ordered_plugins.go:125-195,303-308,311-323,402-407` — Agent/Proxy 预置插件双装
- `internal/backend/router/api-v3/plugin/install.go:28-69,99-133` — 插件 Install API 双装
- `internal/backend/router/api-v3/node/agent/install.go:141,219-224` — Agent 安装时兼容模式判定
- `internal/backend/router/api-v3/node/proxy/install.go:98,147,243-246` — Proxy 安装时兼容模式判定

### V2 工作流
- `internal/backend/manager/plugin_v2_manager.go:128-129,216-217` — V2 Manager 入口（`LaunchUninstallPluginV2` / `LaunchStopPluginV2`，已实现但 uninstall 无 API 调用方）
- `internal/backend/manager/iface/iface.go:89-99` — V2 子接口定义
- `internal/backend/manager/workflowdef/pluginv2/pluginv2.go:11,48` — V2 包注释与 namespace
- `internal/backend/manager/workflowdef/pluginv2/oper_plugin_ensure_plugin_v2.go:46-59` — V2 ensure 工作流
- `internal/backend/manager/workflowdef/pluginv2/action_inject_plugin_base_runtime_v2.go:214-248` — V2 目录与 subconfig 生成
- `internal/backend/manager/workflowdef/pluginv2/action_ensure_and_update_plugin_config_details_v2.go:429,538` — V2 subconfig_path

### V3 工作流与配置渲染
- `internal/backend/manager/plugin_manager.go:476-519` — V3 Manager 入口（含 `LaunchMigrateFromV2`）
- `internal/backend/manager/workflowdef/plugin/oper_install_plugin.go:46-62` — V3 install 工作流
- `internal/backend/manager/workflowdef/plugin/oper_upgrade_plugin.go:46-64` — V3 upgrade 工作流（恢复 dataid 的入口）
- `internal/backend/manager/workflowdef/plugin/oper_restart_process.go:46-53` — V3 restart 工作流（不渲染配置，不能恢复 dataid）
- `internal/backend/manager/workflowdef/plugin/action_ensure_and_update_plugin_config_details.go:161,201` — `fillConfigDetails` 重置 Content 为模板原文（迁移后 dataid 恢复的关键）
- `internal/backend/manager/workflowdef/plugin/action_render_plugin_config.go:150-194` — 模板渲染逻辑
- `internal/backend/manager/workflowdef/plugin/action_trusteeship_process_to_gse.go:35-37` — V3 GSE 托管（默认 namespace）
- `pkg/deployconstant/plugin_deploy_constant.go:26,145-147` — V3 目录与 SubConfigDir
- `pkg/thirdparty/gse/handler_proc.go:67,145` — GSE namespace 常量与进程 key 格式

### 迁移工作流
- `internal/backend/manager/workflowdef/plugin/oper_migrate_from_plugin_process_v2.go:46-59` — 迁移工作流定义（不含 OverwritePluginConfigForCompatibility）
- `internal/backend/manager/workflowdef/plugin/action_stop_plugin_v2_process.go:154-239` — 停 V2 子工作流调用与等待
- `internal/backend/manager/workflowdef/pluginv2/oper_stop_plugin_v2.go:46-53` — V2 stop 子工作流
- `proto/backend/api/v3/plugin.proto:195-206,277-284` — 迁移 proto 定义
- `pkg/types/plugin_workflow.go:69-70` — 迁移工作流类型常量
- `pkg/types/manager.go:128-135` — `MigrateFromV2Param`

### 运营接口（卸载 V2 / 恢复 dataid / 迁移）
- `internal/backend/router/api-v3/plugin/plugin.go:56-65` — plugin 路由注册（install/upgrade/uninstall/apply_subconfig/restart/migrate_from_v2/stop）
- `internal/backend/router/api-v3/plugin/migrate_from_v2.go:24-74` — 迁移 API handler（唯一能停 V2 的入口，不含兼容策略解析）
- `internal/backend/router/api-v3/plugin/upgrade.go:23-75` — upgrade handler（恢复 dataid 的主入口，走兼容策略）
- `internal/backend/router/api-v3/plugin/apply_subconfig.go:23-75` — apply_subconfig handler（同走兼容策略）
- `internal/backend/router/api-v3/plugin/restart.go:23-74` — restart handler（不渲染配置）
- `internal/backend/router/api-v3/plugin/uninstall.go:23-73` — uninstall handler（仅 V3）

### Admin API
- `internal/backend/router/admin/globalsettings/globalsettings.go:48-49` — 路由注册
- `internal/backend/router/admin/globalsettings/plugin_compatibility_mode_policy.go:67,87` — Handler

### 类型与常量
- `pkg/types/release.go:33-40` — ReleaseType（V2/V3 包类型）
- `pkg/types/plugin_deployment.go:22-26,221-226` — PluginDeployment 类型与 TransferOptions
- `pkg/types/plugin_workflow.go:21-33` — PluginWorkflow 类型
- `pkg/types/generation.go` — NodeGeneration 定义
- `docs/developer/plugin_pkg_build_guide.md` — V3 插件包规范（唯一权威文档）

### Changelog
- `support-files/changelog/zh/v3.0.1-alpha.23_2026-05-11.md:9,16,17` — 兼容模式与配置覆写引入
- `support-files/changelog/zh/v3.0.1-alpha.32_2026-05-28.md:8,11` — 迁移 API 与 stop plugin v2 引入
