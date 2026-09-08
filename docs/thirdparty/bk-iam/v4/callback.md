# 资源回调

资源回调用于 IAM 查询接入系统的资源实例、名称与拓扑。协议来源见[官方资料](README.md#官方来源)，项目协议差异见[待核实项](integration.md#待核实项)。

## 适用范围

数据级权限且由 IAM 承载申请、管理时，需要资源回调。若申请与管理完全在接入系统内闭环，回调并非必需，但仍建议实现，否则 IAM 的个人权限汇总、审计与排查可能只能显示资源 ID。资源粒度选择见[权限模型](model.md#按权限粒度选择资源)。

## 调用方向与身份

IAM 调用 System 注册的 `callback_url`，通过同一个专用 `POST` 接口查询不同资源类型。接口按 `method` 分发，不与普通业务 API 共用访问逻辑。

| 项目       | 官方摘录要求                                                                          |
| ---------- | ------------------------------------------------------------------------------------- |
| 认证头     | `Authorization: Basic base64(username:password)`                                      |
| 用户名     | `bk_iam`                                                                              |
| 密码       | 系统注册时生成的 token，不是 app_secret                                               |
| token 查询 | `GET /api/v1/open/rbac/model/systems/{system_id}/auth-token/`；读取 `data.auth_token` |
| 请求追踪   | 记录请求头 `X-Request-Id`，响应头返回同一值                                           |
| 成功封装   | HTTP 2xx，返回 `data`                                                                 |
| 失败封装   | HTTP 非 2xx，返回 `error.code` 和 `error.message`                                     |

两段调用使用不同身份：

- 接入系统获取 token：使用配置中的 System ID，复用项目已有的网关应用认证封装。
- IAM 调用 callback：使用 `bk_iam` 与该 System 的 token 做 Basic Auth；System ID 来自服务端配置，不由 callback 请求体选择。

token 查询需设置超时并校验响应包含非空 `data.auth_token`，不能把缺失值默认为空串后参与认证。开发指南建议缓存 token；缓存策略不改变凭据校验要求，获取失败不能放行请求。

## 请求分发与响应

处理顺序为：校验 Basic Auth、解析 JSON、校验 `type` 与 `method`、定位资源 Provider、执行查询、封装响应。`type` 是已注册的 ResourceType ID，`method` 是请求体中的查询操作，不是 HTTP method。

官方 callback 参考实现使用以下状态与错误码：

| 结果                                    | HTTP status | 响应                          |
| --------------------------------------- | ----------- | ----------------------------- |
| 查询成功                                | 200         | `data`                        |
| JSON 无法解析，或缺少 `type` / `method` | 400         | `error.code=INVALID_ARGUMENT` |
| Basic Auth 校验失败                     | 401         | `error.code=UNAUTHENTICATED`  |
| 资源类型未注册，或查询操作不支持        | 404         | `error.code=NOT_FOUND`        |
| 查询执行异常                            | 500         | `error.code=INTERNAL`         |

错误必须体现在 HTTP status，不能统一返回 HTTP 200 再在正文放业务错误码。成功和失败响应都在 header 中回传请求的 `X-Request-Id`。

最小失败响应：

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "resource not found"
  }
}
```

## 实例与路径

| 字段            | 含义                                                    |
| --------------- | ------------------------------------------------------- |
| `id`            | 资源实例 ID，在系统内同一种资源类型下唯一               |
| `type`          | ResourceType ID                                         |
| `display_name`  | 资源实例展示名称                                        |
| `_bk_iam_path_` | 可选的祖先实例拓扑字符串，采用 `/type,id/type,id/` 格式 |

官方路径示例：主机位于业务 A 的集群 B 下，其祖先路径为 `/biz,A/cluster,B/`。路径描述祖先，不把当前实例重复加入末尾。不要与模型 `ancestors` 混用：模型保存资源类型链，路径包含具体祖先实例 ID。

原文要求：支持拓扑层级配置权限时，鉴权请求同时提供 `_bk_iam_path_`。这不仅是页面显示属性；项目是否按目标环境要求传递该属性，需要单独核实。

## list_instance

根据父资源和关键字分页查询实例。

先限定目标资源类型，再按 `parent.type` 与 `parent.id` 筛选直接子实例，最后追加 `display_name` 关键字过滤。父资源和关键字同时出现时取交集，不是满足任一条件即可。`requires` 不参与该方法的字段选择。

| 请求字段         | 必填 | 限制                                                             |
| ---------------- | ---- | ---------------------------------------------------------------- |
| `type`           | 是   | 目标资源类型                                                     |
| `method`         | 是   | `list_instance`                                                  |
| `filter`         | 否   | 可同时包含 `parent`、`keyword`                                   |
| `filter.parent`  | 否   | 直接上级实例，包含 `type` 与 `id`                                |
| `filter.keyword` | 否   | 必须支持 `display_name` 包含关键字搜索；ID 搜索可扩展            |
| `page`           | 是   | 包含 `page` 和 `page_size`；页码从 1 开始，`page_size` 最大 1000 |

从官方示例保留一个父资源过滤请求：

```json
{
  "type": "host",
  "method": "list_instance",
  "filter": { "parent": { "type": "module", "id": "m1" } },
  "page": { "page": 1, "page_size": 20 }
}
```

`data.count` 是应用父资源与关键字过滤后、分页前的实例总数，不是当前页长度。`data.results` 为当前页，每项包含 `id`、`display_name`：

```json
{
  "data": {
    "count": 1,
    "results": [{ "id": "h1", "display_name": "192.168.1.1" }]
  }
}
```

## fetch_instance_info

按 ID 批量获取实例详情，不采用 `list_instance` 的分页协议。

在目标资源类型内按 `filter.ids` 查询，顶层 `requires` 控制属性选择；`filter.parent`、`filter.keyword` 和 `page` 不参与此方法。响应的 `data` 直接是实例列表，不再套 `count/results`。

| 请求字段     | 必填 | 限制                                         |
| ------------ | ---- | -------------------------------------------- |
| `type`       | 是   | 目标资源类型                                 |
| `method`     | 是   | `fetch_instance_info`                        |
| `filter.ids` | 是   | 字符串 ID 数组，最多 1000 个                 |
| `requires`   | 否   | 需要的属性列表；省略或空数组表示查询所有属性 |

必须实现的属性要求，原文：

> 权限中心为了校验用户提交的资源实例名称是否正确, 会调用fetch_instance_info查询实例的display_name属性, 该接口必须实现display_name属性的查询

最小请求与响应：

```json
{
  "type": "host",
  "method": "fetch_instance_info",
  "filter": { "ids": ["h1"] },
  "requires": ["display_name", "_bk_iam_path_"]
}
```

```json
{
  "data": [
    {
      "id": "h1",
      "display_name": "主机1",
      "_bk_iam_path_": "/biz,1/set,1/module,1/"
    }
  ]
}
```

开发指南的请求表与 callback 参考实现均使用顶层 `requires`。响应说明中的 `attrs` 不能作为请求字段别名。默认属性应覆盖 Provider 支持的属性；`display_name` 必须支持，拓扑与实例审批人属性按模型能力提供，不能把固定字段列表等同于任意资源的全部属性。

实例审批人是可选能力；采用时需确定负责人来源，并关联角色的实例审批流，见[项目接入](integration.md#待核实项)。

## 接入防护

以下是实现与验证要求，不是对缺失协议的默认解释：

- 校验 JSON 顶层为对象，`type`、`method` 为非空字符串，并校验 `filter`、`page`、`requires` 的类型，避免参数错误退化成内部异常。
- `list_instance` 校验页码、页大小与父资源字段；不依赖示例中的缺省页大小，查询应有稳定排序。
- `fetch_instance_info` 不得因 `filter.ids` 缺失或为空而移除 ID 限制、查询全部资源。空数组、缺失实例、非法父资源及越界页的具体响应仍需核实目标协议。
- 区分凭据不匹配与 token 服务不可用，两者都不能放行；依赖故障不能仅作为用户凭据错误处理。错误响应不直接暴露内部异常文本或凭据。
