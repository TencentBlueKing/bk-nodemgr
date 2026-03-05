# pkg/rest — Agent Knowledge Base

## OVERVIEW

Shared RESTful infrastructure for all bk-nodemgr services. Provides standardized HTTP server (gin-based),
HTTP client with retry/tracing, unified error codes, JWT auth middleware, request context propagation,
and Prometheus metrics collection. This package is business-agnostic — no domain logic belongs here.

## STRUCTURE

```
pkg/rest/
|- client/       # HTTP client: retry, tracing, request builder, metrics
|- context/      # IRequestIDContext interface (context.Context + RequestID)
|- discovery/    # Service endpoint discovery (static + etcd-backed)
|- errf/         # Error codes (38xxxxx range), Code↔error maps, Code→HTTP status
|- header/       # BlueKing-specific HTTP headers (tenant, auth JWT, request-id)
|- metrics/      # Prometheus monitor: request count, duration, body size, slow reqs
`- server/       # Gin server: handler wrappers, middleware chain, request/response types
```

## BUILD / LINT / TEST COMMANDS

```bash
# Full build (from repo root)
make pre && make all

# Lint (strict, 60+ linters — see root .golangci.yml)
make lint

# Run ALL Go tests
go1.23.10 test ./...

# Run tests for this package only
go1.23.10 test ./pkg/rest/...

# Run a single test by name
go1.23.10 test ./pkg/rest/... -run TestFunctionName -v

# Run tests with race detector
go1.23.10 test -race ./pkg/rest/...

# Frontend (unrelated to this pkg, for reference)
cd front && pnpm install && pnpm dev
```

> **Note**: Go toolchain is pinned to `go1.23.10`. Always use this version.
> There are currently no test files in `pkg/rest/`. When adding tests, place them alongside source as `*_test.go`.

## CODE STYLE

### File Header

Every `.go` file must start with the TencentBlueKing MIT license block:

```go
/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); ...
 */
```

### Package Comments

Every package declaration requires a `// Package <name> ...` comment ending with a period.
Declaration-scope comments must end with a period (enforced by `godot` linter).

```go
// Package server is the restful API server.
package server
```

### Imports

Three groups separated by blank lines: stdlib, external, internal. Enforced aliases via `importas` linter:

```go
import (
    "fmt"                           // 1. stdlib

    "github.com/gin-gonic/gin"      // 2. external dependencies

    resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"       // 3. internal
    restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
    restmetrics "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/metrics"
    restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
    restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
    restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
    contextx "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/contextx"
)
```

The lint config forbids unaliased imports when an alias is defined — check `.golangci.yml` `importas.alias` list.

### Naming

- **Interfaces**: Prefix with `I` (e.g. `IContext`, `IRequest`, `IClient`, `IAuthIdentity`).
- **Exported types/funcs**: English doc comment required, ending with period.
- **Error codes**: `Code` type (int), 38-prefix + 5 digits (e.g. `3800001`). Sentinel errors prefixed with `Err`.
- **Constants**: Grouped by `const ()` blocks with doc comments per constant.
- **Struct fields**: Doc comment on each exported field.
- **Variable names**: min length 2 chars; single-letter `r`, `n`, `err`, `p`, `f`, `x`, `k`, `v` are allowed.
- **Options pattern**: Use `Opt`/`OptFunc`/`OptFn` type for functional options (e.g. `WithHeaderMasker`, `WithSlowTime`).

### Functions

- Max 120 lines (excluding comments). Max cyclomatic/cognitive complexity: 20.
- Max line length: 150 characters.
- No naked returns; no named returns (enforced by `nakedret`, `nonamedreturns` — use `nolint` only when unavoidable).
- Blank line before return when block size > 2 (`nlreturn`).
- Newline after multi-line `if` and multi-line function signature (`whitespace`).

### Error Handling

- Use `errf.ErrWrap(code, err)` to wrap errors with a code — single-layer only.
- Use `errf.ErrUnwrap(err)` to extract `Code` + detail errors.
- New error codes must be added to **both** `errf/code.go` (constant) and `errf/error.go` (init map) and `errf/http_status.go` (HTTP mapping).
- Error messages: lowercase, no punctuation at end. Wrap with `fmt.Errorf("context: %w", err)`.
- Never use empty `catch(e) {}` blocks. Every error must be checked.

### Interfaces

- Max 10 methods per interface (`interfacebloat`).
- Interface compliance verified with `var _ IFoo = &Foo{}`.
- Request bodies implement `RequestBody` interface: `Validate() error` + `AutoConvert()`.

### Handler Pattern

Route handlers use typed wrapper functions that convert gin handlers:

```go
// Standard JSON response handler.
restserver.Handler(func(rCtx server.IContext) (interface{}, error) { ... })

// Stream response handler.
restserver.StreamHandler(func(rCtx server.IContext) (*server.StreamResponse, error) { ... })

// File download handler.
restserver.FileHandler(func(rCtx server.IContext) (*server.FileResponse, error) { ... })
```

Handlers return `(data, nil)` for success or `(nil, errf.ErrWrap(code, err))` for errors.
The wrapper automatically maps the error code to HTTP status and formats the JSON response.

### Response Format

All JSON responses follow this structure:

```json
{"code": 0, "message": "OK", "request_id": "...", "data": {...}, "error": null, "permission": null}
```

Error responses include `error.system`, `error.message`, and `error.details[]`.

### Logging

Use `pkg/logger` — never `fmt.Println` or `log.*`.

```go
logger.G.Biz(rCtx).With("key", value).Info("message %s", arg)
```

## ANTI-PATTERNS

- Do NOT place business/domain logic in `pkg/rest` — it belongs in `internal/<service>`.
- Do NOT suppress lint errors with `as any` / `@ts-ignore` equivalents; use targeted `// nolint: <linter>` with reason.
- Do NOT add standard HTTP headers to `pkg/rest/header` — only BlueKing-specific headers.
- Do NOT add third-party system headers (e.g. apigw) to `pkg/rest/header`.
- Do NOT use `init()` functions outside of error code registration.

## KEY PATTERNS

- **Middleware chain order**: Recovery → Context init → Tracing → RequestID → Received log → Returned log → Metrics.
- **Client retry**: Max 3 cycles, only retries GET on `ECONNRESET`, 20ms delay between retries.
- **Metrics**: Both server and client sides; prefixed with `server_`/`client_` + service name. Tracks request count, UV, body size, duration histogram, slow requests.
- **Discovery**: Static endpoints (round-robin) or service-based (etcd discover).
- **Auth**: JWT-based via `X-Bknodemgr-Authorization` header; `NoneAuthIdentity` for dev/test.
