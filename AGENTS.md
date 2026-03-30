|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/comments/docs/tech discussions
|Compression Rule:Follow pipe-index format; keep concise; no prose/code blocks
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
|OCP:extend by addition, not mutation|anchor:{internal/*/router/api-v3,internal/backend/auth,pkg/proto/*}|pattern:{new-subpackage,new-constant,new-interface-impl,additive-proto,new-method}|ban:{patching-stable-signatures,rename/remove-proto-fields}
|SRP:one unit=one reason to change|anchor:{internal/<service>,internal/*/router/api-v3,internal/backend/auth,pkg}|pattern:{domain-split,service-private-internal,shared-only-in-pkg}|signal:{multiple-change-reasons,file-sprawl}
|ISP:depend on minimal interfaces only|anchor:{internal/backend/auth,pkg/proto/*}|pattern:{small-interfaces,domain-split-services}|ban:{catch-all-interfaces,methods-unused-by-implementors}
|DIP:high-level code depends on abstractions|anchor:{cmd/*,internal/*/storage,internal/backend/auth,pkg/types}|pattern:{wire-in-cmd,interface-at-boundary,impl-behind-interface}|ban:{handler-to-concrete-impl,direct-dependence-on-pb-structs}
|LoD:only talk to direct friends|anchor:{handler→service→storage,pkg/proto/*}|pattern:{layered-calls,converter-boundary}|ban:{cross-service-internal-imports,chain-calls}
|LSP:implementations must be transparently substitutable|anchor:{IAuthorizer,storage interfaces,pkg/proto converters}|require:{stable-error-semantics,stable-return-meaning,backward-compatible-conversion}|ban:{silent-noop-required-methods}
|ETC/DRY:prefer changes that keep future edits local and cheap|anchor:{pkg helpers,pkg/proto conversions,router patterns}|rule:{merge-true-duplicates,avoid-wrong-abstractions,prefer-project-standard-patterns}
|Go error/logging:handle-or-explicitly-ignore errors|pattern:{fmt.Errorf+%w,error-wrap}|ban:{bare-recover,happy-path-logging}|observe:{pkg/logger,levelled-logs}
|Go engineering:composition over hierarchy|rule:{no-package-globals-without-need,channels-over-shared-state,design-for-testability,control-entropy}|ban:{hardcode,special-case-patterns,type-hierarchies}
|Crash Early:main trunk=happy path|branches=fail fast|pattern:{guard-clause,early-return,error-wrap,boundary-validation}|ban:{masked-error,silent-recover,partial-state-continue}|observe:{pkg/logger.Error,metrics,alerting}
|Self-Doc:self-explaining-code>nearby-comment>doc.go>README|exported symbols=require godoc comments|magic values=require inline reason|if comment explains naming/flow, rename or extract first|package-level context=doc.go
|DBC:contract={preconditions,postconditions,invariants}|handler=input validation boundary|pkg/types=domain contract|pkg/proto/*=conversion contract|storage/service interfaces=require stable error semantics|split when contract cannot be stated simply
|SLA:one function=one abstraction level|handler flow:{validate→service→response}|ban:{inline DB logic in handler,inline proto field shuffling in handler,mixed orchestration+detail}|detail blocks→named sub-functions|80+ lines=refactor signal
|Terminology:naming must align across {proto,pkg/types,API JSON,front/src,docs}|grep existing term before adding new one|same concept=same name|ban:synonyms for existing domain terms
|Business Model:code to domain model, not transient UI/requirement wording|pkg/types=domain anchor|handler/router=transport mapping only|new requirement=extend model or mapping, not ad-hoc branch explosion|special case:belongs to model? yes→model it; no→presentation layer
|Explicit Handling:no ambiguous intent in code|ignore error=_ = F() not bare call|empty branch=require "intentionally empty" comment|unfinished work=require TODO/FIXME + context|zero-value semantic dependence=require explicit comment
|Encapsulation:extract by {layer,feature,domain,component}|function extraction must reduce complexity, not just shorten|package boundaries must match responsibility boundaries|ban:{pkg/common grab-bags,pkg/utils grab-bags,pkg/helper grab-bags}|80+ lines=extract single-responsibility sub-blocks
|Anti-patterns:never hand-edit generated *.pb.go|treat proto structs as boundary types and convert via pkg/proto/* before business logic|do not introduce duplicate helpers or parallel conversion logic before checking current service/router/pkg/proto patterns|do not bypass root lint/build entrypoints for cross-service changes|do not place service-specific logic in pkg when it belongs in internal/<service>|do not modify stable handler signatures or interfaces when extending—add new code instead (OCP)
|Unique Styles:cmd/application is multi-command CLI|cmd/backend|file|relay are direct server starters|tools is a separate Go module|root Makefile orchestrates binaries/frontend/tests/tools/script packaging/docker images
|Commands:make pre|make all|make lint|make clean
|Commands:proto regen=cd proto && make clean && make all
|Commands:frontend local=cd front && pnpm install && pnpm dev
|Commands:integration tests=cd test && make build && make test
