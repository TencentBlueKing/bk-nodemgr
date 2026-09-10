### 描述

- 该接口提供版本：v3.0.1-alpha.83+。
- 该接口所需权限：无。
- 该接口功能描述：供 BlueKing IAM V4 回调查询节点管理资源实例列表和详情，按资源类型和回调方法分发到对应资源提供方。

### URL

POST /api/v3/iam/v4/resource

### 输入参数

#### 请求头

| 参数名称       | 参数类型 | 必选     | 描述                                                                               |
| -------------- | -------- | -------- | ---------------------------------------------------------------------------------- |
| Authorization  | string   | 是       | Basic Auth 凭据，用户名固定为 `bk_iam`，密码为当前租户对应的 IAM V4 system token。 |
| Content-Type   | string   | 是       | `application/json`。                                                               |
| X-Bk-Tenant-Id | string   | 条件必选 | 多租户模式下必需且必须是有效租户 ID；单租户模式使用系统固定租户。                  |
| X-Request-Id   | string   | 否       | 请求追踪标识，回调在成功和失败响应头中回传收到的值。                               |

#### 请求体

| 参数名称 | 参数类型 | 必选     | 描述                                                                                                                                            |
| -------- | -------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| type     | string   | 是       | 非空资源类型。当前已注册 `biz`、`networkarea`、`networkunit`、`package_type`、`package`。                                                       |
| method   | string   | 是       | 仅支持 `list_instance`、`fetch_instance_info`，必须传字符串，不支持数字方法标识。                                                               |
| filter   | object   | 条件必选 | `list_instance` 可省略；`fetch_instance_info` 必需且包含 `ids`。提供时不能为 `null`。                                                           |
| page     | object   | 条件必选 | 仅 `list_instance` 必需；`fetch_instance_info` 不使用分页。提供时不能为 `null`。                                                                |
| requires | array    | 否       | 字符串数组，位于请求体顶层，仅用于 `fetch_instance_info` 属性选择。省略或 `[]` 返回全部受支持属性，未知属性忽略，`id` 始终返回。不能为 `null`。 |

#### filter：list_instance

| 参数名称 | 参数类型 | 必选 | 描述                                                                                                                                                                          |
| -------- | -------- | ---- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| parent   | object   | 否   | 直接父资源，包含非空字符串 `type` 和 `id`。`networkunit` 支持 `networkarea` 父资源，`package` 支持 `package_type` 父资源；`biz`、`networkarea`、`package_type` 不接受父资源。 |
| keyword  | string   | 否   | 资源展示名称过滤关键字。与 `parent` 同时提供时取交集。                                                                                                                        |

省略 `parent` 时查询当前租户的候选资源，不表示授予资源权限。

#### filter：fetch_instance_info

| 参数名称 | 参数类型 | 必选 | 描述                                                                                                                    |
| -------- | -------- | ---- | ----------------------------------------------------------------------------------------------------------------------- |
| ids      | array    | 是   | 非空字符串组成的数组，最多 1000 项。数组本身可为 `[]`，此时返回 `data: []`；不能为 `null`。不存在的实例不出现在结果中。 |

#### page：list_instance

| 参数名称  | 参数类型 | 必选 | 描述                                                    |
| --------- | -------- | ---- | ------------------------------------------------------- |
| page      | int64    | 是   | 页码，从 1 开始，无默认值；计算出的分页偏移量不能溢出。 |
| page_size | int64    | 是   | 每页条数，范围为 1 到 1000，无默认值。                  |

### 调用示例

以下为请求体示例，调用时还需携带上述请求头。资源 ID 和名称仅用于说明响应结构。

#### list_instance

```json
{
  "type": "networkarea",
  "method": "list_instance",
  "filter": {
    "keyword": "default"
  },
  "page": {
    "page": 1,
    "page_size": 100
  }
}
```

#### fetch_instance_info

```json
{
  "type": "networkarea",
  "method": "fetch_instance_info",
  "filter": {
    "ids": ["2"]
  },
  "requires": ["display_name"]
}
```

### 响应示例

#### list_instance：HTTP 200

```json
{
  "data": {
    "count": 1,
    "results": [
      {
        "id": "2",
        "display_name": "default-network-area"
      }
    ]
  }
}
```

#### fetch_instance_info：HTTP 200

```json
{
  "data": [
    {
      "id": "2",
      "display_name": "default-network-area"
    }
  ]
}
```

#### 参数错误：HTTP 400

```json
{
  "error": {
    "code": "INVALID_ARGUMENT",
    "message": "invalid callback arguments"
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型     | 描述                                                                                                 |
| -------- | ------------ | ---------------------------------------------------------------------------------------------------- |
| data     | object/array | 成功时返回。`list_instance` 为包含 `count`、`results` 的对象；`fetch_instance_info` 为实例详情数组。 |
| error    | object       | 失败时返回，包含字符串 `code` 和 `message`，不返回成功数据。                                         |

#### data：list_instance

| 参数名称                | 参数类型 | 描述                                                  |
| ----------------------- | -------- | ----------------------------------------------------- |
| count                   | int64    | 过滤后的匹配资源总数，不是当前页条数。                |
| results                 | array    | 当前页的资源实例数组。无匹配或超出分页范围时为 `[]`。 |
| results[n].id           | string   | 资源实例唯一标识。                                    |
| results[n].display_name | string   | 资源实例展示名称。                                    |

#### data[n]：fetch_instance_info

| 参数名称      | 参数类型 | 描述                                                                                                         |
| ------------- | -------- | ------------------------------------------------------------------------------------------------------------ |
| id            | string   | 资源实例唯一标识，始终返回。                                                                                 |
| display_name  | string   | 资源展示名称，受顶层 `requires` 控制。                                                                       |
| _bk_iam_path_ | string   | 有父资源的实例可返回祖先路径，如网络单元的 `/networkarea,2/`，受 `requires` 控制。回调返回字符串，而非数组。 |

受支持的属性直接平铺在实例对象中，没有嵌套的 `attributes` 对象；不存在或未选择的属性不返回。

#### error

| HTTP 状态码 | error.code       | 描述                                                                        |
| ----------- | ---------------- | --------------------------------------------------------------------------- |
| 400         | INVALID_ARGUMENT | JSON、参数、分页或多租户模式下的租户 ID 无效。                              |
| 401         | UNAUTHENTICATED  | Basic Auth 凭据缺失或无效，响应包含 `WWW-Authenticate: Basic realm="IAM"`。 |
| 404         | NOT_FOUND        | 资源类型未注册或回调方法不支持。                                            |
| 500         | INTERNAL         | token 查询、存储查询、序列化等内部错误。                                    |

### 说明

- 此接口是 IAM V4 专用回调入口，不面向普通业务用户；仅在服务配置了 IAM V4 handler 时注册。
- APIGW 的用户认证、应用认证和资源权限校验均关闭，但后端仍执行专用 Basic Auth 校验。
- V4 使用 HTTP 状态码及 `error.code`、`error.message` 表示失败，不使用 V3 的顶层 `code`、`message`；分页也不使用 V3 的 `offset`、`limit`。
