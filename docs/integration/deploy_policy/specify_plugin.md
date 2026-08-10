# specify_plugin

## Purpose and applicability

Use `specify_plugin` when a third-party platform wants target nodes to have a plugin with a specific `plugin_name` and `version`.

The existing concept document defines this mode as: ensure the target node has the specified plugin name and version; if the plugin does not exist, install it; if the version does not match, upgrade it.

## Input

The mode-specific desired-state fields are:

| Field                   | Required | Meaning                                     |
| ----------------------- | -------- | ------------------------------------------- |
| `plugin_name`           | yes      | Plugin name to ensure on the target node    |
| `version`               | yes      | Plugin version to ensure                    |
| `custom_config_context` | no       | Custom values passed as a structured object |

The policy also needs `scopes` so bk-nodemgr can resolve target nodes. See [scope](../../concepts/deploy_policy/scope.md) for supported scope forms.

## Minimal payload and curl template

The example uses `instance` scope with `host` granularity. It requires `curl` and `jq`. Set the deployment-specific API base URL and target identifiers first.

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_BIZ_ID=2
export BK_HOST_ID=10001

CREATE_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/create" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
  {
    "name": "ensure-gse-plugin-example",
    "description": "Ensure a named plugin version on selected hosts",
    "enabled": true,
    "scopes": [
      {
        "type": "instance",
        "scope": {
          "granularity": "host",
          "bk_biz_id": ${BK_BIZ_ID},
          "instance_ids": [${BK_HOST_ID}]
        }
      }
    ],
    "specs": [
      {
        "type": "specify_plugin",
        "param": {
          "plugin_name": "example_plugin",
          "version": "1.0.0",
          "custom_config_context": {
            "env": "prod"
          }
        }
      }
    ]
  }
EOF
)"

DEPLOY_POLICY_ID="$(printf '%s' "${CREATE_RESPONSE}" | jq -r '.data.deploy_policy_id')"
```

Execute the policy with the returned ID and capture `data.trigger_id`:

```bash
EXECUTE_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/execute" \
  -H "Content-Type: application/json" \
  -d "{\"deploy_policy_id\": ${DEPLOY_POLICY_ID}}")"

TRIGGER_ID="$(printf '%s' "${EXECUTE_RESPONSE}" | jq -r '.data.trigger_id')"
```

## System interpretation

The platform interprets this spec as a desired plugin state: each target selected by `scopes` should have `plugin_name` at `version`.

Per the concept document, a missing plugin leads to installation and a mismatched version leads to upgrade. The public contract does not describe exact package lookup, transfer, restart, or health-check steps.

## Immediate output

Create returns a response whose `data.deploy_policy_id` identifies the created policy.

Execute returns a response whose `data.trigger_id` identifies the launched execution task.

These immediate outputs are not final machine-state proof.

## Eventual or machine-visible artifact

The intended eventual artifact is a target node with the named plugin at the requested version.

The deploy-policy API contract does not define public machine paths, process names, reload behavior, or health-check commands. Do not build integration logic that depends on such details unless your deployment exposes another documented contract.

## Repeat behavior

The documented mode semantics are desired-state based: if the plugin already exists at the requested version, the desired state is already satisfied.

The public contract does not define an idempotency key, retry window, rollback rule, or timing guarantee for repeated API calls.

## Failure cases and limits

- Request validation requires `plugin_name`.
- Request validation requires `version`.
- Invalid `scope` prevents target resolution.
- A successful `execute` response only means an execution task was launched.
- Package source, target reachability, and runtime health are outside the schema-level response contract.

## Contract references

- [Integration overview](README.md)
- [Spec concept](../../concepts/deploy_policy/spec.md)
- [Scope concept](../../concepts/deploy_policy/scope.md)
- [Swagger contract](../../api/swagger/backend/api/v3/deploy_policy.swagger.json)
- [Proto variant definitions](../../../proto/backend/api/v3/deploy_policy.proto)
