# Router 层

## 设计意图

Router 层负责管理所有 Mock 组件的路由注册，采用组件隔离设计，支持不同第三方服务的独立响应格式。

## 目录结构

```
router/
├── router.go          # 主入口，统一注册所有组件
├── common/            # 公共辅助函数
│   └── common.go      # 不同组件的通用函数
├── cmdb/              # CMDB Mock 组件
│   ├── router.go      # Load 函数和 handler
│   ├── helper.go      # CMDB 专属响应格式
│   ├── types.go       # Config、常量、错误码
│   ├── storage.go     # 存储
│   └── business.go    # 业务处理
└── healthz/           # 服务健康检查
```

## 路由前缀规范

所有 Mock API 路由必须遵循统一的前缀格式：`${mock-server-name}/xxxx`

- `${mock-server-name}` 是 Mock 服务的组件名称（如 `cmdb`、`bkrepo`）
- `/xxxx` 是组件内部的具体 API 路径

示例：`/cmdb/api/v3/biz/search/:supplier_account`

## 设计原则

- **组件隔离**：每个组件（cmdb、bkrepo 等）完全独立，响应格式互不影响
- **公共基础**：`common/` 提供通用的组件处理函数, 如 bind json
- **统一注册**：所有组件通过 `router.go` 的 `Load` 函数统一注册
- **路由前缀**：所有组件路由必须遵循 `${mock-server-name}/xxxx` 格式

## 组件结构规范

每个组件目录应包含以下文件（可根据实际需求调整）：

- `router.go` - Load 函数和 handler 结构体定义
- `helper.go` - 组件专属的响应格式封装
- `types.go` - Config、常量、错误码定义
- `storage.go` - 存储实现
- `handlers.go` - 业务处理函数

## 扩展新组件

### 1. 创建组件目录

在 `router/` 下创建新目录，例如 `bkrepo/`

### 2. 实现 Load 函数

遵循统一签名：`Load(rg *gin.RouterGroup, conf *Config) error`

**重要**：必须使用组件名称创建路由组，确保路由前缀遵循 `${mock-server-name}/xxxx` 格式。

### 3. 定义响应格式

在 `helper.go` 中实现 `respondOK` 和 `respondError`，根据该组件的实际 API 响应格式封装。

**重要**：每个组件的响应格式完全独立，不要复用其他组件的 helper。

### 4. 注册到主路由

在 `router/router.go` 的 `Load` 函数中导入并调用新组件的 `Load` 函数。

### 5. 添加预设数据支持（可选，如果有）

在 `config.go` 的 `MockData` 中添加新字段，并在 `Validate()` 中添加验证。

## 注意事项

1. 所有路由必须通过 `router.go` 的 `Load` 函数统一注册
2. 组件间响应格式完全独立，互不影响
3. 使用 `common.BindJSON` 进行请求绑定，使用 `common.RespondJSON` 发送响应
4. 同一组件内的响应格式需保持一致
5. 所有组件路由必须遵循 `${mock-server-name}/xxxx` 前缀格式

## 参考实现

查看 `cmdb/` 目录作为参考实现。
