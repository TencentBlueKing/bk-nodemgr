## ADDED Requirements

### Requirement: NetworkUnitProvider 实现 IProvider 接口
系统 SHALL 实现 `NetworkUnitProvider` 结构体,实现 `IProvider` 接口的所有方法,作为 IAM 权限中心的 Network Unit 资源回调提供者。

#### Scenario: Provider 初始化
- **WHEN** 创建 `NetworkUnitProvider` 实例时
- **THEN** 系统 SHALL 接受 `topo.IStorage` 作为依赖参数
- **THEN** 系统 SHALL 返回完整初始化的 provider 实例

### Requirement: 资源类型常量定义
系统 SHALL 定义 Network Unit 的资源类型常量为 `"networkunit"`,用于 IAM 回调请求的资源类型标识。

#### Scenario: 资源类型标识
- **WHEN** IAM 权限中心发起资源回调请求时
- **THEN** 系统 SHALL 使用 `ResourceTypeNetworkUnit = "networkunit"` 作为资源类型标识符

### Requirement: ListInstance 分页列出实例
系统 SHALL 实现 `ListInstance` 方法,支持分页列出 Network Unit 实例,返回实例 ID 和名称。

#### Scenario: 成功列出实例
- **WHEN** 接收到 `ListInstance` 请求,包含分页参数 `offset` 和 `limit`
- **THEN** 系统 SHALL 调用 `storage.ListNetworkUnit()` 查询数据
- **THEN** 系统 SHALL 将 `NetworkUnit.ID` 转换为字符串作为实例 ID
- **THEN** 系统 SHALL 将 `NetworkUnit.Name` 作为 `display_name`
- **THEN** 系统 SHALL 返回 `ListInstanceData` 包含 `count` 和 `results`

#### Scenario: 空结果处理
- **WHEN** 数据库中没有 Network Unit 数据时
- **THEN** 系统 SHALL 返回 `count=0` 和空的 `results` 数组

#### Scenario: 查询失败处理
- **WHEN** `storage.ListNetworkUnit()` 返回错误时
- **THEN** 系统 SHALL 记录错误日志
- **THEN** 系统 SHALL 返回带有错误信息的 error

### Requirement: FetchInstanceInfo 批量获取实例详情
系统 SHALL 实现 `FetchInstanceInfo` 方法,支持按 IDs 批量获取 Network Unit 详细信息,包括实例 ID、名称和属性。

#### Scenario: 成功获取实例详情
- **WHEN** 接收到 `FetchInstanceInfo` 请求,包含 `filter.ids` 参数
- **THEN** 系统 SHALL 从 filter 中提取 Network Unit IDs (支持字符串或数字格式)
- **THEN** 系统 SHALL 构造 `NetworkUnitCondition` 使用 `ExactInclude.NetworkUnitID`
- **THEN** 系统 SHALL 调用 `storage.ListNetworkUnit()` 查询指定 IDs 的数据
- **THEN** 系统 SHALL 返回 `FetchInstanceInfoData` 包含每个实例的 `id`, `display_name` 和空的 `attributes` map

#### Scenario: 空 IDs 处理
- **WHEN** filter 中没有 ids 字段或 ids 为空数组时
- **THEN** 系统 SHALL 返回空的 `results` 数组

#### Scenario: ID 格式转换
- **WHEN** filter.ids 包含字符串格式的 ID (如 "123")
- **THEN** 系统 SHALL 将字符串解析为 int64
- **WHEN** filter.ids 包含数字格式的 ID
- **THEN** 系统 SHALL 使用 `conv.ToInt64()` 进行类型转换

#### Scenario: 无效 ID 跳过
- **WHEN** filter.ids 中包含无法解析的 ID 值时
- **THEN** 系统 SHALL 跳过该 ID,继续处理其他有效 ID

### Requirement: SearchInstance 关键字搜索
系统 SHALL 实现 `SearchInstance` 方法,支持按关键字搜索 Network Unit,对名称字段进行模糊匹配。

#### Scenario: 成功搜索实例
- **WHEN** 接收到 `SearchInstance` 请求,包含 `filter.keyword` 参数
- **THEN** 系统 SHALL 从 filter 中提取 keyword 并去除首尾空格
- **THEN** 系统 SHALL 构造 `NetworkUnitCondition` 使用 `FuzzyInclude.NetworkUnitName`
- **THEN** 系统 SHALL 调用 `storage.ListNetworkUnit()` 进行模糊查询
- **THEN** 系统 SHALL 返回匹配的实例列表和总数

