# pkg/rest/errf — Agent Knowledge Base

## OVERVIEW

`pkg/rest/errf` is the shared error contract for all services.
It defines:
- stable business error codes (`Code`)
- code ↔ base error mapping
- code → HTTP status mapping
- permission payload contract for IAM-style denial responses

This package is cross-service infrastructure. Keep it business-agnostic.

## CORE FILES

```
pkg/rest/errf/
|- code.go         # Code constants and reserved ranges
|- error.go        # CodeErrMap / ErrCodeMap / ErrWrap / ErrUnwrap
|- http_status.go  # Code.HttpStatusCode()
|- permission.go   # PermissionError and response payload structs
|- error_test.go   # unwrap and permission-upgrade behavior tests
`- README.md       # package scope and design notes
```

## BUILD / LINT / TEST COMMANDS

Run from repository root unless noted.

```bash
# all Go tests
go1.25.12 test ./...

# errf package tests
go1.25.12 test ./pkg/rest/errf -v

# single test function in errf (important)
go1.25.12 test ./pkg/rest/errf -run '^TestErrUnwrap_UpgradeToPermissionDeniedWhenPermissionErrorExists$' -v

# rest server tests related to permission response behavior
go1.25.12 test ./pkg/rest/server -run '^TestAbortWithJSONPermDenied' -v

# project lint
make lint
```

## API SURFACE (GREP ANCHORS)

Use these exact signatures to locate behavior:
- `type Code int`
- `func CodeErrMap(code Code) error`
- `func ErrCodeMap(err error) Code`
- `func ErrWrap(code Code, err error) error`
- `func ErrUnwrap(err error) (Code, []error)`
- `func (code Code) HttpStatusCode() int`
- `type PermissionError interface`
- `PermissionData() Permission`

## REQUIRED USAGE PATTERN

### At router/business boundary

- Convert domain errors to standardized code with `resterrf.ErrWrap(code, err)`.
- Return wrapped error to shared REST handler layer.

Typical pattern:

```go
return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
```

### At REST server boundary

- Decode once with `resterrf.ErrUnwrap(err)`.
- Let server response path map code to HTTP + response payload.

Do not re-implement mapping logic in service routers.

## CODE STYLE & CONSTRAINTS

### Error codes

- Follow existing `38xxxxx` code space.
- `3830403` (`PermissionDenied`) is reserved; do not repurpose it.
- Keep default error text lowercase and without trailing punctuation.

### Three-file sync rule (mandatory)

When adding a new code, update all of:
1. `code.go` — add code constant
2. `error.go` — add default code↔error mapping
3. `http_status.go` — add HTTP status mapping

### Wrapping semantics

- `ErrWrap`/`ErrUnwrap` are designed for single-layer semantics.
- Avoid nested wraps like `ErrWrap(A, ErrWrap(B, err))`.
- Avoid `ErrWrap(code, nil)`; it becomes non-nil base error and changes behavior.

### Permission behavior

- If detail errors contain `PermissionError` (detected via `errors.As`), `ErrUnwrap` can upgrade code to `PermissionDenied`.
- Put IAM permission metadata in `PermissionData()`; do not hardcode permission payloads in handlers.

## INTEGRATION NOTES

- `pkg/rest/server/handler.go` is the centralized decode path.
- `pkg/rest/server/request.go` builds JSON error and permission response bodies.
- Use import alias enforced by lint:
  - `resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"`

## ANTI-PATTERNS

- Returning raw low-level errors from API handlers when standard code exists.
- Implementing service-specific policy logic inside `pkg/rest/errf`.
- Changing one mapping file only (causes inconsistent code/error/status behavior).
- Using ad-hoc HTTP status assignment in routers instead of `Code.HttpStatusCode()`.

## VALIDATION CHECKLIST FOR AGENTS

Before finishing errf changes:
1. New code present in `code.go`, `error.go`, `http_status.go`.
2. `go1.25.12 test ./pkg/rest/errf -v` passes.
3. Related server tests (permission/error response path) pass.
4. `make lint` passes for changed Go files.
5. No business-specific rule leaked into shared package.
