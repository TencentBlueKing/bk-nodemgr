# CMDB MOCK KNOWLEDGE BASE

## OVERVIEW

`cmdb/` provides a mock implementation of CMDB APIs for integration testing. It simulates CMDB's core functionality including business management, cloud areas, hosts, object attributes, and watch events.

## WHERE TO LOOK

- Route registration: `router.go` - Add routes in `Load` function, follow `/cmdb/api/v3/...` pattern
- HTTP handlers: `handler.go` - Use `respondSuccess`/`respondError` from `helper.go`
- Storage logic: `storage.go` - Thread-safe operations, use `conv` package for conversions
- Config types: `types.go` - Add config types and validation in `MockData`
- Response format: `helper.go` - CMDB-specific response wrapper functions
- Unit tests: `storage_test.go` - Test storage layer methods

## DESIGN INTENTIONS

- **Component isolation**: Each mock component (cmdb, bkrepo, etc.) is independent with separate response formats
- **CMDB API compatibility**: Mock responses follow CMDB `BaseBroker[T]` format; error codes match CMDB conventions
- **Thread-safe storage**: In-memory storage initialized once during `Load`, shared across handlers with proper locking
- **Default object attributes**: When config is nil or `objectAttributes` is empty, storage loads default os (`bk_os_type`), cpu (`bk_cpu_architecture`) and cloud (`bk_cloud_vendor`) enum options so integration tests work without explicit mock data config

## CONVENTIONS

- **Thread-safe storage**: All operations use `sync.RWMutex`; read with `RLock()`, write with `Lock()`
- **Error handling**: Storage methods return `(result, error)`; handlers use `CodeInvalidParameter` for client errors, `CodeServerError` for server errors
- **Request/Response**: Use `common.BindJSON` for binding, `respondSuccess`/`respondError` for responses (CMDB `BaseBroker[T]` format)
- **Data conversion**: Use `pkg/runtime/conv` package (prefer `conv.SliceToSlice`); use constants for magic strings (e.g., `enumOptionIDKey`, `enumOptionNameKey`)
- **Storage consistency**: Return empty slices `[]T{}` instead of `nil`
- **Code organization**: Group related handlers with comment separators; keep handlers focused on HTTP concerns

## ANTI-PATTERNS

- Do not use `pkg/types` business types; use `pkg/thirdparty/cmdb` types instead
- Do not bypass `respondSuccess`/`respondError`; always use CMDB response format
- Do not hardcode strings; use constants for API field names
- Do not access storage without proper locking; always use mutex
- Do not return `nil` slices; return empty slices `[]T{}` for consistency
