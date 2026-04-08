### 描述

- 该接口提供版本：v3.0.0+。
- 该接口所需权限：无。
- 该接口功能描述：查询用户对指定操作-资源类型对的授权资源范围。

### URL

POST /api/v3/auth/authorized

### 输入参数

| 参数名称  | 参数类型  | 必选 | 描述                 |
|-------|-------|----|--------------------|
| items | array | 是  | 要查询授权范围的操作-资源类型对列表 |

#### items[n]

| 参数名称          | 参数类型   | 必选 | 描述                                                                                           |
|---------------|--------|----|----------------------------------------------------------------------------------------------|
| action        | string | 是  | IAM 操作标识符（当前内置动作见下方“action 枚举值”；API 层仅校验非空）                                                  |
| resource_type | string | 是  | 要查询授权范围的资源类型（当前内置资源类型：`biz`、`networkarea`、`networkunit`、`package_type`、`package`；API 层仅校验非空） |

##### action 枚举值

`agent_view`、`agent_operate`、`agent_history_view`、`proxy_view`、`proxy_operate`、`proxy_history_view`、`plugin_view`、`plugin_operate`、
`plugin_history_view`、`config_policy_view`、`config_policy_manage`、`config_policy_history_view`、`deploy_policy_view`、`deploy_policy_manage`、
`deploy_policy_history_view`、`networkarea_view`、`networkarea_create`、`networkarea_edit`、`networkarea_delete`、`networkarea_history_view`、
`networkunit_view`、`networkunit_create`、`networkunit_edit`、`networkunit_delete`、`networkunit_use_for_agent`、`networkunit_use_for_proxy`、
`networkunit_history_view`、`package_type_upload`、`package_view`、`package_manage`、`package_history_view`

### 调用示例

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

### 响应示例

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

### 响应参数说明

| 参数名称       | 参数类型   | 描述             |
|------------|--------|----------------|
| code       | int32  | 状态码，0表示成功      |
| message    | string | 请求信息           |
| request_id | string | 请求ID           |
| error      | object | 错误信息（成功时为空）    |
| permission | object | 权限申请信息（无权限时返回） |
| data       | object | 响应数据           |

#### data

| 参数名称    | 参数类型  | 描述         |
|---------|-------|------------|
| results | array | 授权范围查询结果列表 |

#### data.results[n]

| 参数名称          | 参数类型   | 描述                                                               |
|---------------|--------|------------------------------------------------------------------|
| action        | string | 查询的 IAM 操作标识符（取值来源于输入参数 `items[n].action`）                       |
| resource_type | string | 查询的资源类型（取值来源于输入参数 `items[n].resource_type`）                      |
| is_any        | bool   | 是否对该资源类型的所有资源有无限制访问权限。为 true 时表示用户可以访问该类型的所有资源，此时 resources 列表为空 |
| resources     | array  | 授权的资源实例列表。当 is_any 为 true 时为空数组                                  |

#### data.results[n].resources[m]

| 参数名称      | 参数类型   | 描述                                                                           |
|-----------|--------|------------------------------------------------------------------------------|
| system_id | string | 拥有该资源的 IAM 系统标识符（如 "bk_cmdb"、"bk_nodemgr"）                                   |
| type      | string | 资源类型标识符（当前内置资源类型：`biz`、`networkarea`、`networkunit`、`package_type`、`package`） |
| id        | string | 资源实例 ID                                                                      |
