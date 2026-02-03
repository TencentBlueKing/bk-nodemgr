## ADDED Requirements

### Requirement: IAM v3 API 客户端层

系统 SHALL 在 `iamv3.go` 中实现 IAM v3 的 REST API 客户端，使用项目标准的 `restclient.IClient`。

#### Scenario: API 基础配置

- **WHEN** 初始化 IAM v3 客户端
- **THEN** 系统 SHALL 使用 `/api` 作为 baseURL，具体 API 版本（v1/v2）由方法自行拼接

#### Scenario: v2 策略查询

- **WHEN** 调用 `v2PolicyQuery` 方法
- **THEN** 系统 SHALL 发送 POST 请求到 `/v2/policy/systems/{system}/query/`
- **AND** 返回策略表达式数据

#### Scenario: v2 按操作查询策略

- **WHEN** 调用 `v2PolicyQueryByActions` 方法
- **THEN** 系统 SHALL 发送 POST 请求到 `/v2/policy/systems/{system}/query_by_actions/`
- **AND** 返回多个操作的策略列表

#### Scenario: 获取系统 Token

- **WHEN** 调用 `getToken` 方法
- **THEN** 系统 SHALL 发送 GET 请求到 `/v1/model/systems/{system}/token`
- **AND** 从响应中提取 token 字符串

#### Scenario: 获取申请 URL

- **WHEN** 调用 `getApplyURL` 方法
- **THEN** 系统 SHALL 发送 POST 请求到 `/v1/open/application/`
- **AND** 从响应中提取 url 字符串

### Requirement: 响应处理

系统 SHALL 使用泛型 `BaseBroker[T]` 结构统一处理 IAM API 响应。

#### Scenario: 成功响应

- **WHEN** IAM API 返回 `code: 0`
- **THEN** `BaseBroker.IsFailed()` SHALL 返回 nil
- **AND** 响应数据解析到 `Data` 字段

#### Scenario: 错误响应

- **WHEN** IAM API 返回非零 `code`
- **THEN** `BaseBroker.IsFailed()` SHALL 返回包含 code 和 message 的错误

### Requirement: 请求头处理

系统 SHALL 在每个 API 请求中设置必要的请求头。

#### Scenario: IAM 版本头

- **WHEN** 发送 IAM API 请求
- **THEN** 请求 SHALL 包含 `X-Bk-IAM-Version: 1` 头

#### Scenario: 租户 ID 头

- **WHEN** 上下文包含租户 ID
- **THEN** 请求 SHALL 包含 `X-Bk-Tenant-Id` 头

## ADDED Requirements

### Requirement: IAM v3 业务逻辑层

系统 SHALL 在 `handler.go` 中实现 IAM v3 的业务逻辑方法。

#### Scenario: 权限判断 - IsAllowed

- **WHEN** 调用 `IsAllowed(ctx, request)` 方法
- **THEN** 系统 SHALL 调用 `v2PolicyQuery` 获取策略
- **AND** 使用 `expression.ExprCell` 计算策略表达式
- **AND** 返回布尔值表示是否允许

#### Scenario: 带缓存的权限判断 - IsAllowedWithCache

- **WHEN** 调用 `IsAllowedWithCache(ctx, request, ttl)` 方法
- **THEN** 系统 SHALL 先检查缓存是否命中
- **AND** 缓存未命中时调用 `IsAllowed` 并缓存结果

#### Scenario: 批量权限判断 - BatchIsAllowed

- **WHEN** 调用 `BatchIsAllowed(ctx, request, resourcesList)` 方法
- **THEN** 系统 SHALL 一次查询策略，对多个资源分别计算
- **AND** 返回 `map[resourceID]bool` 结果

#### Scenario: 基础认证校验 - IsBasicAuthAllowed

- **WHEN** 调用 `IsBasicAuthAllowed(ctx, username, password)` 方法
- **THEN** 系统 SHALL 验证 username 为 "bk_iam"
- **AND** 验证 password 与系统 token 匹配

### Requirement: 本地缓存

系统 SHALL 使用 go-cache 实现本地内存缓存。

#### Scenario: 缓存初始化

- **WHEN** 初始化 Handler
- **THEN** 系统 SHALL 创建 go-cache 实例，设置合理的默认过期时间和清理间隔

#### Scenario: 缓存键生成

- **WHEN** 缓存权限查询结果
- **THEN** 系统 SHALL 基于 request 的内容生成唯一缓存键

### Requirement: NoOp 实现

系统 SHALL 更新 `NoOpHandler` 实现所有新增的 `IHandler` 接口方法。

#### Scenario: NoOp 权限判断

- **WHEN** 调用 `NoOpHandler.IsAllowed()`
- **THEN** 系统 SHALL 记录 debug 日志并返回 `true, nil`

#### Scenario: NoOp 其他方法

- **WHEN** 调用 `NoOpHandler` 的任意业务方法
- **THEN** 系统 SHALL 记录 debug 日志并返回成功结果
