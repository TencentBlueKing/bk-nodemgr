# storage

## 接口规范
- IStorage: 用于提供整个包的全部接口，主要在初始化时使用
- IDaoxxx: 提供单个 Storage 下各数据表的单表数据访问能力
- IDomainxxx: 提供针对业务领域的数据访问能力

## 能力边界

storage 层用于提供针对场景的数据能力：

### 数据聚合和分析能力

### 异步任务处理能力

### 数据缓存能力

### 跨域数据访问能力

### 非标准数据的处理能力

例如，空值、空字符串、空数组在传递到下层的时候会被拦截并报错，而这些错误是上层业务不应该关心的，因此我们将在 storage
层进行拦截和处理。

```go
// UpsertHosts ...
func (s *storage) UpsertHosts(ctx context.Context, hosts ...*types.Host) error {
if ctx == nil {
return errors.New("ctx is nil")
}

// 我们在此处拦截了空数组，因为这是业务不应该关心的错误
if len(hosts) == 0 {
return nil
}

if err := s.daoHost.UpsertMany(ctx, hosts); err != nil {
return fmt.Errorf("failed to upsert hosts: %v", err)
}

return nil
}
```