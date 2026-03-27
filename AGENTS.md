|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/comments/docs/tech discussions
|Scope:repo-root
|Overview:BlueKing Node Manager monorepo; Go multi-service backend + Vue3/TypeScript frontend|runtime domains split by service in cmd/internal and shared libraries in pkg
|Structure:./:{cmd,internal,pkg,proto,front,tools,install,test,docs}
|Where to look:service startup/wiring:cmd/*|internal/*/service:{cmd/backend/main.go,cmd/file/main.go,cmd/relay/main.go,cmd/application/root.go}
|Where to look:API routers/handlers:internal/*/router/api-v3:routing is service-scoped
|Where to look:shared business models:pkg/types:inter-layer payloads instead of proto structs
|Where to look:persistence/DAO:pkg/dao/mongo|internal/*/storage:DAO in pkg, service orchestration in internal
|Where to look:proto source:proto/**:generated targets live in pkg/proto/**
|Where to look:proto converters:pkg/proto/**:proto lifecycle confined here
|Where to look:build/release:./:{Makefile,tools,script_tools,install}:root Makefile drives binaries/frontend/tests/tools/images
|Where to look:frontend features:front/src:API naming follows backend proto naming
|Where to look:integration tests:test/{cases,mock-server}:router-level API tests and support mocks
|Conventions:toolchain=go1.23.10|lint=.golangci.yml strict baseline|public Go functions/types require English comments
|Conventions:before coding read relevant module + at least one analogous implementation in same service/layer
|Conventions:prefer extending existing code paths/helpers/proto conversions over parallel implementations
|Conventions:use pkg/logger for structured logging
|Conventions:frontend package manager=pnpm@9.8.0|eslint extends @blueking/eslint-config-bk/tsvue3 with import sorting and type-import rules
|Anti-patterns:never hand-edit generated *.pb.go|treat proto structs as boundary types and convert via pkg/proto/* before business logic|do not introduce duplicate helpers or parallel conversion logic before checking current service/router/pkg/proto patterns|do not bypass root lint/build entrypoints for cross-service changes|do not place service-specific logic in pkg when it belongs in internal/<service>
|Unique Styles:cmd/application is multi-command CLI|cmd/backend|file|relay are direct server starters|tools is a separate Go module|root Makefile orchestrates binaries/frontend/tests/tools/script packaging/docker images
|Commands:make pre|make all|make lint|make clean
|Commands:proto regen=cd proto && make clean && make all
|Commands:frontend local=cd front && pnpm install && pnpm dev
|Commands:integration tests=cd test && make build && make test
