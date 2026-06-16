---
name: bk-nodemgr-logger
description: Use when writing, reviewing, or refactoring Go logging in bk-nodemgr, especially code that imports pkg/logger, touches router/service/storage/workflow paths, chooses Biz vs Sys logs, adds WithErr/With/WithDuration fields, adapts third-party logger interfaces, or replaces fmt/log/slog/zap output with the project logger.
---

# bk-nodemgr Logger Skill

## Overview

本 skill 约束 bk-nodemgr 中 `pkg/logger` 的使用方式。它是项目级 logging contract，优先级高于通用 Go `slog`、`zap`、`logrus`
建议。

核心原则：先判断日志属于 `Business` 还是 `System`，再决定 context、字段、级别和打印位置。`pkg/logger`
已经负责文件轮转、标准输出、结构化字段拼装、context values、`err`、`cost`、`trace-id`、`span-id`，调用方不要再造平行日志体系。

## When to Use

使用本 skill，当任务涉及：

- Go 代码新增、修改或审查日志。
- 出现 `pkg/logger`、`logger.G.Sys()`、`logger.G.Biz(ctx)`、`WithErr`、`WithDuration`、`ErrorWriter`。
- 在 `internal/*/router/api-v3`、`internal/*/service`、`internal/*/storage`、`pkg/workflow`、`pkg/scheduler`、`pkg/rest`
  中处理日志。
- 把 `fmt.Print*`、`log.*`、`slog.*`、`zap.*`、`logrus.*` 替换为项目统一日志。
- 给第三方库实现 logger adapter 或 writer bridge。
- 审查日志是否丢失 request context、重复记录错误、字段不成对、级别不当、正常路径噪音过多。

不要用本 skill 处理：

- CLI 启动早期、`logger.Init` 之前必须写到 stdout/stderr 的用户提示或 fatal startup error。参考 `cmd/*` 现有 `fmt.Printf`。
- `pkg/logger` 内部文件管理失败路径。该包内部不能递归使用自身 logger，部分 `fmt.Printf` 是为了避免 deadlock。
- 非 bk-nodemgr 项目的通用 Go logging 设计。那类任务优先使用 Go observability / slog 相关 skill。

## Repo Anchors

先读这些证据，不要凭通用经验改日志：

| 目的                                  | 文件                                                                                             |
|-------------------------------------|------------------------------------------------------------------------------------------------|
| package contract                    | `pkg/logger/README.md`                                                                         |
| public API                          | `pkg/logger/logger.go`, `pkg/logger/iface.go`                                                  |
| file rotation and deadlock boundary | `pkg/logger/file.go`                                                                           |
| usage examples                      | `pkg/logger/example/main.go`                                                                   |
| HTTP request logs                   | `pkg/rest/server/middleware.go`                                                                |
| HTTP client logs                    | `pkg/rest/client/request.go`                                                                   |
| workflow third-party adapter        | `pkg/workflow/logger_adaptor.go`                                                               |
| scheduler third-party adapter       | `pkg/scheduler/logger_adaptor.go`                                                              |
| service initialization              | `cmd/backend/main.go`, `cmd/file/main.go`, `cmd/relay/main.go`, `cmd/application/webserver.go` |

## Logger Model

`pkg/logger` 的调用模型是：

```text
logger.Init(config) -> logger.G -> Sys()/Biz(ctx) -> Ctx/With/WithErr/WithDuration -> Debug/Info/Warn/Error or *Writer
```

关键行为：

- `logger.Init(logger.Config{...})` 使用 `sync.Once`，每个服务入口只初始化一次。
- `logger.G` 默认是 `&Logger{Depth: 2}`，调用栈深度属于 API 语义，不要随意改。
- `Biz(ctx)` 记录 `CategoryBusiness`，并从 `contextx.IContext.Values()` 自动追加上下文字段。
- `Sys()` 记录 `CategorySystem`，必要时可用 `.Ctx(ctx)` 继承 context values。
- `WithErr(err)` 追加 `err` 字段。
- `WithDuration(d)` 追加 `cost=<milliseconds>ms` 字段。
- 当 context 带 OpenTelemetry span 时，logger 自动追加 `trace-id` 和 `span-id`。
- `Flush()` 在服务 shutdown 前落盘，入口通常是 `logger.G.Flush()`。

## Choose Biz or Sys

