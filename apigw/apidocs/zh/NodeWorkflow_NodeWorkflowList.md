### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：agent_view（查看 Agent）、proxy_view（查看 Proxy）。
- 该接口功能描述：查询节点任务流列表，支持分页、按精确条件过滤和按操作时间范围过滤。

### URL

POST /api/v3/node/workflow/list

### 输入参数

| 参数名称                     | 参数类型   | 必选 | 描述               |
|--------------------------|--------|----|------------------|
| page                     | object | 否  | 分页配置             |
| only_count               | bool   | 否  | 是否只返回总数，不返回详情    |
| exact_include_conditions | object | 否  | 精确匹配包含条件         |
| fuzzy_include_conditions | object | 否  | 模糊匹配包含条件，当前无可用字段 |
| operate_time_range       | object | 否  | 操作时间范围，单位秒       |

#### page

| 参数名称   | 参数类型  | 必选 | 描述              |
|--------|-------|----|-----------------|
| offset | int32 | 否  | 分页起始位置，起始值为 `0` |
| limit  | int32 | 否  | 每页条数            |

#### exact_include_conditions

| 参数名称               | 参数类型         | 必选 | 描述                                                                                                                                                                                |
|--------------------|--------------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| workflow_id        | string array | 否  | 任务流 ID 列表                                                                                                                                                                         |
| type               | string array | 否  | 任务流类型列表，可选值：`install_agent`、`install_proxy`、`upgrade_agent`、`upgrade_proxy`、`reconfig_agent`、`reconfig_proxy`、`restart_agent`、`restart_proxy`、`uninstall_agent`、`uninstall_proxy` |
| bk_biz_id          | int64 array  | 否  | 业务 ID 列表                                                                                                                                                                          |
| status             | string array | 否  | 任务流状态列表，可选值：`running`、`success`、`failed`、`partial_failed`                                                                                                                         |
| operator           | string array | 否  | 操作人列表                                                                                                                                                                             |
| bk_host_innerip    | string array | 否  | 主机内网 IPv4 列表                                                                                                                                                                      |
| bk_host_innerip_v6 | string array | 否  | 主机内网 IPv6 列表                                                                                                                                                                      |
| node_role          | string array | 否  | 节点角色列表，可选值：`blank`、`agent`、`proxy`                                                                                                                                                |

#### fuzzy_include_conditions

当前对象为空，传空对象 `{}` 或不传。

#### operate_time_range

| 参数名称                | 参数类型  | 必选 | 描述       |
|---------------------|-------|----|----------|
| start_timestamp_sec | int64 | 否  | 起始时间戳（秒） |
| end_timestamp_sec   | int64 | 否  | 结束时间戳（秒） |

**权限说明**：

- 当未传 `node_role` 时，后端按安全策略同时校验 `agent_view` 和 `proxy_view`。
- 当传入 `node_role` 时，后端会按角色收敛需要的权限动作。

### 调用示例

查询业务 `2` 下 `upgrade_agent` 且状态为 `running` 的任务流。

```json
{
  "page": {
    "offset": 0,
    "limit": 20
  },
  "only_count": false,
  "exact_include_conditions": {
    "bk_biz_id": [
      2
    ],
    "type": [
      "upgrade_agent"
    ],
    "status": [
      "running"
    ],
    "node_role": [
      "agent"
    ]
  },
  "fuzzy_include_conditions": {},
  "operate_time_range": {
    "start_timestamp_sec": 1717171200,
    "end_timestamp_sec": 1719859599
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
        "workflow_id": "wf-20240601-0001",
        "trigger_id": "trigger-9f2b",
        "type": "upgrade_agent",
        "bk_biz_id": [
          2
        ],
        "bk_networkarea_id": [
          0
        ],
        "bk_networkunit_id": [
          1
        ],
        "operator": "admin",
        "operate_time": 1719820000000,
        "finish_time": 1719820300000,
        "status": "running",
        "bk_biz_name": [
          "BlueKing"
        ],
        "node_role": [
          "agent"
        ]
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称       | 参数类型   | 描述           |
|------------|--------|--------------|
| code       | int32  | 状态码，`0` 表示成功 |
| message    | string | 请求信息         |
| request_id | string | 请求 ID        |
| error      | object | 错误信息，成功时通常为空 |
| permission | object | 权限信息         |
| data       | object | 响应数据         |

#### error

| 参数名称    | 参数类型   | 描述     |
|---------|--------|--------|
| system  | string | 错误系统标识 |
| message | string | 错误消息   |
| details | array  | 错误详情列表 |

#### error.details[n]

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | string | 错误代码 |
| message | string | 错误消息 |

#### permission

| 参数名称        | 参数类型   | 描述      |
|-------------|--------|---------|
| system      | string | 权限系统 ID |
| system_name | string | 权限系统名称  |
| apply_url   | string | 权限申请链接  |
| actions     | array  | 关联动作列表  |

#### permission.actions[n]

| 参数名称                   | 参数类型   | 描述     |
|------------------------|--------|--------|
| id                     | string | 动作 ID  |
| name                   | string | 动作名称   |
| related_resource_types | array  | 关联资源类型 |

#### data

| 参数名称  | 参数类型  | 描述             |
|-------|-------|----------------|
| total | int64 | 当前过滤条件命中的总记录条数 |
| items | array | 节点任务流列表        |

#### data.items[n]

| 参数名称              | 参数类型         | 描述                                               |
|-------------------|--------------|--------------------------------------------------|
| workflow_id       | string       | 任务流 ID                                           |
| trigger_id        | string       | 触发器 ID                                           |
| type              | string       | 任务流类型，枚举值同 `exact_include_conditions.type`       |
| bk_biz_id         | int64 array  | 业务 ID 列表                                         |
| bk_networkarea_id | int64 array  | 管控区域 ID 列表                                       |
| bk_networkunit_id | int64 array  | 管控单元 ID 列表                                       |
| operator          | string       | 操作人                                              |
| operate_time      | int64        | 操作时间，Unix 毫秒时间戳                                  |
| finish_time       | int64        | 完成时间，Unix 毫秒时间戳                                  |
| status            | string       | 任务流状态，枚举值同 `exact_include_conditions.status`     |
| bk_biz_name       | string array | 业务名称列表                                           |
| node_role         | string array | 节点角色列表，枚举值同 `exact_include_conditions.node_role` |
