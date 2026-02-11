# IAM Provider 编写指南

本文档介绍如何为 bk-nodemgr 编写 IAM 资源回调 Provider，用于支持蓝鲸权限中心（IAM）的资源查询。

## 概述

IAM Provider 是实现 IAM 资源回调接口的组件，当用户在权限中心配置权限时，IAM 会通过回调接口查询 bk-nodemgr 中的资源实例信息。

## 目录结构

```
internal/backend/router/api-v3/iam/v3/
├── v3.go                       # 路由注册 & Provider 初始化
└── provider/
    ├── provider.go             # IProvider 接口 & 数据结构定义
    ├── dispatcher.go           # 请求分发器（自动处理 filter 反序列化）
    ├── constants.go            # 常量定义
    ├── expression_helper.go    # 策略表达式求值辅助函数
    ├── networkarea.go          # NetworkArea Provider 实现示例
    ├── networkunit.go          # NetworkUnit Provider 实现示例
    ├── package.go              # Package Provider 实现示例（有父资源依赖）
    ├── packagetype.go          # PackageType Provider 实现示例
    └── provider_test.go        # 单元测试
```

## Dispatcher 工作原理

Dispatcher 负责：

1. **请求解析**：从 HTTP 请求中解析 IAM 回调参数
2. **类型转换**：将 filter 从 `map[string]interface{}` 转换为类型安全的 Filter 结构体
3. **分页处理**：转换和验证分页参数
4. **方法分发**：根据 method 调用对应的 Provider 方法
5. **错误处理**：统一处理错误并返回标准响应

**类型安全的 Filter 处理：**

```go
// Dispatcher 会根据不同的 method 自动反序列化 filter
switch RequestMethod(req.GetMethod()) {
case RequestMethodListAttr:
    // 使用 EmptyFilter
    result, err = provider.ListAttr(ctx, &Request[EmptyFilter]{...})
case RequestMethodListAttrValue:
    // 自动反序列化为 ListAttrValueFilter
    result, err = provider.ListAttrValue(ctx, &Request[ListAttrValueFilter]{...})
case RequestMethodFetchInstanceInfo:
    // 自动反序列化为 FetchInstanceFilter
    result, err = provider.FetchInstanceInfo(ctx, &Request[FetchInstanceFilter]{...})
// ...
}
```

这样 Provider 实现者不需要手动处理 filter 的类型转换和验证。

## IProvider 接口

Provider 需要实现 `IProvider` 接口，定义在 `provider/provider.go`。接口使用泛型 `Request[F any]` 实现类型安全的 filter 参数：

```go
type IProvider interface {
    // 列出资源属性（用于权限配置）
    ListAttr(ctx contextx.IContext, req *Request[EmptyFilter]) (*ListAttrData, error)

    // 列出属性值（支持关键字搜索和批量ID过滤）
    ListAttrValue(ctx contextx.IContext, req *Request[ListAttrValueFilter]) (*ListAttrValueData, error)

    // 列出资源实例（支持父资源过滤和分页）
    ListInstance(ctx contextx.IContext, req *Request[ListInstanceFilter]) (*ListInstanceData, error)

    // 获取实例详情（性能关键，用于鉴权）
    FetchInstanceInfo(ctx contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error)

    // 按策略列出实例
    ListInstanceByPolicy(ctx contextx.IContext, req *Request[ListInstanceByPolicyFilter]) (*ListInstanceData, error)

    // 搜索实例（关键字搜索）
    SearchInstance(ctx contextx.IContext, req *Request[SearchInstanceFilter]) (*ListInstanceData, error)

    // 扩展方法
    FetchInstanceList(ctx contextx.IContext, req *Request[FetchInstanceListFilter]) (*ListInstanceData, error)
    FetchResourceTypeSchema(ctx contextx.IContext, req *Request[EmptyFilter]) (*ListInstanceData, error)
}
```

### 方法说明与性能要求