| 场景                                                                                     | 使用                                          | 原因                                     |
|----------------------------------------------------------------------------------------|---------------------------------------------|----------------------------------------|
| HTTP handler、用户触发 API、callback、一次可闭环请求                                                 | `logger.G.Biz(rCtx)` 或 `logger.G.Biz(nCtx)` | 必须能通过 request/message/context 串起完整业务链路 |
| 用户请求链路派生出的业务 manager/file/relay 操作                                                     | `logger.G.Biz(nCtx)`                        | 仍属于用户触发链路                              |
| service startup/shutdown、storage 初始化、DAO 基础设施、periodic task、scheduler、workflow runtime | `logger.G.Sys()`                            | 后台或通用系统流程，不应制造 business 日志             |
| 系统流程但需要继承 trace/message/tenant 字段                                                      | `logger.G.Sys().Ctx(nCtx)`                  | 语义仍是 System，但保留关联字段                    |
| 第三方库要求 logger interface                                                                | 在引用处写 adapter，内部调用 `logger.G.Sys()`         | 第三方日志属于基础设施边界                          |

判断规则：当系统没有用户操作时不应产生 `Business` 日志。看到 `Biz(nil)` 要谨慎，只有 Gin recovery writer 等兼容场景才合理。

## Level Rules

| Level   | 使用场景                                                                | 注意                      |
|---------|---------------------------------------------------------------------|-------------------------|
| `Debug` | 开发排查、低频细节、第三方 adapter 的 verbose output                              | 生产环境禁止依赖 Debug 输出       |
| `Info`  | 关键状态变化、请求收发、启动关闭、workflow launch、可追踪的成功动作                           | 正常路径不要打无意义流水账           |
| `Warn`  | 不符合预期但流程可继续，如 fallback/default policy、unsupported but skipped state | 不要把真正失败降级为 Warn         |
| `Error` | 当前流程无法继续、返回错误、adapter 中第三方 fatal/panic 映射                           | 用 `WithErr(err)` 记录错误对象 |
| `Fatal` | 禁用                                                                  | 常量存在不代表允许使用             |

## Core Patterns

### Handler error

```go
if err := rCtx.BindJSON(req); err != nil {
logger.G.Biz(rCtx).WithErr(err).Error("failed to update proxy, failed to decode request body")
return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
}
```

### Handler success

```go
logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched proxy upgrade")
```

### System flow with fields

```go
logger.G.Sys().With("workflow-id", workflowID).
Info("successful to execute change action")
```

### System flow with inherited context

```go
logger.G.Sys().Ctx(std.Context()).WithErr(err).With("oper-inst-id", instanceID).
Error("failed to get action private data")
```

### Duration

```go
start := time.Now()
// work
logger.G.Biz(rCtx).WithDuration(time.Since(start)).With("code", gCtx.Writer.Status()).
Info("[request done] %s %s", method, fullPath)
```

### Third-party adapter

```go
func (l *loggerAdaptor) Printf(format string, args ...interface{}) {
logger.G.Sys().Debug(format, args...)
}

func (l *loggerAdaptor) Fatalf(format string, args ...interface{}) {
logger.G.Sys().Error(format, args...)
}
```

适配第三方接口时，在边界转换，不要把第三方 logger 类型扩散到业务代码。

## Message and Field Rules

- 日志 message 使用稳定、完整的 English sentence，例如 `failed to update proxy, failed to update host`。
- 错误对象统一放入 `WithErr(err)`，不要再写 `Error("...: %v", err)`。
- 参数优先放入 `With("key", value)`，而不是拼进 message。
- `With` 的参数必须成对；奇数个参数会产生日志噪音 `ignored key without value`。
- 字段 key 使用项目已有 kebab-case 风格：`workflow-id`、`host-id`、`trigger-id`、`oper-inst-id`、`tenant-id`、`client-ip`、
  `count`、`code`、`cost`。
- 不要手动追加 `trace-id` / `span-id`。context 带 span 时 logger 会自动追加。
- 避免 `please`、`maybe`、`!!!`、`DEBUG:` 等临时或不可检索文本。

## Noise Control

- `pkg/rest/server` 已统一打印 `[request recv]` 和 `[request done]`，handler 不要重复打印“收到请求/请求结束”。
- 正常路径只在关键业务节点打 `Info`，例如 launched workflow、updated host、started service。
- 循环体内不要每次迭代打 `Info`/`Error`，除非是低频且每条都代表独立用户可追踪事件。
- Debug 阶段临时日志提交前删除，或确认只保留 `Debug` 且不会泄露敏感字段。
- 错误日志只在有处理责任的边界打印。底层函数若只是返回错误，应优先 `fmt.Errorf("...: %w", err)`，由上层统一记录。

