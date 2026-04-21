### Description

- API Version: v3.0.0+.
- Required Permission: none.
- Function: Verify whether the current user has permission for the requested action-resource pairs.

### URL

POST /api/v3/auth/verify

### Input Parameters

| Parameter | Type  | Required | Description                                             |
| --------- | ----- | -------- | ------------------------------------------------------- |
| items     | array | Yes      | List of action-resource pairs to verify permissions for |

#### items[n]

| Parameter | Type   | Required | Description                                                                                             |
| --------- | ------ | -------- | ------------------------------------------------------------------------------------------------------- |
| action    | string | Yes      | IAM action identifier (current built-in action set is listed below; API layer validates non-empty only) |
| resources | array  | No       | List of resource instances to verify. Empty means an action-level permission check only                 |

##### action enum values

`agent_view`, `agent_operate`, `agent_history_view`, `biz_access`, `proxy_view`, `proxy_operate`, `proxy_history_view`, `plugin_view`, `plugin_operate`,
`plugin_history_view`, `config_policy_view`, `config_policy_manage`, `config_policy_history_view`, `deploy_policy_view`, `deploy_policy_manage`,
`deploy_policy_history_view`, `networkarea_view`, `networkarea_create`, `networkarea_edit`, `networkarea_delete`, `networkarea_history_view`,
`networkunit_view`, `networkunit_create`, `networkunit_edit`, `networkunit_delete`, `networkunit_use_for_agent`, `networkunit_use_for_proxy`,
`networkunit_history_view`, `package_type_upload`, `package_view`, `package_manage`, `package_history_view`

#### items[n].resources[m]

| Parameter | Type   | Required | Description                                                                                                                |
| --------- | ------ | -------- | -------------------------------------------------------------------------------------------------------------------------- |
| system_id | string | Yes      | IAM system identifier that owns this resource (e.g., `bk_cmdb`, `bk_nodemgr`)                                              |
| type      | string | Yes      | Resource type identifier (current built-in resource types: `biz`, `networkarea`, `networkunit`, `package_type`, `package`) |
| id        | string | Yes      | Resource instance ID                                                                                                       |

### Request Example

```json
{
  "items": [
    {
      "action": "agent_view",
      "resources": [
        {
          "system_id": "bk_cmdb",
          "type": "biz",
          "id": "2"
        }
      ]
    },
    {
      "action": "networkarea_create",
      "resources": []
    }
  ]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "results": [
      {
        "action": "agent_view",
        "authorized": true
      },
      {
        "action": "networkarea_create",
        "authorized": true
      }
    ]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                                     |
| ---------- | ------ | --------------------------------------------------------------- |
| code       | int32  | Status code, 0 indicates success                                |
| message    | string | Response message                                                |
| request_id | string | Request ID                                                      |
| error      | object | Error information (empty on success)                            |
| permission | object | Permission application information (returned when unauthorized) |
| data       | object | Response data                                                   |

#### data

| Parameter | Type  | Description                             |
| --------- | ----- | --------------------------------------- |
| results   | array | List of permission verification results |

#### data.results[n]

| Parameter  | Type   | Description                                                                            |
| ---------- | ------ | -------------------------------------------------------------------------------------- |
| action     | string | The IAM action identifier that was verified (value comes from input `items[n].action`) |
| authorized | bool   | Whether the current user has permission for this action                                |
