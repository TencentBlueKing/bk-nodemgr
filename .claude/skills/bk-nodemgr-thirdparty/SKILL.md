---
name: bk-nodemgr-thirdparty
description: Use when adding, extending, refactoring, or reviewing bk-nodemgr third-party adapters, including pkg/thirdparty and CLI-local adapters with handler/client separation, named API request-response contracts, domain interfaces, pagination, or shared APIGW/REST client usage. Not for operating bk-cli or designing migration policy alone.
---

# bk-nodemgr Thirdparty

## Purpose

Handler provides capabilities; domain API methods isolate the external protocol. Keep callers predictable without building another HTTP framework.

Applies to shared providers and explicitly scoped CLI-local adapters. Similar layering does not authorize moving tool-specific registration into `pkg`.

## Start From Evidence

1. Read applicable `AGENTS.md`, provider documentation, existing callers, and 2-3 comparable methods. Compare responsibilities and contracts, not syntax alone.
2. Retrieve the external API's actual method, path, auth, input, output, success statuses, and pagination. Use `bk-cli-apigateway` for BlueKing API discovery and Context7 for applicable third-party documentation. Missing facts are research tasks, not questions for the user to guess.
3. Confirm the semantic API name, purpose, request/response, and Handler capability together with the user before introducing a new API operation. Reuse an already confirmed name; do not repeatedly ask about established operations. HTTP POST alone does not establish an operation's meaning.
4. Trace `caller -> handler interface -> conversion/aggregation -> raw API -> shared client`. Record validation and policy ownership before editing.

Scoped requirements take precedence over examples. The signature convention below is an explicitly approved default for new raw API methods, not a claim that all legacy code already follows it. If a scoped rule conflicts, surface the conflict before the affected edit. Do not normalize unrelated legacy code.

## Responsibility Map

| Surface         | Owns                                                                               | Does not own                                |
| --------------- | ---------------------------------------------------------------------------------- | ------------------------------------------- |
| Caller / runner | Business sequencing, migration policy, dry-run, allowed updates                    | Raw endpoint assembly or wire DTOs          |
| Handler         | Scenario capabilities, DTO conversion, project-facing errors, required aggregation | Migration policy or duplicate HTTP plumbing |
| Raw API method  | One API request, path/query/body, protocol status and envelope handling            | Cross-page iteration or business workflow   |
| Types           | Named protocol requests/responses and explicit field semantics                     | Caller policy                               |
| Shared client   | Existing transport, authentication, context, decoding and logging mechanisms       | Provider business decisions                 |

Keep small packages in `handler.go`, `<provider>.go`, and `types.go`. When domain separation is needed, follow local `handler_<domain>.go` and `<domain>.go` examples; keep common client initialization separate. Do not split files merely to satisfy a diagram.

## API Method Contract

New raw API methods use one predictable shape:

```go
func (c *cli) operateProcV2(
    nCtx contextx.IContext, req *operateProcV2Req,
) (*operateProcV2Resp, error)
```

`operateProcV2` illustrates a confirmed semantic name, not a name to copy for another endpoint. Method, Req, and Resp share that semantic stem; exportedness follows the package contract.

- Exactly context plus one named pointer request, returning a named pointer response and error. Put path IDs, query, and body inputs in that request; serialization tags keep path-only fields out of the body.
- Use a named response even for HTTP 204, with an empty response type if appropriate. Successful empty bodies are not JSON decode failures. Handler may still return only `error` when that is its capability contract.
- A generic `BaseBroker[T]` may help decoding internally; it does not replace the named response in a new raw method signature. Do not create another wrapper if an existing named type already satisfies the contract.
- Maps, `any`, anonymous structs, and extra scalar parameters are not substitutes for a known API contract. Genuinely open payload fields may still use protocol-appropriate types inside the named request.
- This rule does not apply to constructors, conversion helpers, or Handler signatures. Preserve stable existing methods unless their migration is explicitly in scope.

## Handler Contracts

Use handler-layer interfaces as required by the project. Compose domain interfaces such as `IHandlerSystem` and `IHandlerRole` into `IHandler`; consumers depend on the smallest required capability. Required capabilities are compile-time dependencies, not discovered with `handler.(RoleHandler)` and an "unavailable" error at runtime. Truly optional capabilities need an explicit optional contract.

Handler signatures express caller needs: scalars, domain models, slices, or error-only results can be appropriate. Do not force wire Req/Resp onto callers. Reuse existing domain models and converters; a CLI-local model does not automatically belong in shared `pkg/types`. Constructor concrete returns do not contradict interface-based consumption.

## Protocol And Reuse Checks

| Question                     | Required decision                                                                                                                                                                                                                                     |
| ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Who validates input?         | Follow scoped ownership. Separate startup config, caller input, and remote-response checks; do not add `Validate()` to every layer.                                                                                                                   |
| Page or complete collection? | Raw method retrieves one page. Handler aggregates only when its capability promises completeness; never silently return partial state as complete.                                                                                                    |
| Omitted, empty, or null?     | Model distinctions required by the protocol; use pointers when omission must survive. Do not universally prescribe pointer fields.                                                                                                                    |
| Success or absence?          | Check documented HTTP statuses and provider envelope. A failed query is not evidence of absence. Do not copy IAM-specific codes to other providers.                                                                                                   |
| Existing client sufficient?  | Inspect `pkg/rest/client` and APIGW helpers before adding decoding, auth, transport, retries, or pagination helpers. Prefer the existing path if it meets the contract.                                                                               |
| Safety gap?                  | Trace request/response logging, underlying errors, and final diagnostics. Header masking does not sanitize bodies. Explain a concrete gap and seek approval for the smallest necessary deviation; do not silently invent a parallel `decodeResponse`. |

Do not blindly enable body logging or hide useful errors by default. Apply `bk-nodemgr-error-handling` and `bk-nodemgr-logger` to the actual data and handling boundary. Preserve tenant and cancellation with `bk-nodemgr-contextx`.

## Delivery Check

Report briefly: references, confirmed API/capability signatures, responsibility boundaries, deviations or unresolved facts, and verification actually performed. Reviews remain findings-first; no extra verdict framework is required here.

Run scoped build/tests and the applicable project lint entrypoint. Compilation alone is not lint verification. Validate real protocol cases justified by the task, respecting project test-delivery rules. External registration or authorization writes require explicit permission; API discovery is not permission to mutate a system.

## References And Companions

- Read [Adapter Boundaries](references/adapter-boundaries.md) when comparing CLI-local and shared providers, or when a legacy example conflicts with a new signature.
- `bk-nodemgr-how-to`: routing; `bk-nodemgr-code-style`: Go readability.
- `bk-nodemgr-error-handling`, `bk-nodemgr-logger`, `bk-nodemgr-contextx`: their respective contracts, not duplicated here.
- `bk-nodemgr-naming-boundary`: unresolved cross-layer terminology; `bk-nodemgr-architecture-judgment`: placement or architectural deviations, subordinate to applicable scoped rules.
- `code-review`: review workflow; `bk-cli-apigateway`: external API discovery, not Go implementation ownership.

## Evaluation

`evals/evals.json` covers protocol naming, typed empty responses, required interfaces, validation ownership, pagination, and reuse under time pressure. Keep evaluation outputs outside the repository. Existing IAM behavior is evidence, not a general-purpose migration specification.
