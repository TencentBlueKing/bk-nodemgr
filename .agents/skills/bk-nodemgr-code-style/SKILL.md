---
name: bk-nodemgr-code-style
description: Use when writing, reviewing, or refactoring Go code in bk-nodemgr for project-specific readability, control flow, layering, comments, initialization, error/log style, or consistency with mature Node Manager code.
---

# bk-nodemgr code style

## Overview

`bk-nodemgr-code-style` judges Go readability through BlueKing Node Manager's own conventions. First read the owning layer and mature local examples, then decide whether the code matches Node Manager's router, service, storage, proto conversion, runtime helper, workflow, and tooling conventions.

This skill is not a Go tutorial. It keeps style decisions tied to project evidence, scoped `AGENTS.md`, `.golangci.yml`, and mature local Node Manager code.

## When to Use

Use this skill when a Go task in bk-nodemgr asks about:

- Code readability, function shape, line breaking, early returns, or nested control flow.
- Whether a handler/service/storage/helper implementation feels consistent with existing Node Manager code.
- Review findings about style drift, excessive abstraction, mixed abstraction levels, or unclear boundary ownership.
- Naming, comments, data initialization, error wrapping, or log message shape when the answer depends on project style.
- Turning broad readability concerns into bk-nodemgr-specific guidance.

## When Not to Use

Do not use this skill as the only guide when a narrower project skill owns the surface:

- Error handling semantics, wrapping, REST mapping, log-or-return responsibility, or `err == nil` migration: use `bk-nodemgr-error-handling` too.
- Logging API choices: use `bk-nodemgr-logger` too.
- `pkg/runtime/conv` helpers or conversion contracts: use `bk-nodemgr-conv` too.
- `pkg/contextx` and tenant/user/message propagation: use `bk-nodemgr-contextx` too.
- `pkg/runtime/gopool` or concurrency helpers: use `bk-nodemgr-gopool` too.
- `pkg/runtime/retrier` polling/backoff behavior: use `bk-nodemgr-retrier` too.
- New REST/proto endpoint scaffolding: use `api-scaffold` too.
- Router permission action/resource mapping: use `router-permission-supplement` too.

For pure Go language questions outside this repository, use general Go guidance instead of this project-specific skill.

## Source Anchors

Read these before inventing style rules:

- `AGENTS.md`: repository-wide layering, simplicity, surgical changes, scoped instruction priority, and no speculative abstractions.
- `.golangci.yml`: project lint thresholds such as `funlen.lines: 120`, `gocyclo.min-complexity: 20`, `gocognit.min-complexity: 20`, `godot`, `lll`, `revive`, and `cyclop`.
- `.agents/skills/bk-nodemgr-{how-to,error-handling,logger,conv,contextx,gopool,retrier}/SKILL.md`: existing project Go skill style and cross-reference model.
- `.agents/skills/code-review/references/go-standards.md`: current review quick reference for conversions, errors, logging, comments, API, and proto.

Representative code anchors:

- `pkg/runtime/conv/conv.go`: mature runtime helper style with explicit nil/overflow handling, deterministic map/slice behavior, and narrow helpers.
- `pkg/runtime/retrier/polling.go`: small constructor/options/poll loop style in an existing runtime helper package.
- `pkg/dao/mongo/base/param_builder.go`: compact builder helpers with explicit `bson.D` shape and no extra abstraction.
- `internal/backend/router/api-v3/node/agent/install.go`: handler flow with bind, conversion, permission checks, service call, response, and Biz logs.
- `internal/backend/router/api-v3/topo/networkunit.go`: transport boundary style and domain object construction before storage call.
- `internal/backend/storage/topo/topo.go`: storage/domain helpers using guard clauses and small private functions.
- `tools/internal/installer/**`: installer step style when working in the separate tools module; keep it secondary because `tools/` has its own module boundary.

## Quick Reference

