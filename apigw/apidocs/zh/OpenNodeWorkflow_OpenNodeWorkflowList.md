### 描述

- 该接口提供版本：v3.0.1-alpha.114+。
- 该接口所需权限：system_access（系统访问）、agent_history_view（查看 Agent 历史）、proxy_history_view（查看 Proxy 历史）；历史查看权限按下述规则确定。
- 该接口功能描述：以用户身份查询节点任务流及执行状态，支持分页、精确条件和操作时间范围过滤。
- 用户身份由 API 网关认证并传递，请求体不需要用户名或租户参数。

### URL

POST /api/v3/open/node/workflow/list

### 输入参数

| 参数名称                 | 参数类型 | 必选 | 描述                                |
| ------------------------ | -------- | ---- | ----------------------------------- |
| page                     | object   | 否   | 分页配置；`only_count=false` 时必填 |
| only_count               | bool     | 否   | 是否仅返回总数，默认为 `false`      |
| exact_include_conditions | object   | 否   | 精确匹配包含条件                    |
| fuzzy_include_conditions | object   | 否   | 当前没有可用字段，省略或传 `{}`     |
| operate_time_range       | object   | 否   | 操作时间范围，Unix 秒时间戳         |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                                 |
| -------- | -------- | ---- | ------------------------------------ |
| offset   | int32    | 否   | 起始偏移量，默认为 `0`，不得小于 `0` |
| limit    | int32    | 是   | 每页条数，范围 `1`~`500`             |

#### exact_include_conditions

| 参数名称           | 参数类型     | 必选 | 描述                                                               |
| ------------------ | ------------ | ---- | ------------------------------------------------------------------ |
| workflow_id        | string array | 否   | 任务流 ID 列表；已知 ID 时在此筛选，不是顶层参数                   |
| type               | string array | 否   | 任务流类型，例如 `install_agent`、`upgrade_agent`、`restart_proxy` |
| bk_biz_id          | int64 array  | 否   | 业务 ID 列表                                                       |
| status             | string array | 否   | 任务流状态：`running`、`success`、`failed`、`partial_failed`       |
| operator           | string array | 否   | 历史操作人列表；不是本次调用的身份认证参数                         |
| bk_host_innerip    | string array | 否   | 主机内网 IPv4 列表                                                 |
| bk_host_innerip_v6 | string array | 否   | 主机内网 IPv6 列表                                                 |
| node_role          | string array | 否   | 节点角色：`blank`、`agent`、`proxy`                                |
| deploy_policy_id   | int64 array  | 否   | 部署策略 ID 列表                                                   |

#### operate_time_range

| 参数名称            | 参数类型 | 必选 | 描述                                                   |
| ------------------- | -------- | ---- | ------------------------------------------------------ |
| start_timestamp_sec | int64    | 否   | 起始 Unix 秒时间戳，非负，省略时为 `0`                 |
| end_timestamp_sec   | int64    | 否   | 结束 Unix 秒时间戳，非负，省略时为 `0`；不得小于起始值 |

不限制操作时间时省略整个 `operate_time_range` 对象。

### 权限说明

- 始终需要 `system_access`。
- 传入 `exact_include_conditions.workflow_id` 时，按每个匹配任务流实际记录的节点角色和业务校验历史查看权限，再返回分页结果或总数；任一匹配任务流鉴权失败，请求失败。
- 未按任务流 ID 筛选时，查询范围按调用用户的业务授权收敛。`node_role` 为 `blank` 或 `agent` 时需要 Agent 历史权限，为 `proxy` 时需要 Proxy 历史权限；未指定角色、角色未知或涉及两类节点时需要两者。
- `only_count=true` 时可以省略 `page`，不返回详情记录。

### 调用示例

按已知任务流 ID 查询执行状态。

```json
{
  "page": { "offset": 0, "limit": 20 },
  "only_count": false,
  "exact_include_conditions": {
    "workflow_id": ["wf-20240601-0001"]
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 1,
    "items": [
      {
        "tenant_id": null,
        "workflow_id": "wf-20240601-0001",
        "trigger_id": "trigger-9f2b",
        "type": "upgrade_agent",
        "bk_biz_id": [2],
        "bk_networkarea_id": [0],
        "bk_networkunit_id": [1],
        "operator": "admin",
        "operate_time": 1719820000000,
        "finish_time": 0,
        "status": "running",
        "node_role": ["agent"],
        "deploy_policy_ids": []
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                 |
| ---------- | -------- | -------------------- |
| code       | int32    | 状态码，`0` 表示成功 |
| message    | string   | 请求信息             |
| request_id | string   | 请求 ID              |
| error      | object   | 错误信息             |
| permission | object   | 权限信息             |
| data       | object   | 响应数据             |

#### data

| 参数名称 | 参数类型 | 描述                                           |
| -------- | -------- | ---------------------------------------------- |
| total    | int64    | 当前查询范围内匹配的任务流总数，不是当前页条数 |
| items    | array    | 当前页任务流；`only_count=true` 时为空         |

#### data.items[n]

| 参数名称          | 参数类型      | 描述                                                                         |
| ----------------- | ------------- | ---------------------------------------------------------------------------- |
| tenant_id         | string / null | 租户 ID；当前列表响应未填充值，返回 `null`。这是响应字段，请求不需要租户参数 |
| workflow_id       | string        | 任务流 ID，可用于查询该任务流的操作列表                                      |
| trigger_id        | string        | 触发器 ID，不等同于任务流 ID                                                 |
| type              | string        | 任务流类型                                                                   |
| bk_biz_id         | int64 array   | 业务 ID 列表                                                                 |
| bk_networkarea_id | int64 array   | 管控区域 ID 列表                                                             |
| bk_networkunit_id | int64 array   | 管控单元 ID 列表                                                             |
| operator          | string        | 历史操作人                                                                   |
| operate_time      | int64         | 操作时间，Unix 毫秒时间戳                                                    |
| finish_time       | int64         | 完成时间，Unix 毫秒时间戳；尚未记录时为 `0`                                  |
| status            | string        | 任务流状态，取值同请求中的 `status`                                          |
| node_role         | string array  | 任务流记录的节点角色                                                         |
| deploy_policy_ids | int64 array   | 任务流关联的部署策略 ID 列表                                                 |
