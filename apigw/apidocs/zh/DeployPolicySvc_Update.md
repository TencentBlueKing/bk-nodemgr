### 描述

- 该接口提供版本：v3.0.1+
- 该接口所需权限：
- 该接口功能描述：更新部署策略。

**未发布（Unreleased）：** [ensure_absent](../../../docs/concepts/deploy_policy/ensure_absent.md) 仅保存，不触发执行。当前执行器忽略该字段，执行逻辑不变，不保证目标不存在。

### URL

POST /api/v3/deploy_policy/update

### 输入参数

| 参数名称        | 参数类型 | 必选 | 描述                 |
| --------------- | -------- | ---- | -------------------- |
| deploy_policies | array    | 是   | 待更新的部署策略列表 |
| fields          | object   | 是   | 指定需要更新的字段   |

#### deploy_policies[n]

| 参数名称         | 参数类型 | 必选 | 描述                                                        |
| ---------------- | -------- | ---- | ----------------------------------------------------------- |
| deploy_policy_id | int64    | 是   | 部署策略ID                                                  |
| dsu_id           | int64    | 否   | DSU ID                                                      |
| meta             | object   | 否   | 部署策略元信息                                              |
| specs            | array    | 否   | 部署规范列表                                                |
| scopes           | array    | 否   | 部署范围列表                                                |
| operator         | string   | 否   | 操作人                                                      |
| enabled          | bool     | 否   | 是否启用                                                    |
| ensure_absent    | bool     | 否   | 期望状态方向：false 为正向意图（默认），true 为不存在意图。 |

#### deploy_policies[n].meta

| 参数名称    | 参数类型 | 必选 | 描述         |
| ----------- | -------- | ---- | ------------ |
| name        | string   | 否   | 部署策略名称 |
| description | string   | 否   | 部署策略描述 |

#### deploy_policies[n].specs[n]

部署规范定义了期望的最终状态。结构说明参见"创建部署策略"接口文档。

| 参数名称 | 参数类型 | 必选 | 描述                                                                              |
| -------- | -------- | ---- | --------------------------------------------------------------------------------- |
| type     | string   | 是   | 规范类型（枚举值：specify_plugin、specify_plugin_pkg、specify_plugin_sub_config） |
| param    | object   | 是   | 规范参数，结构根据type字段而定                                                    |

#### deploy_policies[n].scopes[n]

部署范围定义了策略应用的目标范围。结构说明参见"创建部署策略"接口文档。

| 参数名称 | 参数类型 | 必选 | 描述                                                                              |
| -------- | -------- | ---- | --------------------------------------------------------------------------------- |
| type     | string   | 是   | 范围类型（枚举值：topo、service_template、set_template、instance、dynamic_group） |
| scope    | object   | 是   | 范围详情，结构根据type字段而定                                                    |

#### fields

指定需要更新的字段。只有设置为true的字段才会被更新。

| 参数名称      | 参数类型 | 必选 | 描述                                               |
| ------------- | -------- | ---- | -------------------------------------------------- |
| meta          | bool     | 是   | 是否更新元信息（name、description）                |
| scopes        | bool     | 是   | 是否更新部署范围                                   |
| specs         | bool     | 是   | 是否更新部署规范                                   |
| enabled       | bool     | 是   | 是否更新启用状态                                   |
| ensure_absent | bool     | 否   | ensure_absent 更新掩码，对所有策略条目按下表生效。 |

| fields.ensure_absent | 条目 ensure_absent | 结果       |
| -------------------- | ------------------ | ---------- |
| 省略或 false         | 任意值或省略       | 保留原值   |
| true                 | true               | 保存 true  |
| true                 | false 或省略       | 保存 false |

### 调用示例

#### 示例1：更新部署策略的名称和描述

```json
{
  "deploy_policies": [
    {
      "deploy_policy_id": 1001,
      "meta": {
        "name": "生产环境监控插件部署策略v2",
        "description": "用于生产环境统一部署 bkmonitorbeat 3.60.3066"
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

#### 示例2：更新部署策略的启用状态

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

#### 示例3：更新部署策略的规范

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

#### 示例4：仅保存不存在意图，不执行

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

恢复正向意图时，将条目值设为 `false`，同时保持 `fields.ensure_absent: true`。

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": null
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                                            |
| ---------- | -------- | ----------------------------------------------- |
| code       | int32    | 状态码，0表示成功                               |
| message    | string   | 请求信息                                        |
| request_id | string   | 请求ID                                          |
| error      | object   | 错误信息（成功时为空）                          |
| data       | null     | 无响应数据；更新不返回策略 ID 或执行 trigger ID |
