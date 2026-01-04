# API 接口开发流程（从 DAO 到 Application）

本文档记录基于 `plugin/list` 接口的完整开发流程，适用于所有类似的 API 接口开发。

## 📋 开发顺序（自底向上）

```
DAO 层 → Storage 层 → Backend Router 层 → Thirdparty 层 → Application Router 层
```

## 🔄 完整调用链路

```
客户端请求
    ↓
Application Router Handler
    ↓ 调用
Thirdparty Backend Client
    ↓ HTTP 调用
Backend Router Handler  
    ↓ 调用
Backend Storage
    ↓ 调用
DAO Handler
    ↓ 调用
MongoDB 数据库
```

---

## 第一步：DAO 层（数据访问层）

### 1.1 编写 Options 函数
**文件：** `pkg/dao/mongo/{module}/options.go`

**内容：**
- 定义过滤选项函数，如：
  - `WithName()` - 精确匹配
  - `WithGroup()` - 精确匹配
  - `WithFuzzyName()` - 模糊匹配
  - `WithFuzzyPkgName()` - 模糊匹配

**示例：**
```go
// WithName filters by name.
func WithName(names ...string) OptFn {
    return base.WithValues(FieldKeyName, names...)
}

// WithFuzzyName filters by fuzzy name.
func WithFuzzyName(names ...string) OptFn {
    return base.WithFuzzyValues(FieldKeyName, names...)
}
```

### 1.2 编写 DAO Handler
**文件：** `pkg/dao/mongo/{module}/handler.go`

**内容：**
- 在 `IHandler` 接口中定义方法：
  - `Count(nCtx, opts...) (int64, error)`
  - `List(nCtx, page, opts...) ([]*types.{Module}, int64, error)`
- 实现方法：
  - 构建 `base.AliveFilter()`
  - 应用 `opts` 选项函数
  - 调用 `tenantDao().Count()` 或 `tenantDao().List()`
  - 转换数据模型为 `types.{Module}`

**关键点：**
- 使用 `tenantDao(nCtx.TenantID())` 获取租户特定的 DAO
- 使用 `convert{Module}ToTypes()` 转换数据

---

## 第二步：Backend Storage 层

### 2.1 定义 Storage 接口
**文件：** `internal/backend/storage/{module}/iface.go`

**内容：**
- 在 `IDao{Module}` 接口中定义方法：
  - `Count{Modules}(nCtx, conditions...) (int64, error)`
  - `List{Modules}(nCtx, page, conditions...) ([]*types.{Module}, int64, error)`

**示例：**
```go
type IDaoPlugin interface {
    CountPlugins(nCtx contextx.IContext, conditions ...*types.PluginCondition) (int64, error)
    ListPlugins(nCtx contextx.IContext, page types.Page, conditions ...*types.PluginCondition) ([]*types.Plugin, int64, error)
}
```

### 2.2 实现 Storage DAO 转换
**文件：** `internal/backend/storage/{module}/dao_{module}.go`

**内容：**
- 实现 `count{Modules}()` - 私有方法，转换 condition 为 options，调用 dao
- 实现 `list{Modules}()` - 私有方法，转换 condition 为 options，调用 dao
- 实现 `convert{Module}ConditionToOptions()` - 将 `types.{Module}Condition` 转换为 `dao{Module}.OptFn` 列表

**关键转换逻辑：**
```go
func convertPluginConditionToOptions(conditions []*types.PluginCondition) ([]daoPlugin.OptFn, error) {
    opts := make([]daoPlugin.OptFn, 0)
    for _, condition := range conditions {
        if condition.ExactInclude != nil {
            opts = append(opts,
                daoPlugin.WithName(condition.ExactInclude.Name...),
                daoPlugin.WithGroup(condition.ExactInclude.Group...),
            )
        }
        if condition.FuzzyInclude != nil {
            opts = append(opts,
                daoPlugin.WithFuzzyName(condition.FuzzyInclude.Name...),
                daoPlugin.WithFuzzyPkgName(condition.FuzzyInclude.PkgName...),
            )
        }
    }
    return opts, nil
}
```

