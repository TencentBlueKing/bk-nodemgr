# specify_plugin_pkg

## Purpose and applicability

Use `specify_plugin_pkg` when a third-party platform wants target nodes to have a plugin installed from a specified plugin package and version.

The existing concept document defines this mode as: ensure the target node has a plugin from the specified plugin package name and version. The plugin name is generated from the deploy policy ID and module ID. If the plugin does not exist, install it; if the version does not match, upgrade it.

## Input

The mode-specific desired-state fields are:

| Field                   | Required | Meaning                                     |
| ----------------------- | -------- | ------------------------------------------- |
| `plugin_pkg_name`       | yes      | Plugin package name to use as the source    |
| `version`               | yes      | Plugin package version to ensure            |
| `custom_config_context` | no       | Custom values passed as a structured object |

Use `service_instance` scope when your integration needs the generated plugin identity to reflect service/module placement. The concept document states that generated plugin names depend on deploy policy ID and module ID.

## Minimal payload and curl template

The example uses `instance` scope with `service_instance` granularity. Set the shell variables first, and append the authentication and tenant headers required by your deployment.

```bash
curl -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/create" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
  {
    "name": "ensure-plugin-package-example",
    "description": "Ensure a generated plugin instance from a plugin package",
    "enabled": true,
    "scopes": [
      {
        "type": "instance",
        "scope": {
          "granularity": "service_instance",
          "bk_biz_id": ${BK_BIZ_ID},
          "instance_ids": [${BK_SERVICE_INSTANCE_ID}]
        }
      }
    ],
    "specs": [
      {
        "type": "specify_plugin_pkg",
        "param": {
          "plugin_pkg_name": "example_plugin_pkg",
          "version": "1.0.0",
          "custom_config_context": {
            "env": "prod"
          }
        }
      }
    ]
  }
EOF
```

Create returns `data.deploy_policy_id`. Execute the policy with that ID:

```bash
curl -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/execute" \
  -H "Content-Type: application/json" \
  -d "{\"deploy_policy_id\": ${DEPLOY_POLICY_ID}}"
```

## System interpretation

The platform interprets this spec as a desired plugin-package state: each resolved target should have a plugin instance generated from `plugin_pkg_name` at `version`.

The concept document says the plugin name is generated from deploy policy ID and module ID. It does not define a public formula that third-party platforms should reimplement.

## Immediate output

Create returns a response whose `data.deploy_policy_id` identifies the created policy.

Execute returns a response whose `data.trigger_id` identifies the launched execution task.

These immediate outputs are not final proof that the generated plugin instance is already present on every target.

## Eventual or machine-visible artifact

The intended eventual artifact is a generated plugin instance on the target node, based on the requested plugin package name and version.

The public deploy-policy contract does not define machine paths, package cache paths, generated file names, reload behavior, or health checks.

## Repeat behavior

The documented mode semantics are desired-state based: if the generated plugin instance already exists at the requested package version, the desired state is already satisfied.

The public contract does not define an idempotency key, retry window, generated-name stability across future versions, rollback rule, or timing guarantee for repeated API calls.

## Failure cases and limits

- `plugin_pkg_name` is required by the type validation.
- `version` is required by the type validation.
- Invalid `scope` prevents target resolution.
- The generated plugin name is a platform-owned result; callers should not derive or depend on an undocumented formula.
- A successful `execute` response only means an execution task was launched.

## Contract references

- [Integration overview](README.md)
- [Spec concept](../../concepts/deploy_policy/spec.md)
- [Scope concept](../../concepts/deploy_policy/scope.md)
- [Swagger contract](../../api/swagger/backend/api/v3/deploy_policy.swagger.json)