| 方法 | 用途 | 性能要求 | 是否必须实现 | 官方文档 |
|------|------|----------|-------------|---------|
| `ListAttr` | 列出资源可配置的属性 | < 50ms | 是（可返回空） | [list_attr](https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback/10-list_attr.md) |
| `ListAttrValue` | 列出属性的枚举值 | 无过滤: < 50ms<br>关键字搜索: < 100ms<br>批量ID过滤(≤10): < 100ms<br>批量ID过滤(>10): < 200ms | 是（可返回空） | [list_attr_value](https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback/11-list_attr_value.md) |
| `ListInstance` | 列出资源实例（支持父资源过滤） | < 50ms | 是 | [list_instance](https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback/12-list_instance.md) |
| `FetchInstanceInfo` | 获取实例详情（鉴权关键） | 单实例: < 20ms<br>批量: < 100ms | 是 | [fetch_instance_info](https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback/13-fetch_instance_info.md) |
| `SearchInstance` | 关键字搜索实例 | < 100ms | 是 | [search_instance](https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback/15-search_instance.md) |
| `ListInstanceByPolicy` | 按策略表达式过滤实例 | < 500ms | 是（可返回空） | - |
| `FetchInstanceList` | 审计中心扩展方法 | - | 是（可返回空） | - |
| `FetchResourceTypeSchema` | 自定义扩展方法 | - | 是（可返回空） | - |

**重要说明：**
- `SearchInstance` 必须支持大小写不敏感搜索，至少支持 `display_name` 字段搜索
- 扫描数据量过大时应返回 code=422，关键字无效时应返回 code=406

## 核心数据结构

### Filter 类型

新版本使用泛型 `Request[F any]` 实现类型安全的 filter 参数，不同的 API 方法使用不同的 Filter 类型：

```go
// EmptyFilter - 用于不需要 filter 参数的方法
type EmptyFilter struct{}

// ListAttrValueFilter - 用于 list_attr_value API
type ListAttrValueFilter struct {
    Attr    string             `json:"attr"`              // 必填：属性 ID
    Keyword string             `json:"keyword,omitempty"` // 可选：搜索关键字
    IDs     []AttributeValueID `json:"ids,omitempty"`     // 可选：属性值 ID 列表（支持 string/int/bool）
}

// ListInstanceFilter - 用于 list_instance API
type ListInstanceFilter struct {
    Parent *ParentFilter `json:"parent,omitempty"` // 可选：直接父资源
}

// ParentFilter - 父资源信息
type ParentFilter struct {
    Type string `json:"type"` // 父资源类型
    ID   string `json:"id"`   // 父资源实例 ID
}

// FetchInstanceFilter - 用于 fetch_instance_info API
type FetchInstanceFilter struct {
    IDs   []string `json:"ids"`             // 必填：资源实例 ID 列表（最多 1000 个，仅支持 string）
    Attrs []string `json:"attrs,omitempty"` // 可选：要查询的属性列表，空表示所有属性
}

// ListInstanceByPolicyFilter - 用于 list_instance_by_policy API
type ListInstanceByPolicyFilter struct {
    Expression map[string]interface{} `json:"expression"` // 必填：策略表达式（动态结构）
}

// SearchInstanceFilter - 用于 search_instance API
type SearchInstanceFilter struct {
    Keyword string        `json:"keyword"`          // 必填：搜索关键字
    Parent  *ParentFilter `json:"parent,omitempty"` // 可选：父资源过滤
}

// FetchInstanceListFilter - 用于 fetch_instance_list API（审计中心）
type FetchInstanceListFilter struct {
    StartTime int64 `json:"start_time"` // 必填：开始时间（毫秒）
    EndTime   int64 `json:"end_time"`   // 必填：结束时间（毫秒）
}
```

### 请求结构

```go
// Request 是泛型请求结构，F 是 filter 类型
type Request[F any] struct {
    Type   string     // 资源类型
    Method string     // 回调方法
    Filter F          // 过滤条件（类型安全）
    Page   types.Page // 分页参数（已转换和验证）
}
```

### 响应数据结构

```go
// 资源属性
type ResourceAttribute struct {
    ID          string `json:"id"`           // 属性唯一标识
    DisplayName string `json:"display_name"` // 属性显示名称
}

// 属性值
type AttributeValue struct {
    ID          interface{} `json:"id"`           // 属性值（可以是 string/int/bool）
    DisplayName string      `json:"display_name"` // 属性值显示名称
}

// 资源实例
type ResourceInstance struct {
    ID          string `json:"id"`                    // 实例唯一标识
    DisplayName string `json:"display_name"`          // 实例显示名称
    ChildType   string `json:"child_type,omitempty"`  // 子类型（可选）
}

// 实例详情
type InstanceInfo struct {
    ID          string                 `json:"id"`
    DisplayName string                 `json:"display_name,omitempty"`
    Attributes  map[string]interface{} `json:"-"`  // 动态属性，序列化时会展平
}
```

