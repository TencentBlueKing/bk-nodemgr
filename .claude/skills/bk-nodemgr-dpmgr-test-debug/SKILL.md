---
name: bk-nodemgr-dpmgr-test-debug
description: Use when testing or debugging bk-nodemgr dpmgr deploy policy behavior with bk-nodemgr-deploy-policy-probe, especially when users ask for standard JSON fragments, Bruno JSON, deploy policy create/update/execute payload parts, plugin package import payloads, config filename checks, trigger_id/workflow tracing, or logs for specify_plugin_sub_config and related plugin deploy specs.
---

# bk-nodemgr dpmgr Test Debug

## Overview

Use this skill to help a user exercise dpmgr deploy policy behavior with the long-lived probe plugin package under `testsuite/plugins/deploy-policy/bk-nodemgr-deploy-policy-probe`.

The default product is machine-usable JSON. The user assembles API requests unless they explicitly ask for a complete request body, prose, curl, or a different shape.

## When To Use

Use this skill when the user wants to:

- upload or import `bk-nodemgr-deploy-policy-probe` for dpmgr testing
- generate standard JSON for deploy policy tests
- generate Bruno-ready JSON when explicitly requested
- test `specify_plugin`, `specify_plugin_pkg`, or `project_plugin_pkg_to_hosts`
- test `specify_plugin_sub_config`, `specify_plugin_sub_config_template`, or `project_plugin_config_template_to_hosts`
- confirm direct sub-config naming versus deploy-policy-derived naming
- debug dpmgr `trigger_id`, plugin workflow visibility, rendered config content, or plugin start/reload failures

Do not use this skill for:

- generic package upload unrelated to dpmgr deploy policy
- production rollout guidance
- adding new deploy specs or changing dpmgr implementation without a failing scenario
- automated `testsuite/support` integration tests; use `bk-nodemgr-testsuite-support` for that

## Required Repo Facts

Read these before producing final JSON or debug guidance:

- `testsuite/plugins/deploy-policy/bk-nodemgr-deploy-policy-probe/README.md`
- `testsuite/plugins/deploy-policy/bk-nodemgr-deploy-policy-probe/bk-nodemgr-deploy-policy-probe-*.tgz.sha256`
- `testsuite/plugins/deploy-policy/bk-nodemgr-deploy-policy-probe/package/bk-nodemgr-deploy-policy-probe/project.yaml`
- `testsuite/plugins/deploy-policy/bk-nodemgr-deploy-policy-probe/package/bk-nodemgr-deploy-policy-probe/plugins_linux_x86_64/definition.yaml`
- `docs/integration/deploy_policy/README.md`
- `docs/concepts/deploy_policy/spec.md`

Compute the package `md5` from the current `.tgz` before giving import JSON. Do not trust old conversation values or memory for checksums.

## Scenario Boundary

The probe plugin covers executable plugin-related deploy specs:

| Deploy spec | Probe purpose |
| --- | --- |
| `specify_plugin` | install or upgrade a fixed plugin instance |
| `specify_plugin_pkg` | create a deploy-policy-generated plugin instance from a package |
| `project_plugin_pkg_to_hosts` | project a source service instance package to placement hosts |
| `specify_plugin_sub_config` | apply a direct explicit sub-config name from a source template |
| `specify_plugin_sub_config_template` | apply a deploy-policy-derived sub-config name |
| `project_plugin_config_template_to_hosts` | project source config template to placement hosts with source identity in the filename |

It does not cover `specify_agent`, `specify_proxy`, or unsupported `specify_plugin_pkg_sub_config`.

## JSON Output Contract

Default to JSON only. Do not add Markdown headings, prose, or code fences unless the user explicitly asks for them.

User-specific output instructions override this default. If the user asks for a complete create body, curl command, explanation, or a different envelope, follow that request.

Default response shape for test JSON:

```json
{
  "scenario": "<scenario-name>",
  "json_scope": "spec_param",
  "endpoint": "<optional-api-path>",
  "body": {},
  "placeholders": [],
  "assertions": []
}
```

Use `json_scope` values consistently:

| `json_scope` | Use when |
| --- | --- |
| `spec_param` | default deploy policy `specs[].param` fragment |
| `api_body` | user explicitly asks for a complete API request body |
| `import_body` | plugin package import body |
| `debug_report` | user pastes logs or runtime symptoms |

Use placeholders only for values the user must supply:

- `{{pkg_download_url}}`
- `{{import_workflow_id}}`
- `{{bk_biz_id}}`
- `{{bk_host_id}}`
- `{{bk_service_instance_id}}`
- `{{placement_host_id}}`
- `{{deploy_policy_id}}`
- `{{source_module_id}}`
- `{{source_host_id}}`

Do not invent concrete host IDs, service instance IDs, module IDs, or deploy policy IDs.

## Standard Test Flow

1. Import the plugin package with `/api/v3/package/workflow/import/v3/plugin`.
2. Query import result with `/api/v3/package/workflow/import_result`.
3. Create or update one deploy policy per scenario.
4. Execute with `/api/v3/deploy_policy/execute`.
5. Use `/api/v3/deploy_policy/list` and workflow APIs to capture policy IDs and trigger IDs.
6. Check rendered configs and plugin process state on target hosts.

Run install specs before sub-config specs. Sub-config specs need an existing plugin instance on the target host.

## Standard Import JSON

Return this shape after computing the current `md5`:

