# notice-integration Specification

## Purpose

集成蓝鲸通知中心（BK-Notice）API，为节点管理平台提供公告获取功能。该模块遵循项目现有的 thirdparty Handler 模式，通过 API 网关与通知中心交互，支持多语言公告内容，并在 Handler 内部自动完成应用注册。

## Requirements

### Requirement: Notice Client Interface

系统 **SHALL** 提供蓝鲸通知中心 API 的客户端接口，支持获取当前公告。

#### Scenario: 获取当前公告列表

当调用 `GetCurrentAnnouncements` 方法时：
- **MUST** 发送 GET 请求到 `/apigw/v1/announcement/get_current_announcements`
- **MUST** 返回当前有效的公告列表
- **MUST** 使用配置中的 `appCode` 作为平台标识（platform 参数）
- **MUST** 使用 `contextx.IContext.BKUsername()` 作为用户名参数
- **MUST** 返回空列表（而非 nil）当没有活动公告时

#### Scenario: 处理 API 网关响应错误

当 API 网关返回非 0 错误码时：
- **MUST** 返回包含错误码和错误消息的 error
- **SHALL** 使用 `BaseBroker[T].IsFailed()` 检查响应状态
- **MUST** 错误信息格式为 `"code(%d), msg(%s)"`

---

### Requirement: Handler Pattern Compliance

模块 **MUST** 遵循项目现有的 thirdparty Handler 模式（参考 `pkg/thirdparty/cmdb/handler.go`）。

#### Scenario: 初始化 Handler

使用 `New(c *restclient.Capability, conf *Config)` 创建 Handler：
- **MUST** 验证配置有效性（调用 `conf.Validate()`）
- **MUST** 初始化 API 网关客户端（调用 `newClient`）
- **MUST** 初始化内部 scheduler 并启动应用注册任务
- **MUST** 返回实现 IHandler 接口的实例
- **MUST** 返回 error 当配置无效或客户端初始化失败时
- **SHALL** 在 scheduler 初始化失败时记录警告但继续返回 Handler

#### Scenario: Handler 方法签名

Handler 方法 **MUST** 遵循以下签名模式：
- **MUST** 第一个参数为 `contextx.IContext`
- **MUST** 返回值包含 error 作为最后一个返回值
- **SHALL** 使用指针类型作为请求/响应参数

---

### Requirement: Internal Application Registration

模块 **MUST** 在 Handler 内部自动完成应用注册，无需对外暴露注册接口。

#### Scenario: 自动注册机制

Handler 创建时：
- **MUST** 在 `New()` 中初始化内部 scheduler
- **MUST** 注册定时任务，周期为每天一次（`scheduler.Daily`）
- **MUST** 在 Handler 创建后立即执行一次注册
- **MUST** 启动 scheduler 进行周期性注册

#### Scenario: 注册 API 调用

内部注册方法：
- **MUST** 发送 POST 请求到 `/apigw/v1/register`
- **MUST** 通过 header 认证，无需请求体
- **SHALL** 注册失败时记录警告日志
- **MUST** 注册失败不阻止 scheduler 继续运行

#### Scenario: 错误处理

注册相关错误处理：
- **MUST** scheduler 初始化失败时记录警告，不阻止 Handler 创建
- **MUST** 立即执行注册失败时记录警告，不阻止 Handler 创建
- **MUST** 定时注册失败时记录警告，返回 nil 允许下次重试

---

### Requirement: Client Implementation

模块 **MUST** 实现私有 `cli` 客户端结构体。

#### Scenario: 客户端结构体

`cli` 结构体 **MUST** 包含：
- `client restclient.IClient` - 复用 apigw client

#### Scenario: newClient 构造函数

`newClient(c *restclient.Capability, conf *Config)` **MUST**：
- 调用 `conf.Validate()` 验证配置
- 调用 `apigwclient.NewClient(c, "/apigw/v1", conf.APIGWUserConfig)` 创建 REST 客户端
- 返回 `(*cli, error)`

#### Scenario: getHeader 方法

`getHeader()` **MUST**：
- 返回 `(http.Header, error)`
- **SHALL** 支持扩展租户信息

---

### Requirement: Configuration

模块 **SHALL** 支持通过配置管理 API 网关连接参数。

#### Scenario: 配置结构

Config 结构体 **MUST** 包含：
- `APIGWUserConfig apigwclient.UserConfig` - API 网关用户配置

调用 `Validate()` 方法：
- **MUST** 调用 `APIGWUserConfig.Validate()` 验证子配置
- **MUST** 返回格式化的错误信息 `"failed to validate notice client config: %v"`

