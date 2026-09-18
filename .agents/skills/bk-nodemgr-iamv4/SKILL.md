---
name: bk-nodemgr-iamv4
description: Use when editing, reviewing, explaining, or designing bk-nodemgr IAM V4 runtime authorization, resource callbacks, manager role grants, IAM V4 model sync, bkiamv4 render/migrate files, callback URL Helm wiring, or boundaries between auth/v4, router/api-v3/iam/v4, pkg/thirdparty/iamv4, and support-files/bkiamv4.
---

# bk-nodemgr IAM V4

## Overview

IAM V4 in bk-nodemgr is split across runtime authorization, IAM resource callbacks, role grants, third-party runtime APIs, and the model sync toolchain. Use this skill to keep those parts connected without collapsing their boundaries.

This skill is not a static navigation page. It gives the project-specific judgment frame for IAM V4 work: inspect the source anchors, decide which path is touched, preserve layer ownership, and report the impact surface when changing behavior.

## When to Use

Use this skill when work touches:

- IAM V4 enablement, disabled handlers, authorizer wiring, or manager role grant wiring.
- `internal/backend/auth/v4/**` permission checks, resource conversion, apply URL construction, authorized scopes, or role grants.
- `internal/backend/router/api-v3/iam/v4/**` resource callback transport, Basic Auth, tenant handling, request IDs, callback errors, or dispatch.
- `pkg/thirdparty/iamv4/**` runtime IAM V4 API calls, system token cache, callback credential validation, or IAM V4 wire request/response types.
- `support-files/bkiamv4/**` render, migrate, IAM V4 model API adapter, model templates, or migration policy.
- Helm values/templates that render IAM V4 callback URLs or run `bkiamv4-sync`.
- Action IDs, resource type IDs, resource attributes, `_bk_iam_path_`, callback fields, IAM V4 tenant semantics, or model/runtime drift.

## When Not to Use

Do not use this skill as the primary guide for:

- Generic IAM concepts with no bk-nodemgr IAM V4 code path.
- IAM V3-specific work; keep IAM V3 in a separate skill.
- Ordinary business router permission mapping outside IAM V4 callback/runtime boundaries; use `router-permission-supplement`.
- Generic third-party adapter design with no IAM V4 runtime or model-sync contract; use `bk-nodemgr-thirdparty`.
- Generic Go context, logging, or error handling unless IAM V4 code is the touched surface.

## Source Anchors

Read the relevant anchors before changing behavior. Do not rely on this skill alone; verify current code first.

Runtime wiring:

- `internal/backend/service/service.go`: `newAuthorizer`, `newIAMV4Handler`, IAM V3/V4 mutual exclusion, disabled handlers, `ManagerRoleGranter` wiring.
- `internal/backend/options/capability.go`: `IAMV4Handler`, `Authorizer`, `ManagerRoleGranter`, and `AuthProviderV4Handler` ownership.

Runtime authorization:

- `internal/backend/auth/v4/auth_iamv4.go`: `NewIAMV4Authorizer`, permission checks, denied-resource collection, authorized-scope queries.
- `internal/backend/auth/v4/auth_iamv4_helpers.go`: IAM V4 resource conversion, apply URL resource paths, `_bk_iam_path_` compatibility.
- `internal/backend/auth/v4/manager_role.go`: manager role grant behavior.

Resource callback:

- `internal/backend/router/api-v3/iam/v4/v4.go`: `/api/v3/iam/v4/resource`, Basic Auth, tenant/request-id handling, callback error envelope, dispatcher call.
- `apigw/apidocs/zh/IAM_IAMV4ResourceCallback.md`: IAM V4 callback protocol, supported methods, page semantics, Basic Auth, error codes.

Runtime IAM V4 adapter:

- `pkg/thirdparty/iamv4/AGENTS.md`: runtime adapter contract and anti-patterns.
- `pkg/thirdparty/iamv4/handler.go`: `IHandler`, auth/token/role scenario interfaces, token cache, apply URL conversion.
- `pkg/thirdparty/iamv4/iamv4.go`: raw IAM V4 runtime endpoints, headers, status and envelope handling.
- `pkg/thirdparty/iamv4/types.go`: IAM V4 wire request/response types and config validation.

Model sync:

