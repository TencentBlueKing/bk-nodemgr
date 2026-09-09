## ensure_absent

`ensure_absent` is a policy-level boolean that records the desired-state direction for the policy's specs. It defaults to `false`, preserving forward deployment. `enabled` controls whether a policy participates; it does not select forward or reverse behavior. Disabling a policy does not request removal.

### Current Behavior

**Unreleased:** Update saves `ensure_absent`; Backend list returns it. The current executor ignores the field, so executing a policy with `ensure_absent: true` still runs existing logic, not safe absence enforcement.

Create is unchanged: new policies and existing policies without the stored field default to `false`. No new endpoints, filters, frontend controls, or Agent/Proxy absence support are added.

### Update Contract

`POST /api/v3/deploy_policy/update` saves without triggering execution. `fields.ensure_absent` is an optional boolean mask applying to every `deploy_policies` entry; `deploy_policies[n].ensure_absent` is the optional boolean value.

| `fields.ensure_absent` | Entry `ensure_absent` | Saved value                 |
| ---------------------- | --------------------- | --------------------------- |
| Omitted or `false`     | Any value or omitted  | Preserve the existing value |
| `true`                 | `true`                | `true`                      |
| `true`                 | `false`               | `false`                     |
| `true`                 | Omitted               | `false`                     |

Save absence intent:

```json
{
  "deploy_policies": [
    {
      "deploy_policy_id": 1001,
      "ensure_absent": true
    }
  ],
  "fields": {
    "ensure_absent": true
  }
}
```

To restore forward intent, keep the mask `true` and set the entry value to `false`. Update returns `data: null`, without a policy ID or trigger ID; Backend list returns `data.items[n].ensure_absent`.

### Agreed Future Semantics

These rules are not implemented. For participating policies, `false` selects each spec's forward state; `true` selects absence of its identified plugin or configuration outputs.

Conflict priority remains oldest-created-first. Reverse policies receive no special priority over forward policies. Absence matching does not filter by version: a matching installed plugin must be removed even when its version differs from the spec's version.

#### Target Selection And Ownership

| Spec category                                                                                 | Future absence target                                                                                                                                                                                           |
| --------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `specify_plugin`                                                                              | Plugins matching `plugin_name` within the **current scope**, regardless of installation origin or ownership. Installed but stopped plugins are included. This does not include hosts outside the current scope. |
| Ownership-backed plugin specs, such as `specify_plugin_pkg` and `project_plugin_pkg_to_hosts` | Outputs attributable to the respective spec, using the plugin group equal to the policy ID as ownership evidence. Include that spec's outputs left at old scope targets or old placement hosts.                 |
| Ownership-backed configuration specs                                                          | Configuration outputs attributable to the respective spec, using the configuration set `deploy_policy_<ID>` as ownership evidence. Include that spec's outputs left at old scope targets or old placements.     |

The plugin `group` is the policy ID as a string; the configuration set is `deploy_policy_<ID>` (for example, `deploy_policy_1001`). These markers prove policy ownership, not complete spec identity. Cleanup must establish attribution to the respective spec, not indiscriminately delete policy-wide historical outputs, other specs' outputs, or outputs of removed specs.

#### Spec Identity

A spec's `type` is immutable under the agreed future rule, not currently enforced. Versions, configuration content, and placement remain mutable; other identity fields and attribution rules remain to be determined per spec.

#### Completion Rules

| Observed state or result                                                  | Future required behavior                                                                                |
| ------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| Matching plugin is installed, running or stopped                          | Uninstall it; stopping the process is not absence                                                       |
| Matching configuration file exists                                        | Actually delete the file and synchronize its management record; deleting only the record is not absence |
| Target is confirmed already absent                                        | No-op; repeated reconciliation is idempotent                                                            |
| Host is unreachable, deletion/uninstall fails, or actual state is unknown | Do not report successful absence                                                                        |

Stopped or missing Running processes do not prove file absence; record-only forward cleanup does not satisfy this contract.

### References

- [Deploy specs](spec.md)
- [Update API](../../../apigw/apidocs/en/DeployPolicySvc_Update.md)
- [Backend list API](../../../apigw/apidocs/en/DeployPolicySvc_List.md)