#### Scenario: 空关键字处理
- **WHEN** filter 中没有 keyword 或 keyword 为空字符串时
- **THEN** 系统 SHALL 返回所有 Network Unit 实例(受分页限制)

#### Scenario: 关键字大小写处理
- **WHEN** 使用关键字搜索时
- **THEN** 系统 SHALL 依赖底层存储层的大小写敏感性实现

### Requirement: ListAttr 返回空属性列表
系统 SHALL 实现 `ListAttr` 方法,返回空的资源属性列表,因为 Network Unit 无自定义属性用于权限配置。

#### Scenario: 返回空属性
- **WHEN** 接收到 `ListAttr` 请求时
- **THEN** 系统 SHALL 返回 `ListAttrData` 包含空的 `results` 数组

### Requirement: ListAttrValue 返回空属性值列表
系统 SHALL 实现 `ListAttrValue` 方法,返回空的属性值列表,因为 Network Unit 无自定义属性。

#### Scenario: 返回空属性值
- **WHEN** 接收到 `ListAttrValue` 请求时
- **THEN** 系统 SHALL 返回 `ListAttrValueData` 包含 `count=0` 和空的 `results` 数组

### Requirement: ListInstanceByPolicy 返回空结果
系统 SHALL 实现 `ListInstanceByPolicy` 方法,返回空结果,该方法用于权限预览功能,当前不需要实现。

#### Scenario: 返回空结果
- **WHEN** 接收到 `ListInstanceByPolicy` 请求时
- **THEN** 系统 SHALL 返回 `ListInstanceData` 包含 `count=0` 和空的 `results` 数组

### Requirement: FetchInstanceList 返回空结果
系统 SHALL 实现 `FetchInstanceList` 方法,返回空结果,该方法用于审计中心功能,当前不需要实现。

#### Scenario: 返回空结果
- **WHEN** 接收到 `FetchInstanceList` 请求时
- **THEN** 系统 SHALL 返回 `ListInstanceData` 包含 `count=0` 和空的 `results` 数组

### Requirement: FetchResourceTypeSchema 返回空 Schema
系统 SHALL 实现 `FetchResourceTypeSchema` 方法,返回空结果,该方法不是标准 IAM 回调 API 的一部分。

#### Scenario: 返回空 Schema
- **WHEN** 接收到 `FetchResourceTypeSchema` 请求时
- **THEN** 系统 SHALL 返回 `ListInstanceData` 包含 `count=0` 和空的 `results` 数组

### Requirement: Provider 注册到 Dispatcher
系统 SHALL 在 IAM 路由初始化时,将 `NetworkUnitProvider` 注册到 `Dispatcher`,使用资源类型 `"networkunit"` 作为标识。

#### Scenario: Provider 注册
- **WHEN** 初始化 IAM v3 路由时
- **THEN** 系统 SHALL 创建 `NetworkUnitProvider` 实例,传入 storage 依赖
- **THEN** 系统 SHALL 调用 `dispatcher.RegisterProvider("networkunit", provider)` 注册 provider

### Requirement: 错误日志记录
系统 SHALL 在所有数据库查询失败时记录详细的错误日志,包含上下文信息。

#### Scenario: 查询失败日志
- **WHEN** 任何 `storage.ListNetworkUnit()` 调用返回错误时
- **THEN** 系统 SHALL 使用 `logger.G.Biz(ctx).WithErr(err).Error()` 记录错误
- **THEN** 日志 SHALL 包含失败的操作名称(如 "failed to list network units")

### Requirement: 性能要求
系统 SHALL 满足 IAM 回调 API 的性能要求,确保响应时间符合规范。

#### Scenario: 单实例查询性能
- **WHEN** `FetchInstanceInfo` 查询单个实例时
- **THEN** 系统 SHALL 在 20 毫秒内返回结果

#### Scenario: 批量查询性能
- **WHEN** `FetchInstanceInfo` 查询多个实例或 `ListInstance` 分页查询时
- **THEN** 系统 SHALL 在 100 毫秒内返回结果

#### Scenario: 搜索性能
- **WHEN** `SearchInstance` 执行关键字搜索时
- **THEN** 系统 SHALL 在 100 毫秒内返回结果
