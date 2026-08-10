# specify_plugin_sub_config

## Purpose and applicability

Use `specify_plugin_sub_config` when a third-party platform wants to declare configuration for a plugin that is already installed.

The existing concept document defines this mode as: update installed plugin configuration file content. It only updates configuration and does not install or upgrade the plugin version.

## Input

The mode-specific desired-state fields are:

| Field                   | Required | Meaning                                                                      |
| ----------------------- | -------- | ---------------------------------------------------------------------------- |
| `plugin_name`           | yes      | Installed plugin whose configuration is being declared                       |
| `config_files_detail`   | no       | Configuration file details to update; provide it when declaring file content |
| `custom_config_context` | no       | Custom values passed as a structured object                                  |

Each `config_files_detail` item follows the proto-documented shape:

| Field            | Meaning                                     |
| ---------------- | ------------------------------------------- |
| `name`           | Configuration file name                     |
| `content`        | Configuration file content                  |
| `is_main_config` | Whether this item is the main configuration |

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
    "name": "apply-plugin-config-example",
    "description": "Declare configuration for an installed plugin",
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
        "type": "specify_plugin_sub_config",
        "param": {
          "plugin_name": "example_plugin",
          "config_files_detail": [
            {
              "name": "example.conf",
              "content": "key=value",
              "is_main_config": true
            }
          ],
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

Create returns `data.deploy_policy_id`, but the current version does not support executing a policy that contains this spec. Keep the ID for policy lookup or update; do not send it to the execute endpoint for configuration delivery.

## System interpretation

The platform interprets this spec as configuration-only desired state for `plugin_name`.

It does not declare plugin installation, plugin package selection, or plugin version upgrade. Use `specify_plugin` or another documented plugin installation flow before using this mode when the plugin may not exist.

## Immediate output

Create returns a response whose `data.deploy_policy_id` identifies the created policy.

There is no successful execute output for this mode in the current version. Creating the policy does not prove that configuration has been pushed to the target, written to the plugin, reloaded, or made active.

## Eventual or machine-visible artifact

The current `deploy_policy` execution flow does not materialize this declaration as a machine-side configuration artifact.

The deploy-policy contract also does not define file paths, merge-vs-replace behavior, reload behavior, health checks, or timing guarantees.

## Repeat behavior

The concept document only defines this mode as configuration update for an installed plugin.

The public contract does not define whether repeated declarations merge, replace, patch, no-op, retry safely, or roll back configuration.

## Failure cases and limits

- Request validation requires `plugin_name`.
- Request validation does not require `config_files_detail`, but a file-content declaration needs at least one meaningful item.
- The target plugin must already be installed; this mode does not install or upgrade it.
- The current version does not support executing this spec through `deploy_policy`.

## Contract references

- [Integration overview](README.md)
- [Spec concept](../../concepts/deploy_policy/spec.md)
- [Scope concept](../../concepts/deploy_policy/scope.md)
- [Swagger contract](../../api/swagger/backend/api/v3/deploy_policy.swagger.json)
- [Proto variant definitions](../../../proto/backend/api/v3/deploy_policy.proto)
