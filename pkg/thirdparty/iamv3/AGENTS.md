# IAMV3 KNOWLEDGE BASE

## OVERVIEW

`iamv3` encapsulates BlueKing IAM v3 integration for Node Manager:
- permission checks (single action, batch, multi-action),
- IAM system token retrieval,
- callback BasicAuth validation,
- permission apply URL generation.

The package exposes `IHandler` as the stable entry point.
`Handler` is the real IAM-backed implementation, and `NoOpHandler` is the explicit bypass mode used when IAM v3 is disabled.

## WHERE TO LOOK

- Handler interface and business-facing methods: `pkg/thirdparty/iamv3/handler.go`
- IAM HTTP API wiring and endpoint paths: `pkg/thirdparty/iamv3/iamv3.go`
- Request/response wire models + validation + cache key helpers: `pkg/thirdparty/iamv3/types.go`
- In-memory cache for permission checks: `pkg/thirdparty/iamv3/cache.go`
- IAM disabled bypass behavior: `pkg/thirdparty/iamv3/noop.go`
- Integration points:
  - Handler initialization: `internal/backend/service/service.go` (`newIAMV3Handler`)
  - IAM callback auth middleware: `internal/backend/router/api-v3/iam/v3/v3.go`

## CONVENTIONS

- Keep IAM-specific payload shapes (`Request`, `MultiActionRequest`, `Application`, etc.) confined to this package.
- Keep all IAM endpoint path assembly in `iamv3.go`; callers should only use `IHandler`.
- Run request validation (`Validate`) before issuing IAM API calls.
- Use constant-time comparison for secret checks in authentication logic.
- Keep permission cache behavior centralized in `IsAllowedWithCache`; do not duplicate cache logic in callers.
- Preserve structured and wrapped errors so callers can trace IAM failures.

## ANTI-PATTERNS

- Do not call low-level `cli` methods from outside this package.
- Do not spread IAM endpoint literals (for example `/v1/model/systems/...`) into upper layers.
- Do not bypass `IHandler` by constructing IAM request payloads in business modules.
- Do not treat `NoOpHandler` as a partial-security mode: it is a full authorization bypass.
- Do not log raw secrets (token/password/auth headers) in errors or debug output.

## TESTING

- Package-level tests: `go test ./pkg/thirdparty/iamv3`
- Focused regression checks:
  - `go test ./pkg/thirdparty/iamv3 -run 'TestIsBasicAuthAllowed'`
  - `go test ./pkg/thirdparty/iamv3 -run 'TestCacheKeyEquivalence|TestBatchResourceIDKey'`
  - `go test ./pkg/thirdparty/iamv3 -run 'TestNewNoOpHandler'`

