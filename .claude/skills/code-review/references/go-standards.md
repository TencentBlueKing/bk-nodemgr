# Go 项目规范速查

本文档提供 bk-nodemgr 项目的 Go 代码规范快速参考。

## 数据转换规范

**来源**: `.cursor/rules/conv.mdc`

### 优先使用 `pkg/runtime/conv` 包

**基本类型转换：**
- 转换为 int64: `conv.ToInt64()` 或 `conv.ToInt64Default()`
- 转换为 string: `conv.ToString()` 或 `conv.ToStringDefault()`
- 字符串转布尔值: `conv.StringToBool()`
- 数字转布尔值: `conv.NumberToBool()`

**结构体和映射转换：**
- map 转 struct: `conv.MapToStruct()`
- struct 转 map: `conv.StructToMap()` 或 `conv.StructToMapIgnoreError()`

**切片操作：**
- 切片去重: `conv.SliceUnique()`
- 切片转换: `conv.SliceToSlice()` 或 `conv.SliceToSliceWithError()`
- 切片转 map: `conv.SliceToMap()`
- map 值转切片: `conv.MapValueToSlice()`
- **map 键转切片（自动排序）**: `conv.MapKeyToSlice()`

**工具函数：**
- 检查值是否为空: `conv.IsEmpty()`

### 使用原则

1. ✅ **优先使用**: 当 conv 包提供了合适的转换函数时
2. ✅ **可以使用标准库**: 如果标准库更合适、更高效
3. ❌ **不要强行使用**: 如果需要特殊的转换逻辑
4. ⚠️ **注意错误处理**: 处理转换函数返回的错误

## 错误处理规范

**项目 owner**: `bk-nodemgr-error-handling`。Code review 中的错误创建、传播、包装、REST 映射、log-or-return 责任边界和 `err == nil` 迁移都以该 skill 为准。

### 标准模式

```go
// ✅ 正确：错误路径显式，内部层保留错误链
if err := someFunc(); err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

Router/API 边界使用 `resterrf.ErrWrap(code, err)` 映射统一 REST 错误码；日志边界使用 `logger.G.*().WithErr(err).Error("stable message")`，不要再把同一个 `err` 格式化进 message。

### 绝对禁令

- 禁止在新增或修改的 Go 代码中出现 `err == nil`。历史存量是 migration debt，不是可复制先例。
- 禁止用 `isNilErr(err)`、`ok := err == nil`、重命名变量、`nolint` 或 helper 马甲绕过。
- 禁止 `fmt.Errorf("...: %v", err)` / `fmt.Errorf("...: %s", err)` 丢失错误链。
- 禁止同一条 error chain 在多个层级重复 `ERROR` 日志。

### 检查要点

- 所有错误都应被处理、明确忽略或带上下文传播。
- 内部包装使用 `%w`，错误检查使用 `errors.Is` / `errors.As`。
- 当前层只添加当前层上下文，不重复包装 REST code。
- API handler 使用 `resterrf.ErrWrap` 做 transport mapping。
- Lower layer 优先返回 wrapped error，由有处理责任的边界决定是否记录日志。
- 遇到 `err == nil` 命中时，先按 `bk-nodemgr-error-handling` 的 migration catalog 分类，再迁移，不能机械替换。

## 日志记录规范

### 标准格式（两行）

```go
logger.G.Sys().With("key1", value1).With("key2", value2).
    Info("message")
```

### 检查项

- 使用结构化日志（`With` 方法添加上下文）
- `With` 和 `Info`/`Error` 等方法分开在不同行
- 日志消息清晰描述操作结果

### 示例

```go
// ✅ 正确
logger.G.Sys().With("workflow-id", workflowID).With("task-count", len(tasks)).
    Info("successful to execute change action")

// ❌ 错误：单行格式
logger.G.Sys().With("workflow-id", workflowID).Info("successful to execute change action")