```json
{
  "scenario": "import-probe-plugin-package",
  "json_scope": "import_body",
  "endpoint": "/api/v3/package/workflow/import/v3/plugin",
  "body": {
    "filename": "bk-nodemgr-deploy-policy-probe-<version>.tgz",
    "download_url": "{{pkg_download_url}}",
    "md5": "<computed_md5>"
  },
  "placeholders": ["{{pkg_download_url}}"],
  "assertions": ["import workflow is created"]
}
```

Import result shape:

```json
{
  "scenario": "query-import-result",
  "json_scope": "api_body",
  "endpoint": "/api/v3/package/workflow/import_result",
  "body": {
    "workflow_id": "{{import_workflow_id}}"
  },
  "placeholders": ["{{import_workflow_id}}"],
  "assertions": ["package import completes successfully"]
}
```

## Config Naming Assertions

Use these assertions when interpreting results:

| Deploy spec | Expected config name | Must not happen |
| --- | --- | --- |
| `specify_plugin_sub_config` | `explicit-direct-subconfig.conf` | `direct-subconfig-template_deploy_<deploy_policy_id>.conf` is created |
| `specify_plugin_sub_config_template` | `policy-managed-subconfig_deploy_<deploy_policy_id>.conf` | the direct explicit name is reused |
| `project_plugin_config_template_to_hosts` | `projected-source-subconfig_deploy_<deploy_policy_id>_<source_module_id>_<source_host_id>.conf` | source module or source host identity is missing |

For direct sub-config, `template_name` selects the source template and request `name` is the final config name. Do not rewrite it as a deploy-policy-derived name.

## Debug JSON

Default debug response shape:

```json
{
  "scenario": "<debug-scenario>",
  "json_scope": "debug_report",
  "symptom": "<one-line symptom>",
  "evidence": [],
  "likely_area": "<area>",
  "next_checks": [],
  "blocked_on": []
}
```

Start from the exact failed surface and preserve IDs.

| Symptom | First checks | Likely area |
| --- | --- | --- |
| package import fails | filename, `download_url`, computed md5, import result workflow | package upload or artifact checksum |
| `execute` returns `trigger_id` but no plugin action is visible | parent trigger logs, deploy policy action logs, plugin workflow list filters | workflow fan-out or plugin workflow metadata |
| direct sub-config rendered with a derived name | `config_files_detail.template_name`, `config_files_detail.name`, dpmgr analyzer logs | `specify_plugin_sub_config` semantics |
| template-derived config keeps placeholders | whether `ExpectedFinalName` was updated after `deploy_policy_id` is known | test input data, not plugin package |
| plugin start says process is not running | package version, control commands, official `start.sh`, real executable presence | probe package lifecycle or GSE process trusteeship |
| reload fails after install succeeds | running process state, official `reload.sh`, `SIGUSR1` handling | process lifecycle or control command |

For logs, search around these stable terms:

- `deploy_policy_id`
- `trigger_id`
- `bk-nodemgr-deploy-policy-probe`
- `specified plugin sub config`
- `config_files_detail`
- `/api/v3/transfer/launch/plugin`
- `operate proc failed`

## Response Patterns

### User asks for standard JSON

Return one or more JSON envelopes. Default `body` is the `specs[].param` fragment so the user can assemble the final request.

### User asks for Bruno JSON

Return JSON envelopes with the same `body`, plus `endpoint` when it helps route the request. Avoid curl commands unless the user explicitly asks for curl.

### User asks for a complete API body

Set `json_scope` to `api_body` and include the complete body for the requested endpoint. This is an explicit user override, not the default.

### User pastes `/deploy_policy/list`

Return a `debug_report` JSON object. Extract these facts into `evidence`:

- `deploy_policy_id`
- deploy spec `type`
- scope granularity and target IDs
- plugin name or plugin package name
- version for install specs
- `config_files_detail.template_name`
- `config_files_detail.name`
- `custom_config_context.ExpectedFinalName`

Then put ready-to-execute policies and required updates in `next_checks`.

### User pastes workflow or GSE logs

Return a `debug_report` JSON object. Preserve the exact error line in `evidence`, then give the next two checks in `next_checks`. Do not jump to code changes until the data path proves the failing layer.

## Common Mistakes

| Mistake | Correction |
| --- | --- |
| Giving stale md5 from memory | compute md5 from the current `.tgz` |
| Defaulting to complete create/update bodies | output standard `spec_param` JSON unless the user asks for full bodies |
| Adding prose around JSON by default | return JSON only |
| Testing sub-config before installing the plugin | run a plugin install spec first |
| Treating `execute` `trigger_id` as the plugin workflow ID | trace parent deploy policy workflow to child plugin workflow |
| Leaving `{{deploy_policy_id}}` in strict expected names after policy creation | update expected names with the actual policy ID before strict checks |
| Assuming `specify_plugin_sub_config` should derive the filename | preserve request `name` as final config name |
| Debugging from frontend visibility alone | inspect backend workflow metadata and business filters |

## Quick References

- Probe plugin: `testsuite/plugins/deploy-policy/bk-nodemgr-deploy-policy-probe/`
- Deploy policy integration doc: `docs/integration/deploy_policy/README.md`
- Deploy spec concept doc: `docs/concepts/deploy_policy/spec.md`
- dpmgr code: `internal/backend/dpmgr/`
- Plugin lifecycle scripts: `script_tools/plugin_scripts/v3/linux/`
