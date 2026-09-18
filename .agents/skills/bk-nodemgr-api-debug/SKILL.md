---
name: bk-nodemgr-api-debug
description: Use when a development agent inside the bk-nodemgr repository needs to call or debug the Node Manager API Gateway with bk-cli, prepare dry-run requests from local apidocs, query hosts or plugins, execute authorized operations, or trace workflow results. Not for implementing endpoints, publishing gateways, or calling unrelated BlueKing systems.
---

# Reference

Use `bk-cli api` as the transport and the current checkout as the API reference. This skill is for repository development agents, not a standalone API catalog or a replacement for the official CLI skills.

All repository paths below are relative to the bk-nodemgr root. Read only the selected resource and its documentation; do not load or copy the entire API catalog.

## Execution Boundary

Execute a mutation only when the user has explicitly authorized the operation and its target scope. Otherwise, produce only a `--dry-run` request for that mutation.

- Scope includes the deployment/context, gateway stage, tenant, and affected business/hosts/plugins/policies as applicable. Resolve identities from user input, confirmed context, or scoped read-only queries; never substitute documentation example IDs.
- An explicit request to start a named plugin on specified hosts in a confirmed environment is authorization. Do not ask again for the same operation and scope.
- A request to investigate a problem is not permission to restart, install, update, retry, terminate, delete, or otherwise change state. A development context is not automatically a safe environment.
- Classify by documented effects, not HTTP method: many Node Manager queries use POST. Read-only requests may execute within the requested investigation scope once the target is resolved.
- If the target is unresolved, show a clearly marked preview template and ask for the missing value. Do not execute a template with placeholders or silently accept the CLI's default `prod` stage.
- Expanded scope, a different operation, or a different target requires authorization covering that change. A timeout is not authorization to submit a mutation again.

## Official CLI Prerequisites

Read `bk-cli-shared`, then `bk-cli-api`, before constructing requests. Load installed copies when available. If missing, fetch the official Markdown directly; do not assume sibling skill directories exist:

- https://raw.githubusercontent.com/TencentBlueKing/bk-cli/master/skills/bk-cli-shared/SKILL.md
- https://raw.githubusercontent.com/TencentBlueKing/bk-cli/master/skills/bk-cli-api/SKILL.md

Check `bk-cli api --help` against the installed version. Official skills own authentication, URL assembly, flags, quoting, and output envelopes. If their instructions or the needed flags cannot be verified, report the gap instead of inventing syntax or installing/upgrading the CLI automatically.

- Prefer per-request `--context` and `--stage`; do not change the active context as a side effect of debugging. Context selects a BlueKing deployment, not a gateway stage.
- Use the confirmed context tenant, or a confirmed per-request `--header 'X-Bk-Tenant-Id:...'` override. Do not infer the business tenant from gateway-management synchronization settings.
- For local setup diagnosis, use the official `bk-cli doctor --offline` and authentication-status guidance. `auth status` returning `ok: true` does not prove credentials exist; check `data.has_credentials`.
- Do not read credential files, ask users to paste secrets, or put tokens into skill files or reports. CLI authentication-header redaction does not guarantee that body fields or arbitrary headers are safe to print. Redact host login credentials and other sensitive values before sharing dry-run/verbose output.
- Do not weaken TLS verification, bypass authentication, or call backend service URLs to work around gateway failures.

## Find the Actual API

| Source                                                                      | Responsibility                                                                                                |
| --------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `apigw/resources.yaml`                                                      | Gateway entry path, method, `operationId`, request/response schema, and `x-bk-apigateway-resource.authConfig` |
| `apigw/apidocs/zh/<operationId>.md`                                         | Business semantics, required permissions, field constraints, minimum version, examples, and response meaning  |
| `apigw/definition.yaml`                                                     | Repository gateway visibility, stages, and release metadata; not proof of deployment                          |
| `install/images/bk-nodemgr-apigw-sync/support-files/bin/sync-apigateway.sh` | Repository gateway name and publication controls                                                              |

1. Search the resource index by the user's business term, candidate path, or operation name. If necessary, search apidocs descriptions for candidate terms, then verify each candidate in the resource index.
2. Read the selected method's schema and auth configuration. Use the gateway entry path, not the nested backend path or a guessed route.
3. Read the matching apidoc and check method/path agreement. If the filename differs, search by the exact URL; do not treat every Markdown filename as an active gateway resource. For example, verify `Plugin_StartPlugin` rather than trusting a similarly named `PluginAPI_StartPlugin` document.
4. Record the operationId, method/path, read-only or mutation classification, required fields, permissions, and minimum version. Resolve material conflicts before sending a request.