### ID 类型说明

IAM 回调接口涉及两种不同类型的 ID，需要区分处理：

| ID 类型 | 适用 API | 支持的数据类型 | 说明 |
|---------|----------|---------------|------|
| **资源实例 ID** | `fetch_instance_info` | **仅 string** | IAM 规范明确定义为 `array(string)` |
| **属性值 ID** | `list_attr_value` | **string/int/bool** | 属性值可以是多种类型 |

**重要区别：**

1. **fetch_instance_info 的 filter.ids**
   - 只支持 string 类型：`{"ids": ["1", "2", "3"]}`
   - 代码中使用 `[]string` 直接接收（`FetchInstanceFilter.IDs`）
   - 如果你的数据库存储的是 int64，需要先用 `strconv.ParseInt` 转换
   - 最多支持 1000 个 ID（`MaxFetchInstanceIDs = 1000`）

2. **list_attr_value 的 filter.ids**
   - 支持多种类型：`{"ids": ["linux", 123, true]}`
   - 代码中使用 `[]AttributeValueID` 类型（自定义反序列化）
   - 该类型会自动将 int/bool 转换为 string

**AttributeValueID 类型定义：**

```go
// AttributeValueID 自动处理 string/int/bool 类型转换
type AttributeValueID string

func (id *AttributeValueID) UnmarshalJSON(data []byte) error {
    // 自动将 string/int/bool 转换为 string
    // ...
}
```

**示例代码：**

```go
// FetchInstanceInfo - 资源实例 ID 是 []string
func (p *MyResourceProvider) FetchInstanceInfo(ctx contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
    // req.Filter.IDs 是 []string（IAM 规范只支持 string）
    ids := req.Filter.IDs

    // 如果数据库使用 int64，需要转换
    int64IDs := make([]int64, 0, len(ids))
    for _, idStr := range ids {
        id, err := strconv.ParseInt(idStr, 10, 64)
        if err != nil {
            continue
        }
        int64IDs = append(int64IDs, id)
    }
    // ...
}

// ListAttrValue - 属性值 ID 支持多种类型
func (p *MyResourceProvider) ListAttrValue(ctx contextx.IContext, req *Request[ListAttrValueFilter]) (*ListAttrValueData, error) {
    // req.Filter.IDs 是 []AttributeValueID（自动转换为 string）
    if len(req.Filter.IDs) > 0 {
        // 可以直接使用，已经是 string 类型
        for _, id := range req.Filter.IDs {
            idStr := string(id) // 转换为 string
            // ...
        }
    }
    // ...
}
```

## 策略表达式处理（ListInstanceByPolicy）

`ListInstanceByPolicy` 方法需要根据 IAM 策略表达式过滤资源实例。项目提供了 `expression_helper.go` 辅助函数来简化实现。

### 策略表达式示例

IAM 策略表达式是一个嵌套的 JSON 结构，用于描述资源过滤条件：

```json
{
  "op": "AND",
  "content": [
    {
      "op": "eq",
      "field": "id",
      "value": "agent"
    },
    {
      "op": "in",
      "field": "status",
      "value": ["active", "pending"]
    }
  ]
}
```

### 使用 evalExpressionFilter 辅助函数

```go
func (p *MyResourceProvider) ListInstanceByPolicy(ctx contextx.IContext, req *Request[ListInstanceByPolicyFilter]) (*ListInstanceData, error) {
    // 1. 加载所有实例（使用大分页限制）
    allPage := types.Page{Limit: MaxListInstanceByPolicyLimit, Offset: 0}
    items, _, err := p.storage.List(ctx, allPage)
    if err != nil {
        logger.G.Biz(ctx).WithErr(err).Error("failed to list instances for policy evaluation")
        return nil, fmt.Errorf("failed to list instances: %w", err)
    }

    // 2. 构建 InstanceForEval 列表
    instances := make([]InstanceForEval, 0, len(items))
    for _, item := range items {
        instances = append(instances, InstanceForEval{
            Instance: ResourceInstance{
                ID:          strconv.FormatInt(item.ID, 10),
                DisplayName: item.Name,
            },
            Attributes: map[string]interface{}{
                IAMAttrID:          strconv.FormatInt(item.ID, 10),
                IAMAttrDisplayName: item.Name,
                // 添加其他需要在表达式中使用的属性
                "status": item.Status,
                "region": item.Region,
            },
        })
    }

    // 3. 使用辅助函数评估表达式并分页
    return evalExpressionFilter(
        req.Filter.Expression,
        ResourceTypeMyResource,
        instances,
        req.Page,
    )
}
```

