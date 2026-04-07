# IAM 回调接口错误代码映射

本文档说明 IAM 资源回调接口的错误代码与 HTTP 状态码的映射关系。

## 错误代码映射表

| IAM 规范状态码 | 业务错误代码 | HTTP 状态码 | 错误场景 | 错误消息 |
|---------------|-------------|------------|---------|---------|
| 0 | `errf.OK` | 200 | 请求成功 | OK |
| 401 | `errf.Unauthorized` | 401 | Basic Auth 认证失败 | unauthorized |
| 404 | `errf.RecordNotFound` | 404 | 资源类型或方法不存在 | record not found |
| 406 | `errf.InvalidKeyword` | 406 | 搜索关键字不符合要求 | invalid keyword |
| 422 | `errf.ResourceScanTooLarge` | 422 | 扫描的资源内容过多 | resource scan too large |
| 429 | `errf.TooManyRequest` | 429 | 请求超过频率控制 | too many request |
| 500 | `errf.Unknown` | 500 | 系统错误或异常 | unknown error |

## 错误代码使用指南

### 1. 认证失败 (401)

**场景**：Basic Auth 中间件验证 IAM 系统 Token 失败

**实现位置**：`v3.go` 的 `basicAuthMiddleware()`

```go
if err := h.capability.IAMV3Handler.IsBasicAuthAllowed(ctx, username, password); err != nil {
    c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
        "code":    http.StatusUnauthorized,
        "message": "invalid credentials",
    })
    return
}
```

### 2. 资源类型不存在 (404)

**场景**：
- 请求的 `type` 字段对应的 Provider 未注册
- 请求的 `method` 字段无效

**实现位置**：`dispatcher.go` 的 `Dispatch()` 方法

```go
provider, exist := d.GetProvider(req.GetType())
if !exist {
    err := fmt.Errorf("resource type %s not supported", req.GetType())
    return nil, resterrf.ErrWrap(resterrf.RecordNotFound, err)
}
```

### 3. 参数验证失败 (400)

**场景**：
- 请求体解析失败
- 分页参数超过限制
- 必需参数缺失

**实现位置**：`provider.go` 的 `ValidateAndNormalizePage()`

```go
if page.Limit > MaxPageLimit {
    return fmt.Errorf("page limit must not exceed %d, got %d", MaxPageLimit, page.Limit)
}
// Dispatcher 会将此错误包装为 resterrf.InvalidParameter
```

### 4. 搜索关键字无效 (406)

**场景**：搜索关键字格式不符合要求（如包含非法字符、长度超限等）

**使用示例**：

```go
func (p *NetworkAreaProvider) SearchInstance(ctx contextx.IContext, req *Request) (*ListInstanceData, error) {
    keyword := p.extractKeywordFromFilter(req.Filter)
    
    // 验证关键字格式
    if keyword != "" && len(keyword) > 100 {
        err := fmt.Errorf("keyword too long: max 100 characters, got %d", len(keyword))
        return nil, resterrf.ErrWrap(resterrf.InvalidKeyword, err)
    }
    
    // ... 继续处理
}
```

### 5. 扫描范围过大 (422)

**场景**：查询或搜索操作需要扫描的资源数量过多，拒绝返回数据以保护系统性能

**使用示例**：

```go
func (p *NetworkAreaProvider) SearchInstance(ctx contextx.IContext, req *Request) (*ListInstanceData, error) {
    // 估算扫描范围
    estimatedScanSize := estimateScanSize(ctx, condition)
    
    if estimatedScanSize > 50000 {
        err := fmt.Errorf("search would scan too many records: %d", estimatedScanSize)
        return nil, resterrf.ErrWrap(resterrf.ResourceScanTooLarge, err)
    }
    
    // ... 继续处理
}
```

### 6. 频率限制 (429)

**场景**：接口请求超过接入系统的频率控制

**说明**：通常由中间件或 API Gateway 处理，Provider 层无需特殊处理

### 7. 系统错误 (500)

**场景**：
- 数据库查询失败
- 依赖服务不可用
- 未预期的异常

**实现位置**：`dispatcher.go` 的错误处理

```go
if err != nil {
    logger.G.Biz(rCtx).WithErr(err).Error("provider execution failed")
    return nil, resterrf.ErrWrap(resterrf.Unknown, err)
}
```

## 分页限制配置

为了防止性能问题，IAM 回调接口实现了分页限制。

**定义位置**：`pkg/proto/backend/api/v3/iam.go`

```go
const (
    // IAMCallbackMaxPageLimit 是 IAM 回调接口单页最大记录数
    IAMCallbackMaxPageLimit = 1000
    
    // IAMCallbackDefaultPageLimit 是默认分页大小
    IAMCallbackDefaultPageLimit = 100
)
```

**转换函数**：

```go
// ConvIAMCallbackPageToTypes 转换并验证分页参数
func ConvIAMCallbackPageToTypes(reqPage *IAMResourceCallbackPage) (types.Page, error)
```

当 `page.limit` 超过 `IAMCallbackMaxPageLimit` 时，将返回 400 错误（InvalidParameter）。

## 最佳实践

1. **参数验证优先**：在访问数据库前验证所有参数，及早返回错误
2. **精确的错误代码**：根据具体错误场景选择合适的错误代码，避免都使用 `Unknown`
3. **详细的错误消息**：包含足够的上下文信息帮助排查问题
4. **日志记录**：所有错误都应该记录日志，包含关键参数
5. **性能保护**：对可能导致大量扫描的操作添加保护逻辑

## 参考文档

- [IAM 回调接口规范](https://github.com/TencentBlueKing/BKDocs/tree/main/ZH/IAM/IntegrateGuide/Reference/API/03-Callback)
- [iam-python-sdk 示例](https://github.com/TencentBlueKing/iam-python-sdk/blob/master/docs/usage.md)
