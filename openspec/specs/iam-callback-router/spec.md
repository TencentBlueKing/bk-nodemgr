## ADDED Requirements

### Requirement: IAM 路由组注册

系统 SHALL 在 `internal/backend/router/api-v3/iam/` 目录下创建独立的 IAM 路由模块，提供 `Load` 函数用于注册路由到 `/api/v3/iam/v3` 路径。

#### Scenario: 路由组正确挂载

- **WHEN** 调用 `iam.Load(rg, capability)` 注册路由
- **THEN** 系统在 `/api/v3/iam/v3` 路径下创建路由组

#### Scenario: 遵循项目路由模式

- **WHEN** 实现 Load 函数
- **THEN** 函数签名 MUST 为 `Load(rg *gin.RouterGroup, capability *options.Capability)`，与其他路由模块（如 callback、proxy）保持一致

### Requirement: IAM 资源回调端点

系统 SHALL 在 `/api/v3/iam/v3/resource` 路径提供 POST 端点，用于接收 IAM 系统的资源回调请求。

#### Scenario: 接收资源回调请求

- **WHEN** IAM 系统发送 POST 请求到 `/api/v3/iam/v3/resource`
- **THEN** 系统 MUST 使用 `resource.NewDispatchHandler(d)` 创建的 handler 处理请求

#### Scenario: 请求体格式

- **WHEN** 接收到回调请求
- **THEN** 请求体 MUST 包含 `type`（资源类型）和 `method`（操作方法）字段

### Requirement: Dispatcher 和 Provider 模式

系统 SHALL 使用 `iam-go-sdk` 的 Dispatcher/Provider 模式处理不同资源类型的回调请求。

#### Scenario: Dispatcher 初始化

- **WHEN** Load 函数执行时
- **THEN** 系统 MUST 创建 Dispatcher 实例：`d := resource.NewDispatcher()`

#### Scenario: Provider 注册

- **WHEN** Dispatcher 创建后
- **THEN** 系统 MUST 通过 `d.RegisterProvider(type, provider)` 注册各资源类型的 Provider

#### Scenario: 请求分发

- **WHEN** 收到资源回调请求且 `type` 字段对应的 Provider 已注册
- **THEN** Dispatcher MUST 根据 `method` 字段调用 Provider 的对应方法（如 `list_instance`、`fetch_instance_info` 等）

#### Scenario: 未注册的资源类型

- **WHEN** 收到资源回调请求但 `type` 字段对应的 Provider 未注册
- **THEN** 系统 MUST 返回 404 错误，消息中包含未支持的资源类型信息

### Requirement: Basic Auth 认证

系统 SHALL 对 IAM 资源回调端点实施 Basic Auth 认证，验证请求来源的合法性。

#### Scenario: 认证成功

- **WHEN** 请求携带正确的 Basic Auth 凭据（username=system_id, password=token）
- **THEN** 系统 MUST 允许请求继续处理

#### Scenario: 认证失败 - 无凭据

- **WHEN** 请求未携带 Basic Auth 凭据
- **THEN** 系统 MUST 返回 401 Unauthorized 错误

#### Scenario: 认证失败 - 凭据错误

- **WHEN** 请求携带的 Basic Auth 凭据与 IAM Token 不匹配
- **THEN** 系统 MUST 返回 401 Unauthorized 错误

#### Scenario: 使用现有认证方法

- **WHEN** 验证 Basic Auth 凭据
- **THEN** 系统 MUST 调用 `capability.IAMV3Handler.IsBasicAuthAllowed(ctx, username, password)` 进行验证

### Requirement: Handler 适配 Gin 框架

系统 SHALL 将 `iam-go-sdk` 的 `http.HandlerFunc` 适配到 Gin 框架。

#### Scenario: 使用 gin.WrapF 适配

- **WHEN** 注册资源回调路由
- **THEN** 系统 MUST 使用 `gin.WrapF(handler)` 将 `resource.NewDispatchHandler(d)` 返回的 handler 适配到 Gin

#### Scenario: HTTP 方法

- **WHEN** 注册资源回调路由
- **THEN** 路由 MUST 使用 POST 方法：`rg.POST("/resource", gin.WrapF(handler))`