The sync script currently names the gateway `bk-nodemgr`, and the definition marks it `is_public: false`. Public gateway discovery is therefore optional, not a prerequisite. Absence from `bk-cli apigateway list_gateways` or public resource details does not prove an API is unavailable. Do not choose another gateway or make it public to proceed.

Use `bk-cli-apigateway` only when discovery or accessible deployed schema information is needed. Compare available deployment evidence with the checkout: local docs/resources do not prove the target stage has that API version. The repository currently declares `prod`; do not invent a `testing` deployment. If publication/version is unknown, say so. A 404 can reflect stage or resource publication drift, not necessarily a bad local path; do not automatically run synchronization or release commands.

## Construct and Preview

Use `bk-cli api <gateway> <method> <entry-path>` with explicit confirmed context and stage. Put parameters in `--query`, `--path`, or `--body` according to the selected schema, preserving array types, integer IDs, enums, and required fields. Read the endpoint-specific pagination limit; limits are not uniform across APIs.

For the first call to an endpoint or during troubleshooting, preview with `--dry-run` and inspect the resulting URL, scope filters, and body before sending. Dry-run validates local request construction only; it does not validate remote permissions, publication, or business constraints.

Example: `Topo_HostList` is a read-only POST. After confirming the shell variables from the user's target, this previews the first 20 hosts of that business:

```bash
bk-cli api bk-nodemgr POST /api/v3/topo/host/list \
  --context "${CONTEXT:?confirm context}" \
  --stage "${STAGE:?confirm stage}" \
  --header "X-Bk-Tenant-Id:${TENANT_ID:?confirm tenant}" \
  --body "$(jq -cn --argjson biz_id "${BK_BIZ_ID:?confirm business ID}" \
    '{page:{offset:0,limit:20},exact_include_conditions:{bk_biz_id:[$biz_id]}}')" \
  --dry-run
```

`jq` is used here only to construct typed JSON; it is not required by bk-cli itself. Do not copy the apidoc's example `running` filter into a fault investigation, since it can exclude the failing hosts.

For a user who requests a complete list, paginate with the documented offset/limit and response fields, keeping the same scope. Report fetched count versus total and any incomplete/truncated result; never label one page as all resources. If an empty page or error prevents completion, stop and report rather than looping indefinitely.

## Responses and Completion

Keep stdout JSON separate from stderr diagnostics. For documented Node Manager JSON responses, there are two envelopes:

| Layer            | Check                                                                                                                  |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------- |
| CLI/HTTP         | Exit status and outer `ok`/`status`; CLI errors may only appear on stderr                                              |
| Node Manager     | Inner `.data.code == 0`; inspect `.data.message`, `.data.error`, `.data.permission`, and `.data.request_id` on failure |
| Business payload | `.data.data`, for example `.data.data.items`, `.data.data.total`, or `.data.data.workflow_id`                          |

HTTP 200 and outer `ok: true` alone do not mean business success. Dry-run output has a `request`, not a business result. Download or other nonstandard responses must follow their own endpoint contract, not this JSON assumption.

Check both gateway identity/resource requirements and the apidoc's business permissions. `userVerifiedRequired: false` does not waive IAM checks. A 403 needs evidence from gateway headers/body versus the backend error/permission payload; do not automatically conclude that logging in again or applying gateway permission fixes every denial.

A returned `workflow_id` means submission succeeded, not that the requested change completed. For plugin start, read `apigw/apidocs/zh/PluginWorkflow_PluginWorkflowList.md` and verify its resource entry; query by the returned workflow ID, then inspect operation/instance details if needed. Node, package, and deploy-policy workflows have their own APIs: retrieve their docs instead of reusing the plugin query blindly. Preserve `trigger_id` when returned; do not substitute it for a workflow ID.

Use bounded polling for read-only status checks and report running, success, failed, partial failure, or unknown from actual results. Do not invent terminal states. After a mutation times out or disconnects, use scoped status/history queries to reconcile the result before proposing a resubmission; do not blindly retry a possibly accepted operation. Retry/terminate APIs are mutations subject to the same authorization rule.

## Delivery and Adjacent Tasks

Report the confirmed target, operationId and method/path, whether the request was only prepared, dry-run, or sent, the business result, and relevant request/workflow IDs. Include the local apidoc path and any unresolved deployment mismatch. Keep secrets and unnecessary host details out of the report.

- For deploy-policy probe payloads and domain-specific debugging, also use `bk-nodemgr-dpmgr-test-debug`; this skill still owns gateway invocation and authorization scope.
- For adding endpoints or writing API docs, use `api-scaffold` or `api-doc` instead. Do not turn a failed call into an unrequested code change or gateway publication.