| Style Question | Prefer in bk-nodemgr | Check |
| --- | --- | --- |
| Handler flow | `BindJSON` / validate -> auth -> convert -> service/storage -> response | Keep business derivation out of transport when possible |
| Service/storage flow | Small domain helpers and guard clauses | Avoid inline DB/proto details in orchestration code |
| Proto conversion | `ConvertXToTypes` / `ConvertXFromTypes`, explicit nil/default handling | Do not leak proto structs into business logic |
| Error handling | Use `bk-nodemgr-error-handling` for wrapping, REST mapping, log-or-return responsibility, and `err == nil` migration | Preserve happy-path linearity with `if err != nil` guards |
| Logging | `logger.G.Biz(rCtx)` for request/business flow, structured fields, `WithErr` for errors | Load `bk-nodemgr-logger` for logging-specific choices |
| Initialization | Explicit `make` / `new` / struct literals for values that cross boundaries | Avoid nil slice/map/object semantics leaking to callers |
| Control flow | Early return, `continue` for skip cases, small private helpers | Avoid 3+ nesting levels and rightward drift; use `golang-non-arrow-control-flow` for deep nested-flow judgment |
| Constants and terms | Existing domain names from `pkg/types`, proto, routers, and frontend | Grep before adding synonyms |
| Shared helpers | Extend existing package-owned helpers only when behavior is portable | Do not create `pkg/common`, `utils`, or caller-local duplicates |

## Core Patterns

### Start from the owning layer

Find the layer first, then apply style:

1. Router: validate request boundary, auth, call service/storage, convert response.
2. Service/manager: orchestrate domain workflow and dependency calls.
3. Storage: own DAO interaction and storage-domain helpers.
4. `pkg/types`: carry business model contracts between layers.
5. `pkg/proto/**`: confine proto lifecycle and conversion.
6. `pkg/runtime/**`: contain business-neutral helpers only.

Do not flatten these roles into one function because the code happens to compile. In Node Manager, style includes boundary ownership.

### Keep happy path linear

Mature code favors a readable top-level sequence with guard clauses:

```go
func (h *handler) AgentInstall(rCtx restserver.IContext) (interface{}, error) {
    req := new(protoBackend.NodeAgentInstallReq)
    if err := rCtx.BindJSON(req); err != nil {
        return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
    }

    nodeDeployments, bizIDs, err := h.generateInstallNodeDeployments(rCtx, req)
    if err != nil {
        return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
    }

    workflowID, err := h.nodeMgrIface.LaunchInstallNode(rCtx, types.InstallNodeParam{...})
    if err != nil {
        return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
    }

    resp := new(protoBackend.NodeAgentInstallResp)
    resp.ConvertWorkflowID(workflowID)
    return resp.GetData(), nil
}
```

If the handler starts mixing request parsing, permission resource construction, storage queries, domain validation, and response conversion in one long block, extract named helpers by responsibility.

Use `golang-non-arrow-control-flow` when the question is specifically about arrow code, terminal `else` branches, nested skip/error handling, or whether guard clauses are being over-applied. In bk-nodemgr, its generic advice is subordinate to the owning layer: do not flatten a function by moving business derivation into a router, leaking proto structs into service/storage code, duplicating a converter, or widening shared helpers. Prefer early returns and `continue` only when they preserve side-effect order, resource lifetime, error mapping, and the router -> service -> storage/proto boundary.

### Use helper extraction to name domain decisions

Good extraction in bk-nodemgr reduces cognitive load and reveals domain intent. It is not extraction for line-count cosmetics.

Examples of useful helper shapes:

- `fetchNetworkunits`: collect IDs, return empty map when no IDs, query storage, map by ID.
- `prepareNetworkUnitUpdate`: short domain precondition before storage mutation.
- `splitAccessPoints`: split old/new access points and use `continue` to keep skip logic flat.
- `generateInstallNodeDeployments`: keep request-to-domain assembly outside the public handler flow.

Avoid helpers named after mechanics such as `processData`, `handleLogic`, or `doCheck`; the name should expose the domain decision.

### Convert at boundaries, not in the middle

Router and proto packages should make data shape explicit:

- Use proto supplement methods such as `Convert...ToTypes` and `Convert...FromTypes`.
- Use `pkg/types` as the business model anchor.
- Use `pkg/runtime/conv` only for portable shape conversion, not business rules.
- Preserve nil/default semantics intentionally, especially when proto `GetX()` could collapse absence into zero value.

If a style question involves adding a converter, first search `pkg/proto/**` and current call sites before writing a local mapper.