#### Scenario: 固定注册配置

注册相关配置 **MUST** 使用固定常量：
- 注册周期：`scheduler.Daily`（每天一次）
- 注册超时：`time.Minute`（1分钟）
- **SHALL NOT** 提供可配置的注册参数

#### Scenario: AppConfig GetAppCode 方法

`apigwclient.AppConfig` **MUST** 提供 `GetAppCode()` 方法：
- **MUST** 返回配置中的 `appCode` 字段值
- **MUST** 用于 Notice Handler 获取 platform 参数

---

### Requirement: Response Handling

模块 **MUST** 使用泛型响应结构处理 API 响应。

#### Scenario: BaseBroker 泛型结构

`BaseBroker[T]` **MUST** 包含：
- `Code int` - 响应码（json tag: `"code"`）
- `Message string` - 响应消息（json tag: `"message"`）
- `Data T` - 响应数据（json tag: `"data"`）

#### Scenario: IsFailed 错误检查

`BaseBroker[T].IsFailed()` **MUST**：
- 返回 `nil` 当 `Code == 0`
- 返回 `error` 当 `Code != 0`，格式为 `"code(%d), msg(%s)"`

---

### Requirement: Internal Data Types

模块 **MUST** 定义私有数据传输对象（DTO）用于 API 响应解析。

#### Scenario: AnnouncementData 公告数据（内部类型）

`AnnouncementData` **MUST** 包含（不导出，匹配 API 响应）：
- `ID int64` - 公告 ID（json tag: `"id"`）
- `Content AnnouncementContent` - 公告内容（json tag: `"content"`）
- `AnnounceType string` - 公告类型（json tag: `"announce_type"`）
- `StartTime string` - 开始时间（json tag: `"start_time"`）
- `EndTime string` - 结束时间（json tag: `"end_time"`）

#### Scenario: AnnouncementContent 公告内容（内部类型）

`AnnouncementContent` **MUST** 包含：
- `Title string` - 标题（json tag: `"title"`）
- `Content string` - 内容（json tag: `"content"`）

#### Scenario: appRegistration 注册结果（内部类型）

`appRegistration` **MUST** 包含（小写不导出）：
- `ID int64` - 应用 ID
- `Code string` - 应用代码
- `Name string` - 应用名称

#### Scenario: RegisterApplicationResp 注册响应（内部类型）

`RegisterApplicationResp` **MUST** 包含：
- `ID int64` - 应用 ID（json tag: `"id"`）
- `Code string` - 应用代码（json tag: `"code"`）
- `Name string` - 应用名称（json tag: `"name"`）

---

### Requirement: Public Data Types

模块 **SHALL** 在 `pkg/types/` 中定义公共类型，用于对外暴露。

#### Scenario: Announcement 公告（公共类型）

`types.Announcement` **MUST** 包含：
- `ID int64` - 公告 ID
- `Title string` - 公告标题
- `Content string` - 公告内容
- `AnnounceType string` - 公告类型
- `StartTime time.Time` - 开始时间
- `EndTime time.Time` - 结束时间

**注意**: 查询参数（platform, username）由 Handler 内部从配置和 context 中获取，无需公共类型。

---

### Requirement: Type Conversion

Handler **MUST** 在内部 DTO 和公共类型之间进行转换。

#### Scenario: 公告数据转换

`GetCurrentAnnouncements` 返回时：
- **MUST** 将 `AnnouncementData` 转换为 `types.Announcement`
- **MUST** 解析时间字符串为 `time.Time`
- **SHALL** 当时间解析失败时返回零值 `time.Time{}`

#### Scenario: 查询参数构建

`GetCurrentAnnouncements` 调用时：
- **MUST** 从 `apigwclient.AppConfig.GetAppCode()` 获取 platform 参数
- **MUST** 从 `contextx.IContext.BKUsername()` 获取 username 参数

---

### Requirement: Interface Design

模块 **MUST** 遵循最小接口原则。

#### Scenario: IHandler 接口

`IHandler` **MUST** 只包含：
- `IAnnouncement` - 公告相关方法

`IHandler` **SHALL NOT** 包含：
- 生命周期管理方法（Start/Terminate）
- 应用注册方法（RegisterApplication）

#### Scenario: IAnnouncement 接口

`IAnnouncement` **MUST** 包含：
- `GetCurrentAnnouncements(nCtx contextx.IContext) ([]*types.Announcement, error)`
