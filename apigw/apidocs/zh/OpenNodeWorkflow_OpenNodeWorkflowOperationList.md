### 描述

- 该接口提供版本：v3.0.1-alpha.114+。
- 该接口所需权限：system_access（系统访问）、agent_history_view（查看 Agent 历史）、proxy_history_view（查看 Proxy 历史）。历史权限按目标任务流记录的业务和节点角色校验：`blank`、`agent` 需要 Agent 历史权限，`proxy` 需要 Proxy 历史权限；涉及两者、角色缺失或未知时需要两者。
- 该接口功能描述：查询指定节点任务流的操作列表、节点部署信息及最新执行实例摘要，支持分页和精确过滤。
- 用户身份由 API 网关认证并传递，请求体不需要用户名或租户参数。

### URL

POST /api/v3/open/node/workflow/operation/list

### 输入参数

| 参数名称                 | 参数类型 | 必选 | 描述                                       |
| ------------------------ | -------- | ---- | ------------------------------------------ |
| workflow_id              | string   | 是   | 非空节点任务流 ID                          |
| only_count               | bool     | 否   | 是否仅返回总数，默认为 `false`             |
| page                     | object   | 是   | 分页配置；仅统计数量时也应提供有效分页参数 |
| exact_include_conditions | object   | 否   | 精确匹配包含条件                           |
| fuzzy_include_conditions | object   | 否   | 当前没有可用字段，省略或传 `{}`            |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                                 |
| -------- | -------- | ---- | ------------------------------------ |
| offset   | int32    | 否   | 起始偏移量，默认为 `0`，不得小于 `0` |
| limit    | int32    | 是   | 每页条数，范围 `1`~`500`             |

#### exact_include_conditions

| 参数名称           | 参数类型     | 必选 | 描述                                              |
| ------------------ | ------------ | ---- | ------------------------------------------------- |
| node_version       | string array | 否   | 节点版本列表                                      |
| bk_host_innerip    | string array | 否   | 主机内网 IPv4 列表                                |
| bk_host_innerip_v6 | string array | 否   | 主机内网 IPv6 列表                                |
| bk_biz_id          | int64 array  | 否   | 业务 ID 列表                                      |
| bk_networkarea_id  | int64 array  | 否   | 管控区域 ID 列表                                  |
| bk_networkunit_id  | int64 array  | 否   | 管控单元 ID 列表                                  |
| state              | string array | 否   | 操作状态列表，例如 `running`、`success`、`failed` |

### 调用示例

查询指定任务流下业务 `2` 的运行中操作。

```json
{
  "workflow_id": "wf-20240601-0001",
  "only_count": false,
  "page": { "offset": 0, "limit": 20 },
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "state": ["running"]
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
    "operations": [
      {
        "operation_id": "op-8a3d",
        "instance_ids": ["inst-001"],
        "operator": "admin",
        "create_time": 1719820000000,
        "node_deployment_info": {
          "bk_host_id": 1001,
          "bk_biz_id": 2,
          "bk_host_innerip_list": ["10.0.0.1"],
          "bk_host_innerip_v6_list": [],
          "bk_networkarea_id": 0,
          "bk_networkunit_id": 1,
          "node_version": "2.1.0"
        },
        "latest_oper_inst_brief_data": {
          "life_cycle": {
            "state": "running",
            "create_time": 1719820000000,
            "start_time": 1719820010000,
            "end_time": 0,
            "stop_time": 0
          },
          "latest_action_inst_brief_data": {
            "name": "InstallAgent",
            "tags": []
          }
        }
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

| 参数名称   | 参数类型 | 描述                                     |
| ---------- | -------- | ---------------------------------------- |
| total      | int64    | 过滤后匹配的记录总数，不是当前页条数     |
| operations | array    | 当前页操作列表；`only_count=true` 时为空 |

#### data.operations[n]

| 参数名称                    | 参数类型     | 描述                                   |
| --------------------------- | ------------ | -------------------------------------- |
| operation_id                | string       | 操作 ID，用于查询该操作的执行实例      |
| instance_ids                | string array | 此操作关联的实例 ID 列表               |
| operator                    | string       | 操作人                                 |
| create_time                 | int64        | 操作创建时间，Unix 毫秒时间戳          |
| node_deployment_info        | object       | 与操作关联的节点部署信息               |
| latest_oper_inst_brief_data | object       | 最新操作实例摘要；尚无摘要时可能不返回 |

#### data.operations[n].node_deployment_info

| 参数名称                | 参数类型     | 描述               |
| ----------------------- | ------------ | ------------------ |
| bk_host_id              | int64        | 主机 ID            |
| bk_biz_id               | int64        | 业务 ID            |
| bk_host_innerip_list    | string array | 主机内网 IPv4 列表 |
| bk_host_innerip_v6_list | string array | 主机内网 IPv6 列表 |
| bk_networkarea_id       | int64        | 管控区域 ID        |
| bk_networkunit_id       | int64        | 管控单元 ID        |
| node_version            | string       | 节点版本           |

#### data.operations[n].latest_oper_inst_brief_data

| 参数名称                      | 参数类型 | 描述                       |
| ----------------------------- | -------- | -------------------------- |
| life_cycle                    | object   | 最新操作实例生命周期       |
| latest_action_inst_brief_data | object   | 最新动作摘要，可能为空对象 |

#### data.operations[n].latest_oper_inst_brief_data.life_cycle

| 参数名称    | 参数类型 | 描述                      |
| ----------- | -------- | ------------------------- |
| state       | string   | 操作实例状态              |
| create_time | int64    | 创建时间，Unix 毫秒时间戳 |
| start_time  | int64    | 开始时间，Unix 毫秒时间戳 |
| end_time    | int64    | 结束时间，Unix 毫秒时间戳 |
| stop_time   | int64    | 停止时间，Unix 毫秒时间戳 |

尚未记录的生命周期时间为 `0`。

#### data.operations[n].latest_oper_inst_brief_data.latest_action_inst_brief_data

| 参数名称 | 参数类型     | 描述     |
| -------- | ------------ | -------- |
| name     | string       | 动作名称 |
| tags     | string array | 动作标签 |