### evalExpressionFilter 函数说明

```go
// evalExpressionFilter 评估 IAM 策略表达式并过滤资源实例
//
// 参数：
//   - expressionMap: IAM 策略表达式（来自回调请求）
//   - resourceType: 资源类型标识符（如 "network_area", "package"）
//   - instances: 待评估的实例列表（包含属性）
//   - page: 分页参数
//
// 返回：
//   - *ListInstanceData: 过滤和分页后的结果（Count 是过滤后的总数）
//   - error: 表达式评估错误
//
// 行为：
//   - 空表达式：返回所有实例（无过滤）
//   - op="any"：返回所有实例（SDK 处理）
//   - 有效表达式：使用 ExprCell.Eval() 过滤实例
//   - 分页：在过滤后应用，Count 是过滤后的总数
```

### 注意事项

1. **性能考虑**：
   - 需要加载所有实例到内存进行表达式评估
   - 使用 `MaxListInstanceByPolicyLimit = 10000` 限制最大加载数量
   - 性能要求 < 500ms

2. **属性映射**：
   - `Attributes` 中的 key 必须与 IAM 权限模型中定义的属性 ID 一致
   - 必须包含 `id` 和 `display_name` 属性
   - 可以包含自定义属性用于表达式过滤

3. **表达式求值**：
   - 使用 `github.com/TencentBlueKing/iam-go-sdk/expression` 包
   - 支持 `eq`, `in`, `contains`, `starts_with`, `ends_with` 等操作符
   - 支持 `AND`, `OR`, `NOT` 逻辑组合

4. **简化实现**：
   - 如果资源数量较少且不需要复杂的策略过滤，可以直接返回空结果
   - IAM 目前很少使用此 API

## 编写新 Provider

### Step 1: 创建 Provider 文件

在 `internal/backend/router/api-v3/iam/v3/provider/` 目录下创建新文件，例如 `myresource.go`。

### Step 2: 定义资源类型常量

```go
package provider

// ResourceTypeMyResource 是 IAM 中定义的资源类型 ID
// 必须与 support-files/bkiamv3/templates/ 中的权限模型定义一致
const ResourceTypeMyResource = "myresource"
```

### Step 3: 创建 Provider 结构体

```go
// MyResourceProvider 实现 IProvider 接口
type MyResourceProvider struct {
    storage myresource.IStorage  // 注入存储层依赖
}

// NewMyResourceProvider 创建 MyResourceProvider 实例
func NewMyResourceProvider(storage myresource.IStorage) *MyResourceProvider {
    return &MyResourceProvider{storage: storage}
}
```

### Step 4: 实现接口方法

参考前面的"核心数据结构"和各个方法的实现示例。

### Step 5: 注册 Provider

在 `internal/backend/router/api-v3/iam/v3/v3.go` 的 `Load` 函数中注册：

```go
func Load(rg *gin.RouterGroup, capability *options.Capability) {
    h := newHandler(rg, capability)
    h.dispatcher = provider.NewDispatcher()

    // 注册现有 Provider
    networkAreaProvider := provider.NewNetworkAreaProvider(capability.StorageTopo)
    h.dispatcher.RegisterProvider(provider.ResourceTypeNetworkArea, networkAreaProvider)

    // 注册新的 Provider
    myResourceProvider := provider.NewMyResourceProvider(capability.StorageMyResource)
    h.dispatcher.RegisterProvider(provider.ResourceTypeMyResource, myResourceProvider)

    // ... 其他代码
}
```

#### ListAttr - 列出资源属性

如果资源没有可配置的属性，返回空列表：

