---
name: bk-nodemgr-how-to
description: Use when planning, implementing, reviewing, or debugging Go changes in bk-nodemgr and deciding which project skills and generic Go companion skills to load.
---

# bk-nodemgr how-to

## Overview

`bk-nodemgr-how-to` is the project-first Go skill router for bk-nodemgr. It adapts the generic `golang-how-to` idea to this repository: route first to bk-nodemgr contracts, scoped `AGENTS.md` files, and project skills; use generic Go skills only as companion background when the task needs language-level guidance.

Use this skill to avoid treating project-owned surfaces such as `pkg/contextx`, `pkg/runtime/conv`, `pkg/runtime/gopool`, `pkg/runtime/retrier`, `pkg/logger`, API scaffolding, and router permission work as generic Go tasks.

## When to Use

Use this skill when a Go task in bk-nodemgr asks:

- Which skills should be loaded for implementation, review, debugging, or design.
- Whether a task is project-specific or generic Go.
- How to combine repo contract skills with generic Go skills.
- How to route work touching `internal/*`, `pkg/*`, `cmd/*`, `proto/**`, router groups, storage, workflow, runtime helpers, context propagation, or logging.
- How to replace generic skill routing such as `golang-how-to` with bk-nodemgr-specific routing.

## When Not to Use

Do not use this skill as the only guide when a narrower project skill clearly applies. Load the narrower skill too.

Do not use it for:

- Non-Go documentation-only tasks with no bk-nodemgr Go implementation or review surface.
- Pure external Go questions outside this repository; use `golang-how-to` and the relevant generic Go skills.
- Git operations; use `git-master` or the project PR/release skills.
- Frontend UI work; use the frontend/visual route instead of Go routing.

## Source Anchors

Read these before inventing routing rules:

- `AGENTS.md`: repository-wide architecture, retrieval-first rules, and scoped `AGENTS.md` priority.
- `.claude/skills/*/SKILL.md`: project skill catalog and local skill style.
- `~/.agents/skills/golang-how-to/SKILL.md`: generic Go routing model to adapt, not copy blindly.
- `pkg/contextx/AGENTS.md`: project context propagation contract.
- `pkg/runtime/conv/AGENTS.md`, `pkg/runtime/criteria/AGENTS.md`, `pkg/runtime/crypter/AGENTS.md`: runtime subpackage contracts.
- `.claude/skills/bk-nodemgr-{conv,gopool,retrier,contextx,logger,error-handling}/SKILL.md`: high-frequency project Go skills.
- `.claude/skills/api-scaffold/SKILL.md`, `.claude/skills/router-permission-supplement/SKILL.md`, `.claude/skills/api-doc/SKILL.md`, `.claude/skills/code-review/SKILL.md`: project workflow and review skills.

## Routing Rules

Project-specific skills take precedence over generic Go skills when the task touches a bk-nodemgr-owned contract. Generic Go skills remain useful as companions for language-level details.

| Task Signal | Load First | Add Generic Companions When Needed |
| --- | --- | --- |
| `pkg/contextx`, `contextx.IContext`, tenant/user/message propagation, `WithoutCancel` | `bk-nodemgr-contextx` | `golang-context`, `golang-concurrency` |
| `pkg/runtime/conv`, data shaping, duplicate conversion helpers | `bk-nodemgr-conv` | `golang-data-structures`, `golang-safety`, `golang-testing` |
| `pkg/runtime/gopool`, fan-out/fan-in, `Wait`, bounded goroutines | `bk-nodemgr-gopool`, `bk-nodemgr-error-handling` | `golang-concurrency`, `golang-safety` |
| `pkg/runtime/retrier`, polling, backoff, fallback candidates, `err == nil` success branches | `bk-nodemgr-retrier`, `bk-nodemgr-error-handling` | `golang-context`, `golang-observability` |
| Go error creation, propagation, wrapping, inspection, aggregation, `err == nil`, panic/recover, `resterrf.ErrWrap`, log-or-return responsibility | `bk-nodemgr-error-handling` | `golang-safety`, `golang-error-handling` only for language details not covered by the project skill |
| `pkg/logger`, Biz/Sys logs, fields, levels, third-party logger adapters | `bk-nodemgr-logger` | `bk-nodemgr-error-handling` for error responsibility, `golang-observability` for general concepts |
| Third-party adapters, handler/client isolation, named API Req/Resp, domain interfaces, CLI-local APIGW adapters | `bk-nodemgr-thirdparty` | Project error/logger/contextx skills as needed; generic Go skills only for language-level gaps |
| `testsuite/support`, package-level Mongo/Redis integration tests, `NODEMGR_TEST_*`, `RequireMongoDatabase`, `RequireRedisClientWithKeyPrefix` | `bk-nodemgr-testsuite-support` | `golang-testing`, `golang-database` |
| Cross-layer placement, competing designs, dependency direction, shared-contract changes, pre-flight/post-flight judgment, explicit deepening scans | `bk-nodemgr-architecture-judgment` | Narrow project skill for the affected surface, then generic Go skills only for language-level gaps |
| New REST/proto endpoint scaffolding | `api-scaffold`, `bk-nodemgr-error-handling` | `golang-grpc`, `golang-testing` when implementation requires them |
| Router permission mapping | `router-permission-supplement` | `golang-security` only for broader security review |
| Proto API reference documentation | `api-doc` | `golang-documentation` only for generic doc style |
| Code review, PR review, lint-risk audit | `code-review` | `golang-lint`, `golang-safety`, `golang-testing`, `golang-security` as findings require |
| README or logic-first docs | `readme-logic-first` | `golang-documentation` only for Go doc comments/examples |
| Release/changelog/version alignment | `release-version-alignment`, `changelog-doc` | Generic Go skills usually unnecessary |
| Scoped `AGENTS.md` creation/compression | `scoped-agentsmd` | Generic Go skills only if source code patterns must be interpreted |

