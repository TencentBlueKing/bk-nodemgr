### 描述

- 该接口提供版本：v3.0.1-alpha.116+。
- 该接口所需权限：system_access（系统访问）和 platform_manage（平台管理），两者均须满足。平台管理资源类型为 `platform_resource`，资源 ID 为 `global_settings`（全局配置），与 `scheduler_workflow`（定时任务调度）权限独立。
- 该接口功能描述：以用户身份查询数据库中已保存的全局配置，包含自定义配置项，支持分页、名称过滤和仅计数。
- 本接口为[应用态配置列表](GlobalSettings_ListGlobalSettings.md)的用户态入口，复用同一查询 handler、条件转换及权限检查。
- 调用须具备 API 网关应用授权并通过用户身份认证；用户身份来自网关认证后的可信请求上下文，不通过请求体指定。
- 配置跨租户全局共享，不按当前租户隔离。配置值以原始字符串返回，不脱敏，可能包含自定义敏感信息；调用方应限制结果的访问、记录和转发。

### URL

POST /api/v3/open/globalsettings/list

### 输入参数

| 参数名称                 | 参数类型 | 必选     | 描述                                                                      |
| ------------------------ | -------- | -------- | ------------------------------------------------------------------------- |
| page                     | object   | 条件必选 | 分页配置；`only_count=false` 时须提供有效分页，`only_count=true` 时可省略 |
| only_count               | bool     | 否       | 是否只返回符合条件的总数，不返回配置详情，默认为 `false`                  |
| exact_include_conditions | object   | 否       | 配置名称的精确包含条件                                                    |
| fuzzy_include_conditions | object   | 否       | 配置名称的模糊包含条件                                                    |
| exact_exclude_conditions | object   | 否       | 配置名称的精确排除条件                                                    |
| fuzzy_exclude_conditions | object   | 否       | 配置名称的模糊排除条件                                                    |

#### page

| 参数名称 | 参数类型 | 必选     | 描述                                                                                  |
| -------- | -------- | -------- | ------------------------------------------------------------------------------------- |
| offset   | int32    | 否       | 分页起始位置，从 `0` 开始，默认为 `0`；不可为负数                                     |
| limit    | int32    | 条件必选 | 非计数查询时必填，取值范围为 `1` 到 `500`；计数查询不使用分页，但传入的值仍不可为负数 |

即使 `only_count=true`，传入的 `page.offset` 和 `page.limit` 仍须通过非负数校验。

#### 名称过滤条件

四种条件对象均只支持 `setting_name` 字段，类型为 `string array`，可省略；省略对象或传入空数组时不应用对应过滤条件。

| 条件对象                 | setting_name 的含义                              |
| ------------------------ | ------------------------------------------------ |
| exact_include_conditions | 保留名称与数组中任一值完全相等的配置，区分大小写 |
| exact_exclude_conditions | 排除名称与数组中任一值完全相等的配置，区分大小写 |
| fuzzy_include_conditions | 保留名称包含数组中任一字符串的配置，不区分大小写 |
| fuzzy_exclude_conditions | 排除名称包含数组中任一字符串的配置，不区分大小写 |

多个非空条件会同时应用。模糊匹配按字面字符串包含关系处理：正则表达式特殊字符会被转义，不能用来指定正则表达式。例如 `a.b` 匹配名称中的字面文本 `a.b`，不是 `a` 与 `b` 之间的任意字符。

### 调用示例

查询指定名称的配置。示例假设数据库中已保存该自定义配置项。

```json
{
  "page": { "offset": 0, "limit": 20 },
  "only_count": false,
  "exact_include_conditions": {
    "setting_name": ["custom_integration_options"]
  }
}
```

仅统计名称包含 `integration` 的配置项数量时，可省略分页：

```json
{
  "only_count": true,
  "fuzzy_include_conditions": {
    "setting_name": ["integration"]
  }
}
```

### 响应示例

以下为第一个调用示例的成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 1,
    "items": [
      {
        "setting_name": "custom_integration_options",
        "value": "{\"batch_size\":20}"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                     |
| ---------- | -------- | ------------------------ |
| code       | int32    | 状态码，`0` 表示成功     |
| message    | string   | 请求信息                 |
| request_id | string   | 请求 ID                  |
| error      | object   | 错误信息，成功时通常为空 |
| permission | object   | 权限信息                 |
| data       | object   | 响应数据                 |

#### data

| 参数名称 | 参数类型 | 描述                                                                          |
| -------- | -------- | ----------------------------------------------------------------------------- |
| total    | int64    | 符合条件的配置项总数，不是当前页条数                                          |
| items    | array    | 当前页配置列表；`only_count=true` 时为空数组；非计数查询无结果时可能为 `null` |

#### data.items[n]

| 参数名称     | 参数类型 | 描述                                                                              |
| ------------ | -------- | --------------------------------------------------------------------------------- |
| setting_name | string   | 已保存的配置名称，包含自定义名称                                                  |
| value        | string   | 数据库中的原始字符串值，不脱敏；即使内容是 JSON、数字或布尔值，返回类型仍为字符串 |

### 处理规则说明

- 查询范围为数据库中已保存的全局配置，不限于预定义名称，也不按租户划分数据；请求体不提供租户选择参数。
- 不读取服务启动配置，不为未保存的配置补齐默认值；返回值不表示进程当前实际生效值。
- 本接口只读，不提供配置写入或删除操作。
- 查询单项时，通过 `exact_include_conditions.setting_name` 指定完整名称；名称未匹配时返回 `total=0` 及空列表，不回退到默认值。