```go
func (p *MyResourceProvider) ListAttr(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListAttrData, error) {
    // 如果资源有属性，返回属性列表
    // 例如: return &ListAttrData{Results: []ResourceAttribute{
    //     {ID: "status", DisplayName: "状态"},
    //     {ID: "region", DisplayName: "地域"},
    // }}, nil

    // 没有属性时返回空列表
    return &ListAttrData{Results: []ResourceAttribute{}}, nil
}
```

#### ListAttrValue - 列出属性值

```go
func (p *MyResourceProvider) ListAttrValue(_ contextx.IContext, req *Request[ListAttrValueFilter]) (*ListAttrValueData, error) {
    // 根据 req.Filter.Attr 返回对应属性的值列表
    // 支持 req.Filter.Keyword 关键字搜索
    // 支持 req.Filter.IDs 批量 ID 过滤

    // 如果没有属性值，返回空列表
    return &ListAttrValueData{Count: 0, Results: []AttributeValue{}}, nil
}
```

#### ListInstance - 列出实例（核心方法）

```go
func (p *MyResourceProvider) ListInstance(ctx contextx.IContext, req *Request[ListInstanceFilter]) (*ListInstanceData, error) {
    // 检查是否有父资源过滤
    // 如果资源在 IAM 权限模型中定义了必需的父资源，需要检查 req.Filter.Parent
    if req.Filter.Parent != nil {
        // 验证父资源类型
        if req.Filter.Parent.Type != ExpectedParentType {
            err := fmt.Errorf("invalid parent type: expected %s, got %s",
                ExpectedParentType, req.Filter.Parent.Type)
            logger.G.Biz(ctx).WithErr(err).Error("invalid parent type")
            return nil, err
        }
        // 根据父资源 ID 过滤
        parentID := req.Filter.Parent.ID
        // ...
    }

    // 使用 req.Page 进行分页查询
    items, total, err := p.storage.List(ctx, req.Page)
    if err != nil {
        logger.G.Biz(ctx).WithErr(err).Error("failed to list resources")
        return nil, fmt.Errorf("failed to list resources: %w", err)
    }

    // 转换为 IAM 响应格式
    results := make([]ResourceInstance, 0, len(items))
    for _, item := range items {
        results = append(results, ResourceInstance{
            ID:          strconv.FormatInt(item.ID, 10),
            DisplayName: item.Name,
        })
    }

    return &ListInstanceData{Count: total, Results: results}, nil
}
```

**父资源过滤说明：**

- 如果资源在 IAM 权限模型中定义了必需的父资源（如 package 必须有 package_type 父资源），当 `req.Filter.Parent == nil` 时应返回空结果
- 如果资源没有父资源依赖（如 networkarea），可以忽略 `req.Filter.Parent`

#### FetchInstanceInfo - 获取实例详情（核心方法）

```go
func (p *MyResourceProvider) FetchInstanceInfo(ctx contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
    // req.Filter.IDs 已经是 []string 类型
    ids := req.Filter.IDs

    if len(ids) == 0 {
        return &FetchInstanceInfoData{Results: []InstanceInfo{}}, nil
    }

    // 如果数据库使用 int64 类型的 ID，需要转换
    int64IDs := make([]int64, 0, len(ids))
    for _, idStr := range ids {
        id, err := strconv.ParseInt(idStr, 10, 64)
        if err != nil {
            continue
        }
        int64IDs = append(int64IDs, id)
    }

    // 根据 ID 查询资源
    items, _, err := p.storage.ListByIDs(ctx, int64IDs)
    if err != nil {
        logger.G.Biz(ctx).WithErr(err).Error("failed to fetch instance info")
        return nil, fmt.Errorf("failed to fetch instance info: %w", err)
    }

    // 转换为 IAM 响应格式
    results := make([]InstanceInfo, 0, len(items))
    for _, item := range items {
        results = append(results, InstanceInfo{
            ID:          strconv.FormatInt(item.ID, 10),
            DisplayName: item.Name,
            Attributes:  make(map[string]interface{}),
        })
    }

    return &FetchInstanceInfoData{Results: results}, nil
}
```

**性能优化建议：**

- 这是性能关键的 API，用于鉴权，必须快速响应
- 单实例查询应 < 20ms，批量查询应 < 100ms
- 考虑使用缓存优化查询性能
- 最多支持 1000 个 ID（`MaxFetchInstanceIDs = 1000`）