// ❌ 错误：没有使用结构化日志
logger.G.Sys().Info(fmt.Sprintf("workflow %s completed", workflowID))
```

## 命名规范

**来源**: `.golangci.yml`

### 通用规则

- 错误类型以 `Error` 结尾（如 `NotFoundError`）
- 错误变量以 `Err` 开头（如 `ErrNotFound`）
- 接口命名简洁明了
- 使用统一的导入别名

### 常见导入别名

```go
import (
    logpb "github.com/xxx/bk-nodemgr/proto/log"
    nodepb "github.com/xxx/bk-nodemgr/proto/node"
)
```

## 注释规范

### 要求

- 公共函数和类型必须有注释
- 注释以句号结尾
- 使用中文注释（项目主要使用中文）
- 注释应解释"为什么"而非"做什么"

### 示例

```go
// ✅ 正确：解释原因
// CreateNode 创建节点。
// 使用 goroutine 异步创建以避免阻塞主流程。
func CreateNode(ctx context.Context) error {
    ...
}

// ❌ 错误：只描述做什么
// CreateNode 创建一个节点
func CreateNode(ctx context.Context) error {
    ...
}
```

## API 开发规范

**来源**: `docs/api/API接口开发流程.md`

### 关键检查项

- 遵循 REST API 设计原则
- 使用统一的错误响应格式（`pkg/rest/errf`）
- 正确映射 HTTP 状态码
- 实现认证和授权检查

### 错误处理

```go
// 使用 pkg/rest/errf 返回标准错误
if err := validate(req); err != nil {
    return errf.New(errf.InvalidParameter, "invalid request: %v", err)
}
```

### HTTP 状态码映射

- 200: 成功
- 400: 客户端错误（Invalid Parameter）
- 401: 未认证
- 403: 无权限
- 404: 资源不存在
- 500: 服务器错误

## Proto 文件规范

**来源**: `proto/README.md`

### 命名规范

- Service 名称使用 PascalCase
- Message 名称使用 PascalCase
- Field 名称使用 snake_case

### 字段编号

- 1-15: 频繁使用的字段（1字节编码）
- 16-2047: 普通字段（2字节编码）
- 保留已删除的字段编号
- 避免使用 19000-19999（Proto 保留）

### 注释要求

```protobuf
// ✅ 正确：完整的注释
// CreateNodeRequest 创建节点请求
message CreateNodeRequest {
    // node_name 节点名称，必填
    string node_name = 1;
    // labels 节点标签，可选
    map<string, string> labels = 2;
}
```

## 代码质量检查清单

### 必须检查项

- [ ] 所有错误都已检查和处理
- [ ] 使用 `fmt.Errorf` 和 `%w` 包装错误
- [ ] 使用结构化日志
- [ ] 日志使用两行格式
- [ ] 使用 `pkg/runtime/conv` 进行数据转换
- [ ] 公共函数有完整注释
- [ ] 遵循项目命名规范

### 性能检查项

- [ ] 避免不必要的循环
- [ ] 避免重复的类型转换
- [ ] 合理使用缓存
- [ ] 及时关闭资源（文件、连接等）

### 安全检查项

- [ ] 没有 SQL 注入风险
- [ ] 没有 XSS 风险
- [ ] 敏感信息已脱敏
- [ ] 输入已验证
- [ ] 权限检查完整

### 并发检查项

- [ ] 没有竞态条件
- [ ] 没有死锁风险
- [ ] goroutine 正确退出
- [ ] 共享资源有适当保护

## 规范文档索引

- **项目规范**: `AGENTS.md`、`CLAUDE.md`
- **API 规范**: `docs/api/API接口开发流程.md`
- **Proto 规范**: `proto/README.md`
- **开发规范**: `docs/developer/README.md`
- **概念文档**: `docs/concepts/README.md`
- **代码转换**: `.cursor/rules/conv.mdc`
- **Linter 配置**: `.golangci.yml`
- **模块规范**: `../patterns/`
