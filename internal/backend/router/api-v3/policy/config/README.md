# config

`config` 路由组统一挂载在 `/policy/config` 前缀下，用于承载配置策略（Config Policy）的生命周期管理和事件查询相关 API。

入口在 `config.go`，本身只负责定义 handler 结构体、构造函数和路由注册，不承载具体业务逻辑。

## 文件结构

| 文件 | 职责 |
| --- | --- |
| `config.go` | 主入口：handler 结构体、newHandler 构造函数、Load 路由注册 |
| `list.go` | ListConfigPolicy - 列表查询配置策略 |
| `get.go` | GetConfigPolicy - 获取单个配置策略详情 |
| `create.go` | CreateConfigPolicy - 创建配置策略 |
| `update.go` | UpdateConfigPolicy - 更新配置策略 |
| `enable.go` | EnableConfigPolicy - 启用配置策略 |
| `disable.go` | DisableConfigPolicy - 禁用配置策略 |
| `delete.go` | DeleteConfigPolicy - 删除配置策略 |
| `reorder.go` | ReorderPrioritiesConfigPolicy - 重排配置策略优先级 |
| `preview.go` | PreviewConfigPolicy - 预览配置策略匹配结果 |
| `event.go` | ListConfigPolicyEvent、DistinctConfigPolicyEvent - 配置策略事件查询 |
| `event_record.go` | 事件记录辅助函数（异步记录配置策略操作事件） |
| `helpers.go` | 通用辅助函数（getConfigPolicy 等） |

## 组织原则

- **按功能拆分**：每个文件对应一个或一组相关的 handler 方法
- **职责单一**：每个文件只负责特定的业务功能
- **辅助函数归类**：
  - 事件记录相关 → `event_record.go`
  - 通用查询辅助 → `helpers.go`
  - 业务逻辑辅助（如 buildReorderedPolicyIDs）→ 放在使用它的文件中（`reorder.go`）
  - 特定操作辅助（如 getEnableConfigPolicyBizIDs）→ 放在使用它的文件中（`enable.go`）

## 参考模式

本包的组织方式参考了 `internal/backend/router/api-v3/node/agent` 的结构：
- 主入口文件定义 handler 和路由
- 按功能独立文件（install.go、upgrade.go、restart.go 等）
- 每个文件包含相关的 handler 方法和辅助函数