#### SearchInstance - 搜索实例（核心方法）

```go
func (p *MyResourceProvider) SearchInstance(ctx contextx.IContext, req *Request[SearchInstanceFilter]) (*ListInstanceData, error) {
    // 提取搜索关键字（必填）
    keyword := strings.TrimSpace(req.Filter.Keyword)

    // 检查父资源过滤（可选）
    if req.Filter.Parent != nil {
        // 根据父资源过滤
        // ...
    }

    // 构建查询条件（大小写不敏感）
    var condition *types.MyResourceCondition
    if keyword != "" {
        condition = &types.MyResourceCondition{
            FuzzyInclude: &types.MyResourceFuzzyFields{
                Name: []string{keyword},
            },
        }
    }

    // 查询资源
    items, total, err := p.storage.List(ctx, req.Page, condition)
    if err != nil {
        logger.G.Biz(ctx).WithErr(err).Error("failed to search resources")
        return nil, fmt.Errorf("failed to search resources: %w", err)
    }

    // 转换为 IAM 响应格式
    results := make([]ResourceInstance, 0, len(items))
    for _, item := range items {
        results = append(results, ResourceInstance{
            ID:          strconv.FormatInt(item.ID, 10),
            DisplayName: item.Name,
        })
    }

    return &ListInstanceData{Count: total, Results: results}, nil
}
```

**重要说明：**

- 搜索必须支持大小写不敏感（case-insensitive）
- 至少支持 `display_name` 字段搜索
- 扫描数据量过大时应返回 code=422
- 关键字无效时应返回 code=406

#### 其他方法（可返回空结果）

```go
// ListInstanceByPolicy - 按策略表达式过滤，目前 IAM 未使用
func (p *MyResourceProvider) ListInstanceByPolicy(_ contextx.IContext, _ *Request[ListInstanceByPolicyFilter]) (*ListInstanceData, error) {
    return &ListInstanceData{Count: 0, Results: []ResourceInstance{}}, nil
}

// FetchInstanceList - 审计中心扩展方法
func (p *MyResourceProvider) FetchInstanceList(_ contextx.IContext, _ *Request[FetchInstanceListFilter]) (*ListInstanceData, error) {
    return &ListInstanceData{Count: 0, Results: []ResourceInstance{}}, nil
}

// FetchResourceTypeSchema - 自定义扩展方法
func (p *MyResourceProvider) FetchResourceTypeSchema(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListInstanceData, error) {
    return &ListInstanceData{Count: 0, Results: []ResourceInstance{}}, nil
}
```

## 完整示例

参考现有实现：

### 简单资源（无父资源依赖）
- `internal/backend/router/api-v3/iam/v3/provider/networkarea.go` - NetworkArea Provider
- `internal/backend/router/api-v3/iam/v3/provider/networkunit.go` - NetworkUnit Provider

### 复杂资源（有父资源依赖）
- `internal/backend/router/api-v3/iam/v3/provider/package.go` - Package Provider（依赖 PackageType 父资源）
- `internal/backend/router/api-v3/iam/v3/provider/packagetype.go` - PackageType Provider

### Package Provider 关键实现

Package 资源在 IAM 权限模型中定义了必需的父资源 `package_type`，因此在 `ListInstance` 方法中需要检查父资源：

```go
func (p *PackageProvider) ListInstance(ctx contextx.IContext, req *Request[ListInstanceFilter]) (*ListInstanceData, error) {
    // Package 必须有父资源 package_type
    if req.Filter.Parent == nil {
        return &ListInstanceData{Count: 0, Results: []ResourceInstance{}}, nil
    }

    // 验证父资源类型
    if req.Filter.Parent.Type != ResourceTypePackageType {
        err := fmt.Errorf("invalid parent type: expected %s, got %s",
            ResourceTypePackageType, req.Filter.Parent.Type)
        logger.G.Biz(ctx).WithErr(err).Error("invalid parent type in package provider")
        return nil, err
    }

    releaseType := types.ReleaseType(req.Filter.Parent.ID)

    // 根据不同的 package_type 返回不同的实例
    switch releaseType {
    case types.ReleaseTypeAgent:
        return &ListInstanceData{
            Count:   1,
            Results: []ResourceInstance{{ID: string(types.ReleaseTypeAgent), DisplayName: string(types.ReleaseTypeAgent)}},
        }, nil
    case types.ReleaseTypePlugin:
        return p.listPluginInstances(ctx, req.Page)
    // ...
    }
}
```