## Initialization and Shutdown

服务入口遵循现有顺序：parse flags -> load config -> validate -> `logger.Init` -> construct service -> start。

`cmd/*` 在 `logger.Init` 之前使用 `fmt.Printf` 输出启动/配置错误是允许的，因为 logger 尚不可用。服务收到 shutdown signal
后应调用 `logger.G.Flush()`，参考 `cmd/backend/main.go`、`cmd/file/main.go`、`cmd/application/webserver.go`。

不要在业务包中再次调用 `logger.Init`，不要试图用测试或局部逻辑重新配置全局 logger。

## Common Mistakes

| Mistake                                                    | Why it fails                                | Fix                                   |
|------------------------------------------------------------|---------------------------------------------|---------------------------------------|
| 在 bk-nodemgr 业务代码里新增 `slog`/`zap`/`logrus` 调用              | 形成平行日志体系，破坏统一文件、级别、context 和 trace 规则       | 使用 `pkg/logger`；第三方库写 adapter         |
| `logger.G.Biz(rCtx).WithErr(err).Error("failed: %v", err)` | 错误被重复记录，message 不稳定                         | `WithErr(err).Error("failed to ...")` |
| 用户请求链路使用 `Sys()`                                           | 丢失 Business 日志语义和 request/message 可过滤性      | 使用 `Biz(rCtx/nCtx)`                   |
| 后台任务使用 `Biz(nil)`                                          | 制造无 context 的 business 日志                   | 使用 `Sys()` 或 `Sys().Ctx(ctx)`         |
| `With("id")`                                               | key/value 不成对，输出 ignored key 噪音             | `With("id", id)`                      |
| 在 handler 重复打印请求入口/出口                                      | middleware 已经打印，产生噪音                        | 只打印业务关键节点和错误                          |
| 在 `pkg/logger/file.go` 内改用 `logger.G.*`                    | logger 内部递归可能 deadlock                      | 保持当前内部 `fmt.Printf` 边界                |
| 使用 `Fatal`/`Panic` 做控制流                                    | 项目 README 禁止 `FATAL`，第三方 adapter 也降级为 Error | 返回 error 或记录 `Error`                  |

## Verification Checklist

审查或提交前至少跑这些搜索：

```bash
rg -n "fmt\.Print|log\.Print|slog\.|zap\.|logrus\." pkg cmd internal test tools -g '*.go'
rg -n "WithErr\([^\)]*\)\.Error\([^\)]*%[wv].*err" pkg internal cmd test tools -g '*.go'
rg -n "logger\.G\.Biz\(nil\)" pkg internal cmd test tools -g '*.go'
rg -n "Info\(\"DEBUG:|With\(\"[^\"]+\"\)" pkg internal cmd test tools -g '*.go'
```

解释结果时保留已知例外：`cmd/*` 初始化前输出、`pkg/logger` 内部文件管理、测试代码、CLI 用户输出。

如果修改了 Go 文件，继续运行相关包的 diagnostics 和测试。共享基础设施变更优先运行：

```bash
go test ./pkg/logger/...
go test ./pkg/rest/...
go test ./pkg/workflow/...
go test ./pkg/scheduler/...
```

## Eval Prompts

用于验证本 skill 是否有效的真实场景：

1. `internal/backend/router/api-v3/node/proxy/update.go` 里新增错误处理日志，要求符合项目 logger 规范。
2. 审查一个 PR：它把 handler 中的错误写成 `WithErr(err).Error("failed: %v", err)`，判断是否应该改。
3. 某第三方库要求实现 `Printf/Fatalf` logger interface，要求接入 bk-nodemgr 日志体系。
4. 把 `internal/*` 中新增的 `slog.Info` 替换为项目统一日志，并解释哪些 `fmt.Printf` 是例外。
5. 判断一个后台 periodic task 应该用 `Biz(ctx)` 还是 `Sys().Ctx(ctx)`，并给出字段建议。

## Cross-References

- Use `golang-error-handling` for error wrapping and single handling rule.
- Use `golang-observability` only for general observability concepts; in bk-nodemgr logging implementation, this skill
  supersedes generic `slog` migration advice.
- Use `code-review` when the task is a broader PR or commit review.
- Use `api-scaffold` when adding new API handler scaffolding; this skill refines the logging portion.