### 2.3 实现 Storage 公共方法
**文件：** `internal/backend/storage/{module}/storage.go`

**内容：**
- 在 `initDao()` 中初始化 dao：`s.dao{Module} = {module}.New(s.Database)`
- 实现公共方法（包装 metric）：
  - `Count{Modules}()` - 调用私有 `count{Modules}()`
  - `List{Modules}()` - 调用私有 `list{Modules}()`

**示例：**
```go
func (s *Storage) CountPlugins(nCtx contextx.IContext, conditions ...*types.PluginCondition) (count int64, err error) {
    metric := s.metric().Start("count_plugins")
    defer metric.End(err)
    
    count, err = s.countPlugins(nCtx, conditions...)
    return count, err
}
```

---

## 第三步：Backend Router 层

### 3.1 定义 Backend Proto
**文件：** `proto/backend/api/v3/{module}.proto`

**内容：**
- 定义 `{Module}ListReq` message：
  - `Page page = 1;`
  - `bool only_count = 2;`
  - `ExactConditions exact_include_conditions = 3;`
  - `FuzzyConditions fuzzy_include_conditions = 4;`
- 定义 `{Module}ListResp` message：
  - `int32 code = 1;`
  - `string message = 2;`
  - `Data data = 5;` (包含 total 和 items)
- 在 service 中定义 RPC：`rpc List{Modules}({Module}ListReq) returns ({Module}ListResp)`

### 3.2 编写 Backend Proto 扩展
**文件：** `pkg/proto/backend/api/v3/{module}.go`

**内容：**
- `{Module}ListReq.ConvertConditionsToTypes()` - proto → types
- `{Module}ListReq.ConvertPageToTypes()` - proto page → types page
- `{Module}ListResp.ConvertPluginFromTypes()` - types → proto
- `{Module}ListResp.Convert{Module}ToTypes()` - proto → types（用于 thirdparty）

### 3.3 实现 Backend Handler
**文件：** `internal/backend/router/api-v3/{module}/list.go`

**内容：**
- 实现 `handler.List()` 方法：
  - 绑定请求：`req := new(protoBackend.{Module}ListReq)`
  - `rCtx.BindJSON(req)`
  - 根据 `only_count` 调用 `dao{Module}.Count{Modules}()` 或 `List{Modules}()`
  - 构建响应：`resp := new(protoBackend.{Module}ListResp)`
  - `resp.Convert{Module}FromTypes(cnt, items)`
  - 返回 `resp.GetData()`

### 3.4 注册 Backend 路由
**文件：** `internal/backend/router/api-v3/{module}/{module}.go`

**内容：**
- 在 `newHandler()` 中初始化：`dao{Module}: capability.Storage{Module}`
- 在 `Load()` 中注册：`h.rg.POST("/list", restserver.Handler(h.List))`

---

## 第四步：Thirdparty Backend 层

### 4.1 定义 Thirdparty 接口
**文件：** `pkg/thirdparty/backend/handler.go`

**内容：**
- 在 `IHandler{Module}` 接口中定义：
  - `List{Modules}(ctx, page, condition) ([]*types.{Module}, int64, error)`
  - `Count{Modules}(ctx, condition) (int64, error)`

### 4.2 实现 Thirdparty Client
**文件：** `pkg/thirdparty/backend/{module}.go`

**内容：**
- 实现 `Count{Modules}()`：
  - 创建 `protoBackend.{Module}ListReq`
  - `req.ConvertConditionFromTypes(condition)`
  - `req.OnlyCount = true`
  - 调用 `h.cli.list{Modules}(ctx, req)`
  - 返回 `resp.GetData().GetTotal()`
- 实现 `List{Modules}()`：
  - 创建 `protoBackend.{Module}ListReq`
  - `req.ConvertConditionFromTypes(condition)`
  - `req.Page = convertPage(page)`
  - `req.OnlyCount = false`
  - 调用 `h.cli.list{Modules}(ctx, req)`
  - `items, total := resp.Convert{Module}ToTypes()`
  - 返回 `items, total, nil`

