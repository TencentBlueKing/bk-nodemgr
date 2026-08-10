# deploy_policy integration

This guide is for third-party platform developers who need to create and execute `deploy_policy` through the bk-nodemgr backend API. It explains the integration flow, the input that your platform must send, the immediate API output, and the machine-visible effect that the policy is intended to produce.

For full field schema, use the Swagger reference: [deploy_policy.swagger.json](../../api/swagger/backend/api/v3/deploy_policy.swagger.json). For domain details, see [scope](../../concepts/deploy_policy/scope.md) and [spec](../../concepts/deploy_policy/spec.md).

## Integration outcome

`deploy_policy` describes a desired state for a group of targets. A third-party platform submits:

- `scope`: which machines or service instances are affected.
- `spec`: what final plugin state those targets should reach.
- `enabled`: whether the policy can be executed.

The API response confirms that the policy was created or that an execution task was launched. It does not, by itself, prove that every target has already reached the desired state.

## API-first quickstart

Use the mode documents for curl templates:

| Goal                                                     | Read                                                      |
| -------------------------------------------------------- | --------------------------------------------------------- |
| Ensure a named plugin and version                        | [specify_plugin](specify_plugin.md)                       |
| Ensure a generated plugin instance from a plugin package | [specify_plugin_pkg](specify_plugin_pkg.md)               |
| Declare configuration for an installed plugin            | [specify_plugin_sub_config](specify_plugin_sub_config.md) |

All modes follow the same API sequence:

1. Build `scopes` from your target selection.
2. Build `specs` from the desired state.
3. Call `POST /api/v3/deploy_policy/create`.
4. Save `data.deploy_policy_id` from the create response.
5. Call `POST /api/v3/deploy_policy/execute` with that `deploy_policy_id`.
6. Save `data.trigger_id` from the execute response for follow-up through the task-observation capability exposed by your deployment.

## Integration flow

### 1. Establish API context

The public deploy-policy endpoints are:

| Step           | Method and path                      | Immediate output                       |
| -------------- | ------------------------------------ | -------------------------------------- |
| Create policy  | `POST /api/v3/deploy_policy/create`  | `data.deploy_policy_id`                |
| Execute policy | `POST /api/v3/deploy_policy/execute` | `data.trigger_id`                      |
| List policies  | `POST /api/v3/deploy_policy/list`    | `data.total`, `data.items`             |
| Update policy  | `POST /api/v3/deploy_policy/update`  | see Swagger for the response structure |

Authentication headers, tenant selection, and gateway base URL depend on your deployment. Do not hard-code values from examples.

### 2. Convert targets into scope

`scope` tells bk-nodemgr how to find target machines or service instances. The supported target result types are documented in [scope](../../concepts/deploy_policy/scope.md):

| Scope type         | Can produce host targets | Can produce service_instance targets |
| ------------------ | ------------------------ | ------------------------------------ |
| `topo`             | yes                      | yes                                  |
| `set_template`     | yes                      | yes                                  |
| `service_template` | yes                      | yes                                  |
| `instance`         | yes                      | yes                                  |
| `dynamic_group`    | yes                      | no                                   |

Use `host` granularity when the policy should apply directly to machines. Use `service_instance` granularity when the policy must distinguish module/service-instance placement.

### 3. Declare desired state

`spec` defines the desired final state for every target selected by `scope`. The public concept document defines `spec` as the expected state that targets should reach.

#### specify_plugin

Use `specify_plugin` when your platform wants target nodes to have a plugin with a specific name and version. The documented behavior is: if the plugin does not exist, install it; if the version differs, upgrade it.

Details and curl: [specify_plugin](specify_plugin.md).

#### specify_plugin_pkg

Use `specify_plugin_pkg` when your platform provides a plugin package name and version. The documented behavior is: target nodes should have a plugin installed from the specified package and version. The plugin name is generated from the deploy policy ID and module ID.

Details and curl: [specify_plugin_pkg](specify_plugin_pkg.md).

#### specify_plugin_sub_config

Use `specify_plugin_sub_config` only to declare configuration for an already installed plugin. The documented behavior is config-only: it updates plugin configuration file content and does not install or upgrade the plugin version.

The current public contract does not prove that `deploy_policy` execution automatically pushes this config to every target. Treat automatic config push as a version-specific capability that must be confirmed before relying on it.

Details and curl: [specify_plugin_sub_config](specify_plugin_sub_config.md).

### 4. Submit the policy

`POST /api/v3/deploy_policy/create` accepts:

- `name`
- `description`
- `enabled`
- `specs`
- `scopes`

The create response contains `data.deploy_policy_id`. Store it as the policy identity for later execution or update.

### 5. Interpret the immediate output

`POST /api/v3/deploy_policy/execute` accepts `deploy_policy_id` and returns `data.trigger_id`.

`trigger_id` means the platform launched an execution task. It is not the same as final plugin health, final process status, or guaranteed machine convergence.

### 6. Observe the eventual artifact

Machine-visible effects depend on the spec mode:

| Mode                        | Intended eventual effect                                                                                              |
| --------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `specify_plugin`            | target has the named plugin at the requested version                                                                  |
| `specify_plugin_pkg`        | target has a generated plugin instance based on the requested package and version                                     |
| `specify_plugin_sub_config` | installed plugin has the declared configuration content, when config push is supported by the current execution chain |

This guide does not define machine file paths, reload behavior, health-check commands, or timing guarantees because they are not part of the deploy-policy Swagger contract.

### 7. Repeat or revise the declaration

The concept documentation describes desired final states, not request de-duplication or retry keys. Do not assume a generic idempotency key, retry safety, replacement behavior, or rollback behavior unless your deployment exposes a separate contract for it.

Mode-specific repeat behavior is documented on each mode page only when the existing concept documents support it.

### 8. Handle failures and limits

Treat these as integration boundaries:

- Invalid request shape or unsupported fields are API contract problems. Check the Swagger reference.
- Unsupported `scope` and target combinations are concept problems. Check [scope](../../concepts/deploy_policy/scope.md).
- Unsupported `spec` semantics are desired-state problems. Check [spec](../../concepts/deploy_policy/spec.md).
- A successful create or execute response is an immediate API result, not final machine verification.
- `specify_plugin_sub_config` is config-only and requires an already installed plugin.

## Contract references

- [Scope concept](../../concepts/deploy_policy/scope.md)
- [Spec concept](../../concepts/deploy_policy/spec.md)
- [Swagger contract](../../api/swagger/backend/api/v3/deploy_policy.swagger.json)
- [specify_plugin](specify_plugin.md)
- [specify_plugin_pkg](specify_plugin_pkg.md)
- [specify_plugin_sub_config](specify_plugin_sub_config.md)
