Always use:
- serena for semantic code retrieval and editing tools
- context7 for up to date documentation on third party code
- sequential thinking for any decision making
- use chinese to answer questions about the codebase, but use english for code comments and documentation and technical discussions

## OVERVIEW

BlueKing Node Manager monorepo: Go multi-service backend plus Vue3/TypeScript frontend.
Core runtime domains are split by service (`cmd` + `internal`) and shared platform libraries (`pkg`).

## STRUCTURE

```
bk-nodemgr/
|- cmd/           # service entrypoints (backend/application/file/relay)
|- internal/      # service business/router/manager/storage by domain
|- pkg/           # shared libs, DAO, REST, workflow, thirdparty, proto adapters
|- proto/         # protobuf source definitions
|- front/         # Vue3 + Vite + pnpm frontend
|- tools/         # standalone Go module for installer/tooling binaries
|- install/       # deployment assets (Helm charts, docker artifacts)
|- test/          # integration test harness and mock server
`- docs/          # developer/api/ops/concepts docs index
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Service startup/wiring | `cmd/*` + `internal/*/service` | `cmd/backend/main.go`, `cmd/file/main.go`, `cmd/relay/main.go`, `cmd/application/root.go` |
| API router and handlers | `internal/*/router/api-v3` | Routing is service-scoped under each internal domain |
| Shared business models/types | `pkg/types` | Used as inter-layer payload instead of proto structs |
| Persistence and DAO | `pkg/dao/mongo` + `internal/*/storage` | DAO in `pkg`, service storage orchestration in `internal` |
| Protocol schema source | `proto/**` | Generated targets live in `pkg/proto/**` |
| Proto conversion helpers | `pkg/proto/**` | Package READMEs require proto lifecycle confinement |
| Build and release | `Makefile`, `tools/`, `script_tools/`, `install/` | Cross-arch and image build paths are in root `Makefile` |
| Frontend feature work | `front/src` | API naming follows backend proto naming |
| Integration tests | `test/cases` + `test/mock-server` | Router-level API tests and support mock service |

## CONVENTIONS

- Go toolchain is pinned to `go1.23.10` in root build flow.
- Lint baseline is centralized in `.golangci.yml` (strict, many enabled linters, generated-file rules enabled).
- Public Go functions/types require English comments (project rule).
- Before writing new code, read the relevant module and at least one analogous implementation in the same service or layer.
- Prefer extending an existing code path, helper, or proto conversion over introducing a parallel implementation.
- Use structured logging via `pkg/logger`.
- Frontend package manager is `pnpm` (`front/package.json`, `packageManager: pnpm@9.8.0`).
- Frontend lint extends `@blueking/eslint-config-bk/tsvue3` with import sorting and type-import rules.

## ANTI-PATTERNS (THIS PROJECT)

- Never hand-edit generated `*.pb.go` files (`Code generated ... DO NOT EDIT`).
- Treat proto structs as boundary types; convert to/from `pkg/types` before business logic (`pkg/proto/*/README.md`).
- Do not introduce duplicate helpers or parallel conversion logic before checking whether the same capability already exists in the current service, router package, or `pkg/proto/**`.
- Do not bypass root lint/build entrypoints when changing cross-service behavior.
- Do not place service-specific logic in `pkg` when it belongs in `internal/<service>`.

## UNIQUE STYLES

- `cmd/application` is multi-command CLI style, while `cmd/backend|file|relay` are direct server starters.
- `tools/` is a separate Go module (`tools/go.mod`) with its own lifecycle.
- Root `Makefile` orchestrates binaries, frontend, tests, tools, script packaging, and docker images.

## COMMANDS

```bash
make pre
make all
make lint
make clean

# Proto regenerate
cd proto && make clean && make all

# Frontend local
cd front && pnpm install && pnpm dev

# Integration tests
cd test && make build && make test
```