### Initialize values that cross boundaries

When returning collections, response payloads, option objects, or domain maps, prefer explicit usable values over accidental nil semantics:

- `make(map[int64]*types.NetworkUnit)` for empty lookup results.
- `make([]int64, 0, len(items))` when collecting known-size IDs.
- `new(protoBackend.SomeResp)` before response conversion.
- Field-named struct literals for request/domain payloads.

Nil can be correct only when it has a stable contract in that package. If callers would need to guess, make the value explicit.

### Treat lint thresholds as late warnings, not style targets

`.golangci.yml` allows up to 120 function lines and complexity 20, but those are guardrails. In normal feature work, simplify earlier:

- 80+ lines means inspect abstraction levels.
- 3+ nesting levels means use guard clauses, `continue`, or extract a domain helper; if nesting represents state policy, model it with `switch`, a table, or a domain-owned helper instead of scattered guards.
- Multiple reasons to change means split by layer or responsibility.
- Repeated conversion or filtering loops means search existing helpers before duplicating.

## Common Mistakes

- Applying broad Go readability advice without checking the owning Node Manager layer.
- Adding a helper in the caller when `pkg/runtime/conv`, proto supplement methods, or storage/domain helpers already express the shape.
- Treating `internal/backend/router/api-v3/**` as a place for business derivation instead of transport orchestration.
- Passing proto structs through service/storage layers instead of converting to `pkg/types`.
- Using nil slices/maps as accidental output contracts for API/proto/storage results.
- Hiding domain branches inside anonymous inline blocks instead of naming them with private helpers.
- Adding comments to explain confusing flow instead of renaming or extracting the flow.
- Letting platform, service, or tenant special cases become broad `if` ladders; split by package, helper, or domain owner.
- Introducing shared `pkg/common`, `pkg/utils`, or broad interfaces before proving the behavior is business-neutral and reused.

## Verification Checklist

- Search existing style anchors before changing style: `rg "Convert.*ToTypes|Convert.*FromTypes|resterrf.ErrWrap|logger.G.Biz|conv\." internal pkg`.
- Check nearest scoped `AGENTS.md` for touched paths before deciding a style rule.
- Compare against 2-3 analogous files in the same service/layer before making a style judgment.
- For changed Go files, run `gofmt`/project formatting through the repo's normal tool path and check diagnostics.
- Run the narrowest adjacent test or build surface that exercises the changed layer.
- If a style decision crosses into error handling, logging, conversion, context, gopool, or retrier behavior, load the narrower bk-nodemgr skill and follow it first.

## Eval Prompts

Use prompts like these to test whether the skill changes behavior:

- "Review this new backend API handler for Go style; does it match bk-nodemgr router/service/storage conventions?"
- "Refactor this long storage function without changing behavior; prioritize mature Node Manager style."
- "I added a small conversion helper inside a workflow action. Should it stay there or move to existing project conversion surfaces?"
- "This function passes lint but feels hard to read. Judge it against mature bk-nodemgr code style."

Good with-skill answers should cite project anchors, layer ownership, and concrete patterns; weak answers stay at broad readability advice.

## Cross-References

- `bk-nodemgr-how-to`: project-first Go skill routing.
- `code-review`: finding-oriented Go review and project review report style.
- `bk-nodemgr-error-handling`: project Go error semantics, wrapping, REST mapping, log-or-return responsibility, and `err == nil` migration.
- `bk-nodemgr-logger`: logger API, Biz/Sys log choice, fields, levels, and message shape.
- `bk-nodemgr-conv`: `pkg/runtime/conv` and conversion helper contracts.
- `bk-nodemgr-contextx`: project context propagation and value-preserving cancellation/deadline behavior.
- `bk-nodemgr-gopool`: project goroutine pool and fan-out/fan-in contracts.
- `bk-nodemgr-retrier`: polling/backoff/retry primitive selection.
- `golang-non-arrow-control-flow`: detailed Go guidance for arrow code, terminal `else`, guard clauses, `continue` guards, and safe exceptions.
- `api-scaffold`: new REST/proto endpoint scaffolding.
- `router-permission-supplement`: handler permission action/resource mapping.
