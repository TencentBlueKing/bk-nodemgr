|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/logger
|Overview:Shared logging infrastructure for bk-nodemgr; owns global logger initialization, Business/System categories, structured context fields, error/cost fields, file rotation, stderr mirroring, and io.Writer bridges
|Overview:Infrastructure boundary only; callers decide when/what to log, while this package preserves logger API semantics and output formatting consistency
|Parent AGENTS:pkg/AGENTS.md
|Structure:pkg/logger:{logger.go public logger/options/writers,iface.go interfaces,file.go file rotation and symlink handling,blog.go low-level backend glue,README.md package contract,example/main.go usage example}
|Where to look:public API:logger.go:{Init,Logger,Sys,Biz,Flush,Option,Ctx,With,WithErr,WithDuration,AssignWhenLogging,Debug/Info/Warn/Error,*Writer}
|Where to look:interface contract:iface.go:{ILogger,ILoggerOption,ILoggerPrinter}:keep exported method meaning stable for callers and adapters
|Where to look:file output:file.go|blog.go:rotation, max size/num, category/level naming, symlink updates, stderr routing
|Where to look:usage contract:README.md:Business vs System decision, WithErr/With/WithDuration, message and level rules
|Where to look:service initialization:cmd/{backend,file,relay}/main.go|cmd/application/webserver.go:logger.Init once per process and logger.G.Flush on shutdown
|Where to look:adapters/examples:pkg/{workflow,scheduler}/logger_adaptor.go|pkg/logger/example/main.go|pkg/rest/server/middleware.go|pkg/rest/client/request.go
|Conventions:Preserve `logger.Init(config)` single-process global setup via sync.Once; service entrypoints initialize, business packages must not reconfigure global logger
|Conventions:Preserve `logger.G` default caller depth and category semantics; caller-depth changes affect file/line attribution and are API-visible
|Conventions:`Biz(ctx)` is for user-triggered closed-loop request chains and should carry contextx values; `Sys()` is for startup/background/infrastructure flows; `Sys().Ctx(ctx)` keeps System category while inheriting context fields
|Conventions:`WithErr(err)` owns the `err` field; `WithDuration(d)` owns `cost=<milliseconds>ms`; `With(k,v,...)` fields should remain stable key/value pairs and tolerate existing ignored-key behavior
|Conventions:Context parsing must keep contextx values before explicit fields, then err/cost, then OpenTelemetry trace-id/span-id; do not duplicate trace fields in callers
|Conventions:Messages should be stable English text; keep formatting rules compatible with existing log search and file rotation outputs
|Conventions:Third-party logger integrations belong at the importing package boundary and should call this package; do not make pkg/logger depend on third-party adapters
|Anti-patterns:Do not add service-specific policy, parameter validation, domain decisions, or request lifecycle logging rules into pkg/logger
|Anti-patterns:Do not introduce parallel logging systems (`log`, `slog`, `zap`, `logrus`) for bk-nodemgr runtime code; route through pkg/logger unless startup-before-init or logger-internal failure boundary applies
|Anti-patterns:Do not recursively use `logger.G.*` inside low-level file/backend failure handling where logger internals are already active; keep explicit non-recursive fallback output when needed
|Anti-patterns:Do not log errors both as formatted message args and `WithErr(err)`; do not rely on `Fatal` for control flow even though the level constant exists
|Anti-patterns:Do not change category/level file naming, rotation limits, symlink behavior, or additionMessage format casually; these are operational compatibility surfaces
|Verification:go1.23.10 test ./pkg/logger/...
