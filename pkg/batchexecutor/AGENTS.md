|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/batchexecutor
|OVERVIEW:Generic fixed-size sequential batch execution helpers; owns item slicing, shared timeout context, fail-fast execution, and optional result collection
|BOUNDARY:Infrastructure helper only|callers own business batch size choice, input ordering meaning, retry/backoff, pagination, logging, and side-effect idempotency
|Structure:pkg/batchexecutor:{batch_executor.go,batch_executor_test.go,AGENTS.md}
|WHERE TO LOOK:public contracts:batch_executor.go:{Option,Result,ExecuteFn,CollectFn,WithBatchSize,WithTimeout,Execute,Collect}
|WHERE TO LOOK:batch loop/defaults:batch_executor.go:{defaultBatchSize,defaultTimeout,eachBatch,buildOption,normalizeBatchSize,normalizeTimeout}
|WHERE TO LOOK:expected behavior:batch_executor_test.go:{TestExecute,TestCollect}|split order|empty input|handler error|context cancellation|collection total
|WHERE TO LOOK:current caller pattern:pkg/thirdparty/cmdb/handler.go:findHostBizRelations uses Collect with caller-owned CMDB batch size and timeout
|CONVENTIONS:use contextx.IContext on public callback APIs; derive cancellation/deadline through contextx helpers to preserve tenant/user/message values
|CONVENTIONS:Execute and Collect process batches sequentially in original slice order; preserve fail-fast behavior on context cancellation or callback error
|CONVENTIONS:non-positive WithBatchSize/WithTimeout values normalize to package defaults; keep default values explicit and covered by tests when changed
|CONVENTIONS:Collect appends callback outputs in batch order and sets Result.Total to len(Result.Items), not len(input items)
|CONVENTIONS:callbacks receive sub-slices of the original input; callers needing mutation isolation must copy before passing data or inside callbacks
|ANTI-PATTERNS:do not add goroutine fan-out/fan-in, worker pools, throttling, or shared result mutation here; use/extend pkg/runtime/gopool for concurrent execution semantics
|ANTI-PATTERNS:do not add retry/backoff/polling, partial-success aggregation, or compensation policy; those contracts belong to callers or pkg/runtime/retrier
|ANTI-PATTERNS:do not add service/domain types, storage/network calls, logging, metrics, or internal/* imports to this generic pkg helper
|ANTI-PATTERNS:do not silently continue after callback errors or context cancellation; callers depend on first-error fail-fast semantics
|ANTI-PATTERNS:do not reinterpret Result.Total as requested/input count without updating all callers and tests
|DEPENDENCIES:pkg/contextx only; keep package dependency-light and generic
|COMMANDS:go test ./pkg/batchexecutor