If no project skill owns the surface, fall back to `golang-how-to` and load the relevant generic Go skills directly.

## Generic Go Companion Matrix

Use generic Go skills for language concerns after the project surface is identified.

| Language Concern | Generic Skill |
| --- | --- |
| Naming packages, exported identifiers, test names | `golang-naming` |
| Error creation, wrapping, `errors.Is/As`, panic/recover | `bk-nodemgr-error-handling` first; use `golang-error-handling`, `golang-safety` only for language-level gaps |
| Goroutines, channels, locks, shared state, cancellation observation | `golang-concurrency`, `golang-context` |
| Context cancellation/deadlines without project identity values | `golang-context` |
| Tests, table-driven cases, testify, flaky tests | `golang-testing`, `golang-stretchr-testify` |
| Lint, staticcheck, vet, style issues | `golang-lint`, `golang-code-style` |
| Security-sensitive input, crypto, secrets, filesystem/network risk | `golang-security`, `golang-safety` |
| Performance optimization after measurement | `golang-benchmark`, then `golang-performance` |
| Database/sqlx/transactions/query safety | `golang-database`, `golang-security`; use `bk-nodemgr-error-handling` for project error propagation |
| CLI/cobra/viper command behavior | `golang-cli`, `golang-spf13-cobra`, `golang-spf13-viper` |

## Core Pattern

### Route in four steps

1. Identify the repo surface: path, package, router group, service, or workflow.
2. Read nearest `AGENTS.md` and relevant project skill first.
3. Add generic Go skills only for language-level gaps the project skill does not own.
4. Verify through the project surface: source anchors, diagnostics, adjacent tests, build, or manual/API checks as applicable.

### Prefer narrow project skills over broad routing

If work touches `pkg/runtime/gopool`, load `bk-nodemgr-gopool` first, not only `golang-concurrency`. If work touches `pkg/contextx`, load `bk-nodemgr-contextx` first, not only `golang-context`. This keeps guidance aligned with bk-nodemgr contracts and still allows generic Go skills as companions.

For architecture judgment signals, load `bk-nodemgr-architecture-judgment` as a companion guardrail; it supplements scoped `AGENTS.md` and narrow project skills, and should not replace the skill that owns the affected surface.

## Common Mistakes

- Loading only `golang-how-to` for a task that has a project skill with stricter source anchors.
- Treating this skill as permission to skip scoped `AGENTS.md`; scoped instructions still override broad routing.
- Loading every Go skill. Choose the smallest set that covers the project surface and the language concern.
- Replacing project contracts with generic Go advice, especially for `contextx`, runtime helpers, logger, router permissions, or API scaffolding.
- Forgetting non-Go project skills for Go work that includes docs, changelog, PR, release, or permission matrices.
- Adding a new project skill before checking whether an existing one already owns the domain.

## Verification Checklist

- Search existing skills before routing: `rg "name:|description:|Cross-References|When to Use" .claude/skills`.
- Search source anchors before deciding a project surface is generic: `rg "pkg/contextx|pkg/runtime|pkg/logger|api-v3|workflow" internal pkg cmd proto`.
- If replacing a generic skill reference, verify the new reference is project-specific and not a pure Go fallback.
- Keep `golang-how-to` as fallback for pure Go tasks outside bk-nodemgr contracts.
- When a new project skill is added, update this router if it changes routing decisions.

## Evals

Pressure prompts live in `evals/evals.json`. Keep run outputs, timing, grading, benchmarks, and review artifacts outside git unless explicitly requested.

## Cross-References

- `golang-how-to`: generic Go skill orchestrator and fallback outside bk-nodemgr project contracts.
- `bk-nodemgr-contextx`: project context propagation and value-preserving cancellation/deadline handling.
- `bk-nodemgr-conv`: project conversion helper contracts.
- `bk-nodemgr-gopool`: project grouped goroutine execution contracts.
- `bk-nodemgr-retrier`: project retry primitive selection.
- `bk-nodemgr-error-handling`: project Go error semantics, wrapping, REST mapping, log-or-return responsibility, and `err == nil` migration.
- `bk-nodemgr-logger`: project logging conventions.
- `bk-nodemgr-thirdparty`: provider capabilities, raw API contracts, and shared client reuse, including CLI-local adapters.
- `bk-nodemgr-testsuite-support`: project package-level Mongo/Redis integration test support.
- `bk-nodemgr-architecture-judgment`: pre-flight/post-flight architecture judgment and explicit deepening scans.
- `api-scaffold`: project API endpoint scaffolding.
- `router-permission-supplement`: router permission action/resource mapping.
- `code-review`: project Go change review.