---

## 第五步：Application Router 层

### 5.1 定义 Application Proto
**文件：** `proto/application/api/v3/{module}.proto`

**内容：**
- 定义 `{Module}ListReq` message（与 backend 类似）
- 定义 `{Module}ListResp` message（与 backend 类似，但可能不包含 Plugin 内部结构）
- 在 `{Module}API` service 中定义 RPC

### 5.2 编写 Application Proto 扩展
**文件：** `pkg/proto/application/api/v3/{module}.go`

**内容：**
- `{Module}ListReq.ConvertConditionsToTypes()` - proto → types
- `{Module}ListReq.ConvertPageToTypes()` - proto page → types page
- `{Module}ListResp.Convert{Module}FromTypes()` - types → proto

### 5.3 实现 Application Handler
**文件：** `internal/application/router/api-v3/{module}/list.go`

**内容：**
- 实现 `handler.List()` 方法：
  - 绑定请求：`req := new(protoApplication.{Module}ListReq)`
  - `rCtx.BindJSON(req)`
  - 根据 `only_count` 调用 `backendHandler.Count{Modules}()` 或 `List{Modules}()`
  - 构建响应：`resp := new(protoApplication.{Module}ListResp)`
  - `resp.Convert{Module}FromTypes(cnt, items)`
  - 返回 `resp.GetData()`

### 5.4 注册 Application 路由
**文件：** `internal/application/router/api-v3/{module}/{module}.go`

**内容：**
- 在 `newHandler()` 中初始化：`backendHandler: capability.BackendHandler`
- 在 `Load()` 中注册：`h.rg.POST("/list", restserver.Handler(h.List))`

---

## 📝 关键要点总结

### 1. 数据流转
- **请求方向：** Proto → Types → Options → MongoDB Filter
- **响应方向：** MongoDB Data → Types → Proto → JSON

### 2. 转换方法命名规范
- `Convert{From}To{To}()` - 如 `ConvertConditionsToTypes()`
- `Convert{From}From{To}()` - 如 `ConvertConditionFromTypes()`

### 3. 错误处理
- Application/Backend Handler：使用 `resterrf.ErrWrap()` 包装错误
- Storage 层：使用 `fmt.Errorf()` 包装错误信息
- DAO 层：直接返回数据库错误

### 4. Metric 记录
- Storage 层的公共方法需要记录 metric
- 使用 `s.metric().Start("operation_name")` 和 `defer metric.End(err)`

### 5. 条件转换
- `types.{Module}Condition` → `dao{Module}.OptFn[]` 在 Storage DAO 层完成
- 支持 `ExactInclude` 和 `FuzzyInclude` 两种条件类型

### 6. 分页处理
- 使用 `types.Page` 统一分页结构
- `maxLimit` 参数控制最大返回数量（0 表示无限制）

---

## 🎯 快速检查清单

开发新接口时，按以下顺序检查：

- [ ] DAO 层：Options 函数已定义
- [ ] DAO 层：Handler 接口和实现已完成
- [ ] Storage 层：接口定义已添加
- [ ] Storage 层：DAO 转换方法已实现
- [ ] Storage 层：公共方法已实现（含 metric）
- [ ] Backend Proto：消息定义已完成
- [ ] Backend Proto：扩展方法已实现
- [ ] Backend Router：Handler 已实现
- [ ] Backend Router：路由已注册
- [ ] Thirdparty：接口已定义
- [ ] Thirdparty：实现已完成
- [ ] Application Proto：消息定义已完成
- [ ] Application Proto：扩展方法已实现
- [ ] Application Router：Handler 已实现
- [ ] Application Router：路由已注册

---

## 📚 参考示例

完整实现参考：
- `internal/application/router/api-v3/plugin/list.go`
- `internal/backend/router/api-v3/plugin/list.go`
- `internal/backend/storage/plugin/dao_plugin.go`
- `pkg/dao/mongo/plugin/handler.go`

