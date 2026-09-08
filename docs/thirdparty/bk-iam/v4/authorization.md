# 鉴权与授权范围

鉴权判断用户能否执行操作，授权范围查询用于收敛数据查询范围，申请 URL 用于引导用户补充权限。协议来源见[官方资料](README.md#官方来源)。

## 调用身份

IAM API 调用复用项目已有的网关应用认证封装；`subject` 指定被鉴权的用户，与调用 API 的应用身份不同。

以下路径保留官方摘录的 `/api` 前缀，实际 endpoint 拼接见 [项目接入](integration.md#配置与路由)。

## 直接鉴权

`POST /api/v1/open/rbac/authorization/systems/{system_id}/auth/`

| 字段                  | 位置 | 必填       | 含义                                                 |
| --------------------- | ---- | ---------- | ---------------------------------------------------- |
| `system_id`           | path | 是         | 注册系统 ID                                          |
| `subject.type`        | body | 是         | 原文只支持 `user`                                    |
| `subject.id`          | body | 是         | 被鉴权用户 ID                                        |
| `action_id`           | body | 是         | 操作 ID                                              |
| `resource`            | body | 按模型     | Action 关联资源类型时必填，否则无需填写              |
| `resource.id`         | body | 有资源时是 | 资源实例 ID                                          |
| `resource.attributes` | body | 否         | 资源属性；拓扑授权时原文要求同时提供 `_bk_iam_path_` |

官方最小示例，保留原资源与操作名称，不代表 bk-nodemgr 模型：

```json
{
  "subject": { "type": "user", "id": "jiananzhang" },
  "action_id": "execute_job",
  "resource": {
    "id": "ping",
    "attributes": { "_bk_iam_path_": "/biz,1/set,2/" }
  }
}
```

成功调用返回：

```json
{ "data": { "allowed": true } }
```

必须区分三个结果：`allowed=true` 为允许，`allowed=false` 为拒绝；HTTP/协议错误为鉴权调用失败，不能当成允许。原文的错误示例：

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": " action(execute_job) not found"
  }
}
```

## 批量鉴权与授权范围

批量鉴权与授权范围查询需确认以下 API 契约：

| 需要补齐的官方能力         | 要核实的契约                                                 |
| -------------------------- | ------------------------------------------------------------ |
| 一个 Action 对多个资源鉴权 | API 路径、资源属性、批量上限、逐项结果与缺项语义             |
| 多个 Action 对一个资源鉴权 | API 路径、操作上限、资源类型限制、逐项结果                   |
| 查询已授权资源范围         | API 路径、请求与响应字段、全量标记、父级授权、分页或数量限制 |

原文仅把授权关系 API 概括为 `/api/v1/open/rbac/authorization/systems/{system_id}/relations/*`，未给出完整协议。不能由该通配路径推导实际 endpoint；尤其应核实 `relations` 与项目使用的 `relation` 的差异。

项目使用的接口与范围解释见[运行时映射](integration.md#运行时映射)。

## 无权限申请 URL

`POST /api/v1/open/application/permission-apply-urls/`

接入系统提交缺失的 Action 和资源信息，IAM 返回申请 URL。该 API 不执行鉴权，也不直接授予权限。

| 字段                      | 必填       | 含义                                                    |
| ------------------------- | ---------- | ------------------------------------------------------- |
| `system_id`               | 是         | 注册系统 ID                                             |
| `permissions`             | 是         | 权限对象数组                                            |
| `permissions[].action_id` | 是         | 缺失的操作 ID，不是角色 ID                              |
| `permissions[].resources` | 否         | 资源拓扑列表                                            |
| `resources[].id` / `type` | 有资源时是 | 实例 ID 与资源类型                                      |
| `resources[].ancestors`   | 否         | 从根到当前实例直接上级的祖先列表，每项包含 `id`、`type` |

官方请求示例的最小部分：

```json
{
  "system_id": "system_id",
  "permissions": [
    {
      "action_id": "project_view",
      "resources": [{ "id": "1", "type": "project" }]
    }
  ]
}
```

成功返回 `data.url`。使用返回 URL，不根据官方示例中的 host、cache_id 或查询参数格式自行拼接。获取 URL 失败也不能改变原本的权限拒绝结果。
