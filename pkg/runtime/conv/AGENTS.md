|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/runtime/conv
|OVERVIEW:Portable type-conversion helpers for runtime-only data shaping:{primitive↔string↔number,struct↔map,slice↔map,slice↔slice,emptiness checks,bool coercion}
|OVERVIEW:Boundary=business-agnostic conversion logic only; keep service policy, domain rules, and feature-specific coercion outside this package
|Structure:pkg/runtime/conv:{conv.go,conv_test.go,README.md,AGENTS.md}
|WHERE TO LOOK:core conversions:conv.go:{ToInt64,ToInt64Default,ToString,ToStringDefault,MapToStruct,StructToMap,StructToMapIgnoreError,SliceUnique,SliceIntersect,MapValueToSlice,MapKeyToSlice,SliceToMap,SliceToMapIgnore,IsEmpty,StringToBool,SliceToSlice,SliceToSliceWithError,NumberToBool,NonEmptyOr}
|WHERE TO LOOK:edge cases and semantics:conv_test.go|nil/pointer/overflow/underflow/duplicate-key/order-stability/panic-recovery coverage
|WHERE TO LOOK:package intent and examples:README.md|conversion scope and expected use cases
|CONVENTIONS:Keep this package stdlib-only and dependency-light; do not import internal/\* or add service-specific abstractions
|CONVENTIONS:Preserve existing behavior for nil values, pointer chains, overflow/underflow checks, duplicate-key handling, and recovered panics in map/slice helpers
|CONVENTIONS:Maintain deterministic results where already defined:{MapKeyToSlice sort order|SliceUnique first-seen order|SliceIntersect left-side order}
|CONVENTIONS:Prefer extending the existing helper set over adding parallel converters; keep new helpers generic, reusable, and narrowly scoped
|CONVENTIONS:Keep error messages explicit and stable; when behavior changes, update tests in conv_test.go alongside the code
|ANTI-PATTERNS:Do not add business/domain types, service policy, workflow state, or UI/API-specific conversion rules here
|ANTI-PATTERNS:Do not change public helper semantics or string/error contracts without coordinated caller updates and test coverage
|ANTI-PATTERNS:Do not add logging, configuration loading, storage, network calls, or cross-service imports in this package
|ANTI-PATTERNS:Do not duplicate conversion helpers in callers when the logic is genuinely reusable at runtime scope
