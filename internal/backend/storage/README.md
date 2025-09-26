# storage

## 文件职责
- iface.go:
  - 定义 IStorage 接口
  - 定义 IDomainxxx 接口
  - 定义 IDaoxxx 接口
- storage.go: 
  - 包入口点, 所有的导出函数仅可在此文件中实现;
  - 提供指标采集能力
  - Storage 的数据结构声明
  - Storage 的初始化函数
  - 此处需要对传入参数进行检查，如果传入参数不合法，应该返回错误
- dao_xxx.go:
  - 用于提供针对于单表的数据访问能力
- domain_xxx.go:
  - 用于提供针对业务领域的数据访问能力

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
func (s *storage) UpsertHosts(nCtx contextx.IContext, hosts ...*types.Host) error {
if nCtx == nil {
return errors.New("nCtx is nil")
}

// 我们在此处拦截了空数组，因为这是业务不应该关心的错误
if len(hosts) == 0 {
return nil
}

if err := s.daoHost.UpsertMany(nCtx, hosts); err != nil {
return fmt.Errorf("failed to upsert hosts: %v", err)
}

return nil
}
```

### 通用函数命名规范

- Create: 创建, 等价于 dao 层的 Create
- CreateMany: 批量创建, 等价于 dao 层的 CreateMany
- Update: 更新, 等价于 dao 层的 Update
- UpdateMany: 批量更新, 等价于 dao 层的 UpdateMany
- Delete: 删除, 等价于 dao 层的 Delete
- DeleteMany: 批量删除, 等价于 dao 层的 DeleteMany
- Store: 存储, 通常是带有特殊的数据处理逻辑，如加密，压缩等
- Load: 加载, 通常是带有特殊的数据处理逻辑，如解密，解压等