- `support-files/bkiamv4/migrate/iamv4/AGENTS.md`: model API adapter contract, separate from runtime authorization adapter.
- `support-files/bkiamv4/migrate/iamv4/handler.go`: scenario interfaces, conversion, pagination, aggregation, completeness checks.
- `support-files/bkiamv4/migrate/iamv4/iamv4.go`: single-request IAM V4 model API methods, headers, response decoding.
- `support-files/bkiamv4/migrate/iamv4/types.go`: named model API request/response types and optional-field semantics.
- `support-files/bkiamv4/migrate/migrate.go` and `support-files/bkiamv4/migrate/resource_type.go`: migration ordering, upsert decisions, dry-run, allowed-update policy.
- `install/helm/bk-nodemgr/templates/bkiamv4-sync-configmap.yaml`: callback URL rendering into `vars.yaml`.
- `install/helm/bk-nodemgr/templates/bkiamv4-sync-job.yaml`: `iam-render` and `iam-migrate` job wiring.

## Relationship Map

| Path               | Main owners                                                                  | Contract                                                                                                                          |
| ------------------ | ---------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| Service wiring     | `internal/backend/service`, `internal/backend/options`                       | Select IAM V3 or IAM V4, install disabled handlers for the inactive mode, build IAM V4 handler and authorizer, wire role granter. |
| Permission check   | `internal/backend/auth/v4`, `pkg/thirdparty/iamv4`                           | Convert bk-nodemgr action/resource requests into IAM V4 runtime auth calls; keep wire DTOs inside `pkg/thirdparty/iamv4`.         |
| Resource callback  | `internal/backend/router/api-v3/iam/v4`, `internal/backend/auth/v4/provider` | Authenticate IAM callback, validate transport request, dispatch resource queries, return IAM V4 callback envelope.                |
| Manager role grant | `internal/backend/auth/v4`, `pkg/thirdparty/iamv4`                           | Grant concrete IAM V4 role authorization through the runtime adapter, separate from read-only permission checks.                  |
| Runtime adapter    | `pkg/thirdparty/iamv4`                                                       | Own runtime IAM V4 APIGW endpoints, headers, token cache, callback credential validation, and wire request/response types.        |
| Model sync         | `support-files/bkiamv4`, Helm templates                                      | Render IAM V4 model vars/templates and migrate model definitions with model API adapter; do not perform runtime authorization.    |

## Evidence Pass

Before editing or reviewing IAM V4 work:

1. Identify which paths are touched: service wiring, permission check, callback, role grant, runtime adapter, or model sync.
2. Read the nearest `AGENTS.md` for every touched path.
3. Inspect the relevant source anchors and at least one comparable implementation in the same path.
4. Decide whether action IDs, resource type IDs, attributes, `_bk_iam_path_`, callback URL, tenant semantics, or IAM model definitions changed.
5. If model/runtime drift is possible, inspect both runtime and `support-files/bkiamv4` before recommending code.

Facts from this skill are starting points. Current code and scoped instructions win over stale memory.

## Hard Boundaries

- `pkg/thirdparty/iamv4` is the runtime authorization adapter. It exposes scenario interfaces through `IHandler`; internal packages must not construct IAM V4 wire DTOs or raw endpoint paths.
- `support-files/bkiamv4/migrate/iamv4` is the model API adapter. Keep it separate from `pkg/thirdparty/iamv4`; do not reuse runtime-only DTOs, token cache, callback credential logic, or HTTP client internals.
- `internal/backend/router/api-v3/iam/v4` owns IAM V4 callback transport only: Basic Auth, tenant/request-id propagation, request validation, error envelope, and dispatch. It must not make business authorization decisions.
- `internal/backend/auth/v4` owns the mapping from bk-nodemgr `auth.IAuthorizer` semantics to IAM V4 handler/provider calls.
- `internal/backend/service/service.go` and `internal/backend/options/capability.go` own IAM V4 enablement, disabled-handler selection, authorizer construction, and manager role granter wiring.
- Helm and `support-files/bkiamv4` own IAM V4 model sync and callback URL rendering. They must not implement runtime permission checks or callback credential validation.

## Runtime Path Checks

### Service wiring

Check `newAuthorizer` before changing IAM enablement behavior. IAM V3 and IAM V4 are mutually exclusive; inactive handlers are intentionally disabled. Do not bypass disabled handlers to make direct IAM calls work.

### Permission check

