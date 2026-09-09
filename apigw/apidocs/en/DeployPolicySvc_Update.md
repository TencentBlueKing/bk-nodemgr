### Description

- API Version: v3.0.1+
- Required Permission:
- Function: Update a deploy policy.

**Unreleased:** [ensure_absent](../../../docs/concepts/deploy_policy/ensure_absent.md) is saved only, without triggering execution. The current executor ignores it, so execution remains unchanged rather than enforcing absence.

### URL

POST /api/v3/deploy_policy/update

### Request Parameters

| Parameter       | Type   | Required | Description                      |
| --------------- | ------ | -------- | -------------------------------- |
| deploy_policies | array  | Yes      | Deploy policy list to be updated |
| fields          | object | Yes      | Specify which fields to update   |

#### deploy_policies[n]

| Parameter        | Type   | Required | Description                                                                           |
| ---------------- | ------ | -------- | ------------------------------------------------------------------------------------- |
| deploy_policy_id | int64  | Yes      | Deploy policy ID                                                                      |
| dsu_id           | int64  | No       | DSU ID                                                                                |
| meta             | object | No       | Deploy policy metadata                                                                |
| specs            | array  | No       | Deploy specification list                                                             |
| scopes           | array  | No       | Deploy scope list                                                                     |
| operator         | string | No       | Operator                                                                              |
| enabled          | bool   | No       | Whether enabled                                                                       |
| ensure_absent    | bool   | No       | Desired-state direction: false for forward intent (default), true for absence intent. |

#### deploy_policies[n].meta

| Parameter   | Type   | Required | Description               |
| ----------- | ------ | -------- | ------------------------- |
| name        | string | No       | Deploy policy name        |
| description | string | No       | Deploy policy description |

#### deploy_policies[n].specs[n]

Deploy specifications define the desired final state. For structure details, refer to the "Create Deploy Policy" API documentation.

| Parameter | Type   | Required | Description                                                                     |
| --------- | ------ | -------- | ------------------------------------------------------------------------------- |
| type      | string | Yes      | Spec type (enum: specify_plugin, specify_plugin_pkg, specify_plugin_sub_config) |
| param     | object | Yes      | Spec parameters, structure depends on type field                                |

#### deploy_policies[n].scopes[n]

Deploy scopes define the target range where the policy applies. For structure details, refer to the "Create Deploy Policy" API documentation.

| Parameter | Type   | Required | Description                                                                      |
| --------- | ------ | -------- | -------------------------------------------------------------------------------- |
| type      | string | Yes      | Scope type (enum: topo, service_template, set_template, instance, dynamic_group) |
| scope     | object | Yes      | Scope details, structure depends on type field                                   |

#### fields

Specify which fields to update. Only fields set to true will be updated.

| Parameter     | Type | Required | Description                                                              |
| ------------- | ---- | -------- | ------------------------------------------------------------------------ |
| meta          | bool | Yes      | Whether to update metadata (name, description)                           |
| scopes        | bool | Yes      | Whether to update deploy scopes                                          |
| specs         | bool | Yes      | Whether to update deploy specifications                                  |
| enabled       | bool | Yes      | Whether to update enabled status                                         |
| ensure_absent | bool | No       | Update mask for ensure_absent, applied to every policy entry as follows. |

| fields.ensure_absent | Entry ensure_absent  | Result                    |
| -------------------- | -------------------- | ------------------------- |
| Omitted or false     | Any value or omitted | Preserve the stored value |
| true                 | true                 | Save true                 |
| true                 | false or omitted     | Save false                |

### Request Examples

#### Example 1: Update deploy policy name and description

```json
{
  "deploy_policies": [
    {
      "deploy_policy_id": 1001,
      "meta": {
        "name": "Production Monitor Plugin Deploy Policy v2",
        "description": "Unified deployment of bkmonitorbeat 3.60.3066 for production environment"
      }
    }
  ],
  "fields": {
    "meta": true,
    "scopes": false,
    "specs": false,
    "enabled": false
  }
}
```

#### Example 2: Update deploy policy enabled status

```json
{
  "deploy_policies": [
    {
      "deploy_policy_id": 1001,
      "enabled": false
    }
  ],
  "fields": {
    "meta": false,
    "scopes": false,
    "specs": false,
    "enabled": true
  }
}
```

#### Example 3: Update deploy policy specifications

```json
{
  "deploy_policies": [
    {
      "deploy_policy_id": 1001,
      "specs": [
        {
          "type": "specify_plugin",
          "param": {
            "plugin_name": "bkmonitorbeat",
            "version": "3.60.3066"
          }
        }
      ]
    }
  ],
  "fields": {
    "meta": false,
    "scopes": false,
    "specs": true,
    "enabled": false
  }
}
```

#### Example 4: Save absence intent without executing

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

To restore forward intent, set the entry value to `false` while keeping `fields.ensure_absent: true`.

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": null
}
```

### Response Parameters

| Parameter  | Type   | Description                                                                     |
| ---------- | ------ | ------------------------------------------------------------------------------- |
| code       | int32  | Status code, 0 for success                                                      |
| message    | string | Response message                                                                |
| request_id | string | Request ID                                                                      |
| error      | object | Error information (empty on success)                                            |
| data       | null   | No response payload; update does not return a policy ID or execution trigger ID |
