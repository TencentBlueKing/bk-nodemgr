# Adapter Boundaries

## Read The Owning Contract

| Reference                                                      | What to inspect                                                           |
| -------------------------------------------------------------- | ------------------------------------------------------------------------- |
| `pkg/thirdparty/AGENTS.md`                                     | Handler interface dependency and protocol isolation                       |
| `pkg/thirdparty/backend/handler.go`                            | Aggregate `IHandler`, private client, concrete constructor                |
| `pkg/thirdparty/backend/host.go`                               | Domain capability signatures and conversion                               |
| `pkg/thirdparty/cmdb/handler.go`                               | Domain interface composition; do not copy unrelated legacy error handling |
| `support-files/bkiamv4/migrate/iamv4/AGENTS.md`                | CLI-local adapter ownership and runner validation                         |
| `support-files/bkiamv4/migrate/iamv4/handler_resource_type.go` | Capability-level pagination and completeness                              |
| `support-files/bkiamv4/migrate/iamv4/resource_type.go`         | Single-page/list, array-body create, empty-body update                    |
| `support-files/bkiamv4/migrate/iamv4/types.go`                 | Named DTOs and omitted/empty semantics                                    |
| `support-files/bkiamv4/migrate/resource_type.go`               | Upsert and allowed-update policy outside the adapter                      |
| `pkg/rest/client/request.go`                                   | `Into`, `RawData`, body lifetime, error content, body logging             |

Read current symbols instead of assuming these files are unchanged. Responsibility is the reference, not a fixed file count.

## ResourceType Example

The migration runner decides whether to create, rename, skip, or reject ancestor drift. Handler exposes ResourceType capabilities and obtains a complete collection when the runner needs a snapshot. The raw client assembles a single request and checks its protocol outcome. The transport already provides context, auth/header handling, and response consumption.

This separation allows changing the runner's allowed-update policy without changing endpoint assembly. It does not justify adding unsupported protocol operations, a generic migration framework, or registration methods to the runtime authorization adapter.

### Legacy Signatures Are Not The New Default

Some raw methods return `*BaseBroker[...]`; some 204 methods return only `error`. These are existing implementations, not templates for the approved new signature convention. A new API operation uses `*OperationReq -> (*OperationResp, error)`, including a named empty response for 204. Keep an existing envelope helper inside decoding where useful. Handler remains free to return only the capability result.

Do not fix all such legacy methods while extending one operation. If the caller or public contract makes changing an existing signature necessary, establish that scope first.

### Required Interfaces Are Not Feature Detection

An aggregate `IHandler` embeds domain `IHandlerXxx` interfaces. A role runner that requires role operations accepts the appropriate capability statically; it should not discover a missing required method set after execution has begun. This does not require adding interfaces to every internal helper or forcing every consumer to accept the entire aggregate.

### Reuse Must Be Verified

Inspect `Result.Into` / `RawData` before proposing a custom decoder. Also inspect error returns and logging, not only successful JSON decoding: raw response data may enter diagnostics even when headers are masked. Conversely, the existence of a hypothetical sensitive response is not evidence that every client needs a new decoder. Where an actual safety requirement cannot be met, disclose the mismatch and obtain approval for a bounded change. Never conceal it with a blanket "reuse is safe" or "disable all errors" rule.

## Do Not Generalize IAM Policy

- Migration runner input validation does not override runtime providers that require outbound validation.
- Complete migration snapshots do not replace explicit pagination in business-facing APIs.
- IAM-specific `404 + NOT_FOUND`, batch limits, ancestor rules, role additions, dry-run state, and retry policy remain provider/caller contracts.
- Runtime noop, caching, token handling, and model registration have separate scope rules.
- Existing logging calls and error strings are implementation facts, not authority over current security or error-handling rules.

## Evidence Labels

Distinguish scoped requirement, observed implementation, approved new convention, and unresolved conflict. When they disagree, state which applies and why. Do not resolve a conflict by counting examples or choosing the shortest implementation.