Keep business authorization flow in `internal/backend/auth/v4`. The authorizer builds project-level requests, enriches attributes through the provider, and calls the runtime IAM V4 handler. Do not leak IAM V4 wire request structs into handlers, services, or storage.

### Resource callback

IAM V4 callback is a protocol boundary. Preserve Basic Auth semantics, tenant handling, request-id response headers, supported methods, page validation, `requires`, and IAM V4 error envelope. Business resource query dispatch belongs behind the provider dispatcher.

### Manager role grant

Role grants are writes to IAM and are separate from read-only permission checks. Validate concrete subject, role, related resource type, single resource, and expiration through the runtime adapter contract. External writes require explicit user approval when operating against a real IAM system.

## Model Sync Checks

Model sync becomes part of the impact surface when a change affects:

- Action IDs or action semantics.
- Resource type IDs, resource hierarchy, parent-child relationships, or attributes.
- `_bk_iam_path_` compatibility behavior.
- Callback URL, callback path, or IAM system metadata.
- IAM V4 model templates, vars, render output, or migration rules.

The model adapter retrieves and updates IAM model definitions. Raw model API methods perform one request; handler-level code owns pagination, aggregation, and completeness checks. Migration policy such as dry-run behavior, allowed updates, topology-drift rejection, and ordering stays in the runner.

## Output Contract

For implementation, refactor, or review tasks, include a compact impact map before or after the change:

```text
IAMV4 Impact Map
Intent:
Runtime paths touched:
Model/sync paths touched:
Boundary decision:
Contract risks:
Verification:
```

For explanation-only tasks, keep the answer shorter but still name the relevant paths and source anchors. Do not force the full template when the user only asks for a conceptual explanation.

For review tasks, lead with findings first. Map findings to boundary drift, model/runtime drift, callback protocol drift, tenant/auth drift, or verification gaps.

## Common Mistakes

- Treating IAM V4 as one package and moving runtime, callback, and model migration logic into a shared helper.
- Updating runtime action/resource conversion without checking whether IAM model templates or migration policy also changed.
- Reusing `pkg/thirdparty/iamv4` runtime DTOs in `support-files/bkiamv4/migrate/iamv4`, or the reverse.
- Putting authorization decisions into the callback router instead of dispatching resource queries through the provider.
- Adding IAM V4 fields or Helm knobs for anticipated future UI behavior without a current endpoint or model contract.
- Returning IAM V3-style callback envelopes from the IAM V4 callback path.
- Logging callback credentials, auth headers, raw remote error bodies, or model sync secrets.
- Skipping tenant checks because single-tenant mode worked locally.

## Verification Checklist

- Search current IAM V4 source anchors before editing; do not rely on remembered paths.
- Confirm scoped `AGENTS.md` rules for every touched subtree.
- For runtime changes, verify service wiring, auth/v4 conversion, callback protocol, and runtime adapter contracts as applicable.
- For action/resource/model changes, verify `support-files/bkiamv4` templates, render/migrate behavior, and Helm callback URL wiring.
- For callback changes, verify Basic Auth, tenant, page/page_size, `requires`, supported methods, request-id response header, and error envelope.
- For third-party adapter changes, use named request/response types, existing APIGW/rest client behavior, context propagation, and wrapped errors.
- Run scoped checks justified by the touched path, for example `go test ./support-files/bkiamv4/migrate/...` for migrate changes or the relevant backend package tests for runtime changes.

## Cross-References

- `bk-nodemgr-thirdparty`: use for `pkg/thirdparty/iamv4` or `support-files/bkiamv4/migrate/iamv4` adapter signatures, handler/client separation, pagination, and protocol DTOs.
- `bk-nodemgr-contextx`: use for tenant, user, message ID, cancellation, callback contexts, and third-party calls.
- `bk-nodemgr-error-handling`: use for wrapped errors, callback error envelopes, and remote API failure semantics.
- `bk-nodemgr-logger`: use before adding IAM V4 logs or changing secret-bearing diagnostics.
- `bk-nodemgr-architecture-judgment`: use for cross-layer boundary decisions, runtime/model sync drift, new shared helpers, or unapproved architecture departures.
- `router-permission-supplement`: use for ordinary business router permission action/resource mapping, not the IAM V4 callback endpoint itself.

## Evals

Pressure prompts live in `evals/evals.json`. Keep run outputs, timing, grading, benchmarks, and review artifacts outside git unless explicitly requested.
