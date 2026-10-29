---
name: bk-nodemgr-proto
description: Use when adding, changing, or reviewing bk-nodemgr proto/** API contracts, RPC/message/field comments, Service/RPC naming or tags, or investigating unexpected Chinese descriptions, summary/description, operationId, or labels in generated Swagger and APIGW resources.
---

# bk-nodemgr Proto

## Role and boundaries

Treat Proto documentation as part of the API caller's contract, not lint filler. Produce useful Chinese descriptions at the source so developers do not need to translate generated Swagger by hand.

Apply this guidance to new or changed interfaces and their relevant comments. Do not bulk-rewrite historical interfaces, translate tags, rename identifiers for presentation, or introduce OpenAPI options as a default documentation pattern. Ordinary Proto comments are the first choice; discuss a specific generation limitation before adding overrides or changing the generation pipeline.

The project flow is **Proto → generated Swagger → manual import into APIGW → resource export → `apigw/resources.yaml`**. It is not a local automatic Swagger-to-resources build step, nor proof of a lossless round trip. Do not execute gateway import, synchronization, deletion, or publication.

## Start from the current contract

1. Read applicable `AGENTS.md`, `proto/README.md`, the target Proto, and two or three comparable interfaces in the same service when available. Report a missing analogue rather than inventing one. Read `proto/Makefile` and `.mise.toml` when generation behavior matters.
2. Locate the operation by HTTP method/path in `docs/api/swagger/<service>/api/v3/`; for message-only or callback definitions, inspect the corresponding schema, including `docs/api/swagger/backend/callback/`, rather than inventing an HTTP operation. Inspect `apigw/resources.yaml` only if the task reaches the exported gateway surface. Shared messages may affect multiple operations: identify their affected consumers, not unrelated endpoints.
3. Verify business meaning in the relevant `pkg/types`, `pkg/proto` converters/validation, handler/service, or existing API reference. A Proto type or an old comment alone does not establish requiredness, valid values, defaults, permissions, or task-completion semantics.
4. Before choosing comment paragraphs, punctuation, or explaining tags/import results, read [references/projection.md](references/projection.md). Distinguish the installed generator, the cited APIGW implementation, and the deployed gateway version.

## Write for API callers

Use Chinese for caller-facing RPC, Message, and field documentation. Keep identifiers and established terms such as Agent, Proxy, and `bk_host_id` unchanged. This is a scoped exception to the general English-comment rule, not a change to other source directories or bilingual API-reference requirements.

| Element                  | Explain when relevant                                                                  | Avoid                                                                           |
| ------------------------ | -------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| RPC                      | Action, business object, scope; important preconditions or result meaning              | `Foo provides foo`, internal call chains, promises not backed by implementation |
| Request/response Message | What conditions, inputs, or results the object represents                              | `FooReq describes the request`, merely translating “request body”               |
| Field                    | Business referent; confirmed values, units, conditional use, absence/default semantics | Repeating the field name/type, guessing a required flag or enum from the UI     |
| Service                  | Its caller-relevant business scope if an explanation is useful                         | Renaming the Service to change its displayed tag                                |

Do not require a fixed checklist in every comment. Short, unambiguous fields can have short explanations; document non-obvious distinctions rather than padding. Do not add a symbol-name prefix merely to imitate Go lint comments. Comments describe behavior; they do not implement validation or create OpenAPI `required`/`enum` constraints.

### RPC paragraphs

- Start with one concise, complete purpose statement. Add further paragraphs only when callers need more explanation.
- A blank comment line separates Swagger `summary` from `description`; ordinary line wrapping does not.
- The checked APIGW importer prefers nonempty `description` to `summary`. A later paragraph must therefore stand alone as a resource description, not read “also supports…” without explaining what the interface does.
- A complete single paragraph is fine for a simple operation. Do not force every RPC into one paragraph, duplicate the same text into options, or strip necessary limitations just to shorten it. Long calling instructions belong to the full API reference handled by `api-doc`.

For the confirmed purpose of `NodeAgentInstall`, replace `// NodeAgentInstall installs node agent.` with `// 批量安装节点 Agent。` without changing the RPC or its HTTP binding. Add manual-mode, version, or asynchronous-result details only after checking their actual semantics.

For a multi-paragraph comment, make the second paragraph self-contained. Given a confirmed batch-status query, `// 也支持批量查询。` is insufficient; `// 按主机 ID 列表查询节点 Agent 状态。` conveys the purpose independently. These are writing examples, not new endpoint contracts.

### Preserve naming and tags

Follow `proto/README.md`: PascalCase for Message/Service/RPC names, snake_case for fields, and the project's `int64`/`uint64` defaults for new integers. Preserve existing field numbers, wire/JSON names, types, and HTTP bindings in documentation-only work; do not normalize legacy declarations.

With current defaults, **Service name → Swagger operation tag → APIGW resource label**, and **Service + `_` + RPC name → `operationId` → resource name / API-doc filename**. For example, `NodeAgent` / `NodeAgentInstall` yields tag `NodeAgent` and ID `NodeAgent_NodeAgentInstall`.

Keep existing tags and categories. Explain a discrepancy between generated and exported labels without automatically “repairing” either; a prior gateway edit or export may be involved. If an explicitly requested rename is necessary, first identify affected Swagger IDs, `apigw/apidocs/{zh,en}` filenames, and `apigw/definition.yaml` resource-name references, and obtain approval for that scope. See `bk-nodemgr-naming-boundary` for a domain-term decision.

## Generate and inspect the result

For an actual Proto change, use the standard entrypoint from `proto/`: **`make clean && make all`**. It formats sources and changes generated files; check the working tree first and preserve unrelated user edits. Do not substitute a bare `protoc` command or hand-edit Swagger/`*.pb.go` to fix documentation. For a review or discussion without edits, inspect existing artifacts; do not regenerate by default.

Check the touched operations and affected shared schemas, not just the command's exit code:

- RPC `summary`/`description` and schema `title`/`description` contain readable Chinese; no mechanical filler, implementation notes, unsupported claims, or writer-facing TODOs appear in the published descriptions.
- Paragraphs and punctuation land in the intended keys. In particular, schema handling distinguishes ASCII `.` from Chinese `。`; do not apply that rule to RPC `summary` or impose a punctuation migration.
- Field names and effective types match the contract and project post-processing. Request-body `required: true` is not a declaration that every body field is required.
- Documentation-only edits preserve method/path, field numbers/types, `operationId`, and tags. For a real contract change, compare against its approved behavior instead.
- Generated diffs contain no unintended semantic changes. Do not turn formatting-only drift into unrelated cleanup.

When gateway validation is needed, hand off the specific operations and expected description, labels, and schemas for manual import/export inspection. Do not patch `resources.yaml` as a substitute for fixing Proto documentation; separately authorized gateway-policy work belongs to `bk-nodemgr-apigw-sync`. An upstream mapping reference or a local generation pass is not evidence of a successful import into the deployed gateway.

Keep essential unknowns with the author, not in caller-facing comments. Verify or ask about a material contract gap before publishing the affected claim; omit nonessential unverified assertions. Use `bk-nodemgr-audience-boundary` for this distinction.

## Hand off adjacent work

| Task                                                               | Companion skill                                                                   |
| ------------------------------------------------------------------ | --------------------------------------------------------------------------------- |
| Endpoint scaffolding, converters, router and handler wiring        | `api-scaffold`; use this skill for its Proto step                                 |
| Full bilingual endpoint reference                                  | `api-doc`; use source comments as evidence, not a replacement for semantic checks |
| Exported gateway inputs, policy, sync-image/deployment integration | `bk-nodemgr-apigw-sync`                                                           |
| Domain terminology across API/model layers                         | `bk-nodemgr-naming-boundary`                                                      |

Deliver only the paths/comments changed or review findings, relevant projection/compatibility effects, checks actually performed, and any concrete author decision or manual handoff still needed. Do not add a mandatory full-gateway audit report.

## Trial scenarios

The three focused scenarios are in [evals/evals.json](evals/evals.json): caller-facing comments, multi-paragraph projection, and documentation-only stability. Run trials in an isolated copy or temporary directory; keep generated artifacts and trial reports outside the repository. They do not require a full benchmark or live APIGW operations.
