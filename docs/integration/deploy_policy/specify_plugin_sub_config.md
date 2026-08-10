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

The example uses `instance` scope with `host` granularity. Set the shell variables first, and append the authentication and tenant headers required by your deployment.

```bash
curl -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/create" \
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
```

Create returns `data.deploy_policy_id`. Execute the policy with that ID only after confirming your current bk-nodemgr version supports automatic config push for this mode:

```bash
curl -X POST "${BK_NODEMGR_API_BASE}/api/v3/deploy_policy/execute" \
  -H "Content-Type: application/json" \
  -d "{\"deploy_policy_id\": ${DEPLOY_POLICY_ID}}"
```

## System interpretation

The platform interprets this spec as configuration-only desired state for `plugin_name`.

It does not declare plugin installation, plugin package selection, or plugin version upgrade. Use `specify_plugin` or another documented plugin installation flow before using this mode when the plugin may not exist.

## Immediate output

Create returns a response whose `data.deploy_policy_id` identifies the created policy.

Execute returns a response whose `data.trigger_id` identifies the launched execution task.

These immediate outputs are not proof that configuration has been pushed to the target, written to the plugin, reloaded, or made active.

## Eventual or machine-visible artifact

The intended artifact is installed-plugin configuration content matching `config_files_detail` when the current execution chain supports applying this spec.

The current public contract does not confirm automatic configuration push through `deploy_policy` for this mode. Do not promise or depend on machine-side configuration materialization until your target version exposes that behavior as a supported contract.

The deploy-policy contract also does not define file paths, merge-vs-replace behavior, reload behavior, health checks, or timing guarantees.

## Repeat behavior

The concept document only defines this mode as configuration update for an installed plugin.

The public contract does not define whether repeated declarations merge, replace, patch, no-op, retry safely, or roll back configuration.

## Failure cases and limits

- `plugin_name` is required by the type validation.
- The current type validation does not require `config_files_detail`, but a file-content declaration needs at least one meaningful item.
- The target plugin must already be installed; this mode does not install or upgrade it.
- Automatic config push through `deploy_policy` must be confirmed for the target bk-nodemgr version.
- A successful `execute` response only means an execution task was launched, not that configuration is active on the machine.

## Contract references

- [Integration overview](README.md)
- [Spec concept](../../concepts/deploy_policy/spec.md)
- [Scope concept](../../concepts/deploy_policy/scope.md)
- [Swagger contract](../../api/swagger/backend/api/v3/deploy_policy.swagger.json)
