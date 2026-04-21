### 描述

- 该接口提供版本：v3.0.0+。
- 该接口所需权限：无。
- 该接口功能描述：校验当前用户是否对指定的操作-资源对拥有权限。

### URL

POST /api/v3/auth/verify

### 输入参数

| 参数名称  | 参数类型  | 必选 | 描述             |
|-------|-------|----|----------------|
| items | array | 是  | 要校验权限的操作-资源对列表 |

#### items[n]

| 参数名称      | 参数类型   | 必选 | 描述                                          |
|-----------|--------|----|---------------------------------------------|
| action    | string | 是  | IAM 操作标识符（当前内置动作见下方“action 枚举值”；API 层仅校验非空） |
| resources | array  | 否  | 要校验的资源实例列表。为空时表示仅校验 action 级权限              |

##### action 枚举值

`biz_access`、`agent_view`、`agent_operate`、`agent_history_view`、`proxy_view`、`proxy_operate`、`proxy_history_view`、`plugin_view`、`plugin_operate`、
`plugin_history_view`、`config_policy_view`、`config_policy_manage`、`config_policy_history_view`、`deploy_policy_view`、`deploy_policy_manage`、
`deploy_policy_history_view`、`networkarea_view`、`networkarea_create`、`networkarea_edit`、`networkarea_delete`、`networkarea_history_view`、
`networkunit_view`、`networkunit_create`、`networkunit_edit`、`networkunit_delete`、`networkunit_use_for_agent`、`networkunit_use_for_proxy`、
`networkunit_history_view`、`package_type_upload`、`package_view`、`package_manage`、`package_history_view`

#### items[n].resources[m]

| 参数名称      | 参数类型   | 必选 | 描述                                                                           |
|-----------|--------|----|------------------------------------------------------------------------------|
| system_id | string | 是  | 拥有该资源的 IAM 系统标识符（如 `bk_cmdb`、`bk_nodemgr`）                                   |
| type      | string | 是  | 资源类型标识符（当前内置资源类型：`biz`、`networkarea`、`networkunit`、`package_type`、`package`） |
| id        | string | 是  | 资源实例 ID                                                                      |

### 调用示例

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

| 参数名称    | 参数类型  | 描述       |
|---------|-------|----------|
| results | array | 权限校验结果列表 |

#### data.results[n]

| 参数名称       | 参数类型   | 描述                                          |
|------------|--------|---------------------------------------------|
| action     | string | 被校验的 IAM 操作标识符（取值来源于输入参数 `items[n].action`） |
| authorized | bool   | 当前用户是否拥有该操作对应的权限                            |
