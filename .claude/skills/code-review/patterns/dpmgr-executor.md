# Executor 模式审查指南

## 适用范围

本文档专门针对 `internal/backend/dpmgr/executor.go` 文件中的 executor 函数模式审查。

## Executor 函数模式

该文件中的 executor 函数遵循特定的对称模式：

### Agent 系列函数
- `executeChangeActionAgentInstall`
- `executeChangeActionAgentUninstall`
- `executeChangeActionAgentUpgrade`

### Plugin 系列函数
- `executeChangeActionPluginInstall`
- `executeChangeActionPluginUninstall`
- `executeChangeActionPluginUpgrade`

### PluginPkg 系列函数
- `executeChangeActionPluginPkgInstall`
- `executeChangeActionPluginPkgUpgrade`
- `executeChangeActionPluginPkgUninstall`

## 审查检查清单

### 1. Plugin 系列函数对比

| 函数 | InstallOptions | CustomConfigContext | PluginConf | Manager 调用 |
|------|---------------|---------------------|------------|-------------|
| **Plugin Install** | ✅ 有 Version | ✅ 有 | ✅ 有 | `LaunchInstallPlugin` |
| **Plugin Uninstall** | ❌ 无 | ❌ 无 | ✅ 空结构体 | `LaunchUninstallPlugin` |
| **Plugin Upgrade** | ✅ 有 Version | ✅ 有 | ✅ 有 | `LaunchUpgradePlugin` |

**检查项：**
- Uninstall 函数不应该有 `InstallOptions` 和 `CustomConfigContext`
- Upgrade 函数应该有 `InstallOptions`（包含 Version）和 `CustomConfigContext`
- 所有函数都应该使用 `PluginDeploymentPluginConf`，Uninstall 使用空结构体

### 2. PluginPkg 系列函数对比

| 函数 | Upsert Plugins | InstallOptions | CustomConfigContext | Manager 调用 |
|------|---------------|---------------|---------------------|-------------|
| **PluginPkg Install** | ✅ 需要 | ✅ 有 Version | ✅ 有 | `LaunchInstallPlugin` |
| **PluginPkg Upgrade** | ❌ 不需要 | ✅ 有 Version | ✅ 有 | `LaunchUpgradePlugin` |
| **PluginPkg Uninstall** | ❌ 不需要 | ❌ 无 | ❌ 无 | `LaunchUninstallPlugin` |

**检查项：**
- Install 函数需要先创建 `plugins` 数组，然后调用 `UpsertManyPlugins`
- Upgrade 和 Uninstall 函数不需要 upsert（插件已存在）
- Upgrade 函数应该有 `InstallOptions`（包含 Version）和 `CustomConfigContext`
- Uninstall 函数不应该有 `InstallOptions` 和 `CustomConfigContext`

### 3. 日志格式一致性

**标准格式（两行）：**
```go
logger.G.Sys().With("workflow-id", workflowID).
    Info("successful to execute change action ...")
```

**检查项：**
- 所有 executor 函数的成功日志都应该使用两行格式
- `With` 和 `Info` 应该分开在不同行
- 日志消息格式：`"successful to execute change action {action name}"`

### 4. 错误消息格式一致性

**Plugin 系列错误消息：**
- `GetSpecifyPluginParam` 错误：`"failed to schedule and execute change action: %w"`
- Manager 调用错误：`"failed to execute change action plugin {action}: %w"`

**PluginPkg 系列错误消息：**
- `GetSpecifyPluginPkgParam` 错误：`"failed to get specify plugin pkg param for task: %w"`
- Manager 调用错误：`"failed to execute change action plugin pkg {action}: %w"`

### 5. 数据结构构建模式

**Plugin 系列：**
```go
pluginDeployments := make([]*types.PluginDeployment, len(tasks))
hostMap := make(map[int64]struct{})
for idx, task := range tasks {
    param, err := task.Spec.GetSpecifyPluginParam()
    if err != nil {
        return fmt.Errorf("failed to schedule and execute change action: %w", err)
    }
    
    pluginDeployments[idx] = types.NewPluginDeployment(...)
    hostMap[task.Target.Host.HostID] = struct{}{}
}

hostIDs := conv.MapKeyToSlice(hostMap)
workflowID, err := executor.pluginManager.Launch{Action}Plugin(...)
```

**PluginPkg 系列：**
```go
// Install 需要创建 plugins
plugins := make([]*types.Plugin, len(tasks))
pluginDeployments := make([]*types.PluginDeployment, len(tasks))
hostMap := make(map[int64]struct{})
for idx, task := range tasks {
    param, err := task.Spec.GetSpecifyPluginPkgParam()
    if err != nil {
        return fmt.Errorf("failed to get specify plugin pkg param for task: %w", err)
    }
    
    plugins[idx] = &types.Plugin{...}
    pluginDeployments[idx] = types.NewPluginDeployment(...)
    hostMap[task.Target.Host.HostID] = struct{}{}
}

// Install 需要 upsert
if err := executor.daoPlugin.UpsertManyPlugins(nCtx, plugins...); err != nil {
    return fmt.Errorf("failed to upsert plugins: %w", err)
}

hostIDs := conv.MapKeyToSlice(hostMap)
workflowID, err := executor.pluginManager.Launch{Action}Plugin(...)
```

## 常见问题

### 问题 1: 日志格式不一致
**症状：** 新增函数使用一行日志格式，而其他函数使用两行格式
**解决：** 统一为两行格式

### 问题 2: PluginPkg Upgrade/Uninstall 错误地 upsert plugins
**症状：** Upgrade 或 Uninstall 函数中包含了 `UpsertManyPlugins` 调用
**解决：** 移除 upsert 调用，因为插件在 Install 时已经创建

### 问题 3: Uninstall 函数包含 InstallOptions
**症状：** Uninstall 函数错误地包含了 `InstallOptions` 字段
**解决：** Uninstall 不应该有 InstallOptions，应该使用空的 `PluginDeploymentPluginConf`

### 问题 4: 错误消息格式不一致
**症状：** 不同系列函数使用不同的错误消息格式
**解决：** 遵循各自系列的错误消息模式

## 审查步骤

当审查 `executor.go` 中的新增或修改函数时：

1. **识别函数系列**：确定是 Plugin 还是 PluginPkg 系列
2. **找到对称函数**：找到同系列的其他函数（Install/Uninstall/Upgrade）
3. **对比数据结构**：检查 PluginDeployment 的构建是否一致
4. **检查 Manager 调用**：验证调用的 Manager 方法是否正确
5. **验证日志格式**：确保使用标准的两行格式
6. **检查错误处理**：确认错误消息格式符合系列规范
7. **特殊检查**：
   - PluginPkg Install 是否包含 upsert
   - PluginPkg Upgrade/Uninstall 是否错误地包含 upsert
   - Uninstall 函数是否错误地包含 InstallOptions

## 示例审查

### 示例：新增 Plugin Uninstall 函数

**检查项：**
- ✅ 没有 InstallOptions
- ✅ PluginConf 使用空结构体
- ✅ 使用 `LaunchUninstallPlugin`
- ✅ 日志格式为两行
- ✅ 错误消息格式正确

### 示例：新增 PluginPkg Upgrade 函数

**检查项：**
- ✅ 没有 upsert plugins（插件已存在）
- ✅ 有 InstallOptions（包含 Version）
- ✅ 有 CustomConfigContext
- ✅ 使用 `LaunchUpgradePlugin`
- ✅ 日志格式为两行
- ✅ 错误消息格式正确
