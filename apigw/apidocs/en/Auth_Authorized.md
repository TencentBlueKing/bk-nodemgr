### Description

- API Version: v3.0.0+.
- Required Permission: none.
- Function: Query the authorized resource scope for the requested action-resource type pairs.

### URL

POST /api/v3/auth/authorized

### Input Parameters

| Parameter | Type  | Required | Description                                                      |
|-----------|-------|----------|------------------------------------------------------------------|
| items     | array | Yes      | List of action-resource type pairs to query authorized scope for |

#### items[n]

| Parameter     | Type   | Required | Description                                                                                                                                                                       |
|---------------|--------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| action        | string | Yes      | IAM action identifier (current built-in action set is listed below; API layer validates non-empty only)                                                                           |
| resource_type | string | Yes      | Resource type to query authorized scope for (current built-in resource types: `biz`, `networkarea`, `networkunit`, `package_type`, `package`; API layer validates non-empty only) |

##### action enum values

`agent_view`, `agent_operate`, `agent_history_view`, `proxy_view`, `proxy_operate`, `proxy_history_view`, `plugin_view`, `plugin_operate`,
`plugin_history_view`, `config_policy_view`, `config_policy_manage`, `config_policy_history_view`, `deploy_policy_view`, `deploy_policy_manage`,
`deploy_policy_history_view`, `networkarea_view`, `networkarea_create`, `networkarea_edit`, `networkarea_delete`, `networkarea_history_view`,
`networkunit_view`, `networkunit_create`, `networkunit_edit`, `networkunit_delete`, `networkunit_use_for_agent`, `networkunit_use_for_proxy`,
`networkunit_history_view`, `package_type_upload`, `package_view`, `package_manage`, `package_history_view`

### Request Example

```json
{
  "items": [
    {
      "action": "agent_view",
      "resource_type": "biz"
    },
    {
      "action": "networkarea_create",
      "resource_type": "networkarea"
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
        "resource_type": "biz",
        "is_any": false,
        "resources": [
          {
            "system_id": "bk_cmdb",
            "type": "biz",
            "id": "2"
          },
          {
            "system_id": "bk_cmdb",
            "type": "biz",
            "id": "5"
          }
        ]
      },
      {
        "action": "networkarea_create",
        "resource_type": "networkarea",
        "is_any": true,
        "resources": []
      }
    ]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                                     |
|------------|--------|-----------------------------------------------------------------|
| code       | int32  | Status code, 0 indicates success                                |
| message    | string | Response message                                                |
| request_id | string | Request ID                                                      |
| error      | object | Error information (empty on success)                            |
| permission | object | Permission application information (returned when unauthorized) |
| data       | object | Response data                                                   |

#### data

| Parameter | Type  | Description                            |
|-----------|-------|----------------------------------------|
| results   | array | List of authorized scope query results |

#### data.results[n]

| Parameter     | Type   | Description                                                                                                                                                             |
|---------------|--------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| action        | string | The IAM action identifier that was queried (value comes from input `items[n].action`)                                                                                   |
| resource_type | string | The resource type that was queried (value comes from input `items[n].resource_type`)                                                                                    |
| is_any        | bool   | Whether the user has unrestricted access to all resources of this type. When true, the user can access all resources of this type, and the resources list will be empty |
| resources     | array  | List of authorized resource instances. Empty array when is_any is true                                                                                                  |

#### data.results[n].resources[m]

| Parameter | Type   | Description                                                                                                                |
|-----------|--------|----------------------------------------------------------------------------------------------------------------------------|
| system_id | string | IAM system identifier that owns this resource (e.g., "bk_cmdb", "bk_nodemgr")                                              |
| type      | string | Resource type identifier (current built-in resource types: `biz`, `networkarea`, `networkunit`, `package_type`, `package`) |
| id        | string | Resource instance ID                                                                                                       |