### 测试

完整的单元测试示例参考：
- `internal/backend/router/api-v3/iam/v3/provider/provider_test.go`

## IAM 权限模型配置

Provider 的资源类型必须与 IAM 权限模型定义一致，权限模型模板位于：

```
support-files/bkiamv3/templates/
├── 0001_bk_nodemgr_system.json.tpl        # 系统注册
├── 0002_bk_nodemgr_resource_type.json.tpl # 资源类型定义
├── 0003_bk_nodemgr_actions.json.tpl       # 操作定义
├── 0004_bk_nodemgr_action_groups.json.tpl # 操作分组
└── 0005_bk_nodemgr_common_actions.json.tpl # 常用操作
```

确保 Provider 的 `ResourceType` 常量与 `0002_bk_nodemgr_resource_type.json.tpl` 中定义的 `id` 字段一致。

## 参考资料

### IAM 官方文档

所有 IAM 回调 API 的详细文档已在"方法说明与性能要求"表格中列出，包括：

- [list_attr](https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback/10-list_attr.md) - 拉取某个资源类型可用于配置权限的属性列表
- [list_attr_value](https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback/11-list_attr_value.md) - 获取一个属性的值列表
- [list_instance](https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback/12-list_instance.md) - 根据过滤条件查询实例
- [fetch_instance_info](https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback/13-fetch_instance_info.md) - 批量获取资源实例详情
- [search_instance](https://raw.githubusercontent.com/TencentBlueKing/BKDocs/refs/heads/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback/15-search_instance.md) - 搜索资源实例

### 代码示例

- `internal/backend/router/api-v3/iam/v3/provider/` - Provider 实现目录
- `internal/backend/router/api-v3/iam/v3/provider/provider_test.go` - 单元测试示例

## 注意事项

1. **性能要求**：IAM 回调接口对性能有严格要求，务必注意响应时间
   - `FetchInstanceInfo` 是鉴权关键 API，单实例 < 20ms，批量 < 100ms
   - `ListInstance` 和 `ListAttr` 应 < 50ms
   - `SearchInstance` 应 < 100ms，必须支持大小写不敏感搜索

2. **类型安全**：使用泛型 `Request[F any]` 实现类型安全的 filter 参数
   - 不同的 API 方法使用不同的 Filter 类型
   - Dispatcher 会自动处理 filter 的反序列化和类型转换

3. **ID 格式**：
   - 资源实例 ID 在返回时统一转换为字符串格式
   - `FetchInstanceInfo` 的 filter.ids 只支持 `[]string`（最多 1000 个）
   - `ListAttrValue` 的 filter.ids 支持 `[]AttributeValueID`（string/int/bool）

4. **父资源过滤**：
   - 如果资源在 IAM 权限模型中定义了必需的父资源，`ListInstance` 必须检查 `req.Filter.Parent`
   - 当 `Parent == nil` 时应返回空结果
   - 如果资源没有父资源依赖，可以忽略 `Parent` 字段

5. **错误处理**：
   - 使用 `logger.G.Biz(ctx)` 记录错误日志，便于问题排查
   - 返回有意义的错误信息，帮助定位问题

6. **空结果**：
   - 对于不适用的方法，返回空列表而不是 nil
   - 例如：`&ListAttrData{Results: []ResourceAttribute{}}`

7. **分页参数**：
   - `req.Page` 已在 Dispatcher 中完成转换和验证，可直接使用
   - 默认 limit 为 20，最大为 `MaxListInstanceByPolicyLimit = 10000`

8. **常量定义**：
   - `MaxFetchInstanceIDs = 1000` - fetch_instance_info 最多支持 1000 个 ID
   - `MaxListInstanceByPolicyLimit = 10000` - 策略表达式评估时的最大分页限制

9. **搜索要求**：
   - `SearchInstance` 必须支持大小写不敏感搜索
   - 至少支持 `display_name` 字段搜索
   - 扫描数据量过大时返回 code=422
   - 关键字无效时返回 code=406
