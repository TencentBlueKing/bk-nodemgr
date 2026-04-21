### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：agent_operate（操作Agent）、proxy_operate（操作Proxy）。
  - 权限说明：根据任务流涉及的节点角色动态收敛。若任务流仅涉及 Agent 节点，则只需 agent_operate；若仅涉及 Proxy 节点，则只需 proxy_operate；若涉及两者或角色未知，则需要两者。
  - 该接口功能描述：查询指定节点任务流下的操作列表，支持分页、按节点部署维度精确过滤、按操作状态过滤。

### URL

POST /api/v3/node/workflow/operation/list

### 输入参数

| 参数名称                 | 参数类型 | 必选 | 描述                             |
| ------------------------ | -------- | ---- | -------------------------------- |
| only_count               | bool     | 否   | 是否只返回总数，不返回详情       |
| page                     | object   | 否   | 分页配置                         |
| workflow_id              | string   | 是   | 节点任务流 ID                    |
| exact_include_conditions | object   | 否   | 精确匹配包含条件                 |
| fuzzy_include_conditions | object   | 否   | 模糊匹配包含条件，当前无可用字段 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                       |
| -------- | -------- | ---- | -------------------------- |
| offset   | int32    | 否   | 分页起始位置，起始值为 `0` |
| limit    | int32    | 否   | 每页条数                   |

#### exact_include_conditions

| 参数名称           | 参数类型     | 必选 | 描述                                                                                              |
| ------------------ | ------------ | ---- | ------------------------------------------------------------------------------------------------- |
| node_version       | string array | 否   | 节点版本列表                                                                                      |
| bk_host_innerip    | string array | 否   | 主机内网 IPv4 列表                                                                                |
| bk_host_innerip_v6 | string array | 否   | 主机内网 IPv6 列表                                                                                |
| bk_biz_id          | int64 array  | 否   | 业务 ID 列表                                                                                      |
| bk_networkarea_id  | int64 array  | 否   | 管控区域 ID 列表                                                                                  |
| bk_networkunit_id  | int64 array  | 否   | 管控单元 ID 列表                                                                                  |
| state              | string array | 否   | 操作状态列表，可选值：`init`、`launched`、`running`、`success`、`failed`、`timeout`、`terminated` |

#### fuzzy_include_conditions

当前对象为空，传空对象 `{}` 或不传。

### 调用示例

查询任务流 `wf-20240601-0001` 下状态为 `running` 且业务为 `2` 的操作列表。

```json
{
  "only_count": false,
  "page": {
    "offset": 0,
    "limit": 20
  },
  "workflow_id": "wf-20240601-0001",
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "state": ["running"],
    "node_version": ["2.1.0"]
  },
  "fuzzy_include_conditions": {}
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
        "instance_ids": ["inst-001", "inst-002"],
        "operator": "admin",
        "create_time": 1719820000000,
        "node_deployment_info": {
          "bk_host_id": 1001,
          "bk_biz_id": 2,
          "bk_host_innerip_list": ["10.0.0.1"],
          "bk_host_innerip_v6_list": ["fe80::1"],
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
            "name": "install_agent",
            "tags": ["main"]
          }
        }
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

#### error

| 参数名称 | 参数类型 | 描述         |
| -------- | -------- | ------------ |
| system   | string   | 错误系统标识 |
| message  | string   | 错误消息     |
| details  | array    | 错误详情列表 |

#### error.details[n]

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| code     | string   | 错误代码 |
| message  | string   | 错误消息 |

#### permission

| 参数名称    | 参数类型 | 描述         |
| ----------- | -------- | ------------ |
| system      | string   | 权限系统 ID  |
| system_name | string   | 权限系统名称 |
| apply_url   | string   | 权限申请链接 |
| actions     | array    | 关联动作列表 |

#### permission.actions[n]

| 参数名称               | 参数类型 | 描述         |
| ---------------------- | -------- | ------------ |
| id                     | string   | 动作 ID      |
| name                   | string   | 动作名称     |
| related_resource_types | array    | 关联资源类型 |

#### data

| 参数名称   | 参数类型 | 描述                         |
| ---------- | -------- | ---------------------------- |
| total      | int64    | 当前过滤条件命中的总记录条数 |
| operations | array    | 操作列表                     |

#### data.operations[n]

| 参数名称                    | 参数类型     | 描述                          |
| --------------------------- | ------------ | ----------------------------- |
| operation_id                | string       | 操作 ID                       |
| instance_ids                | string array | 操作实例 ID 列表              |
| operator                    | string       | 操作人                        |
| create_time                 | int64        | 操作创建时间，Unix 毫秒时间戳 |
| node_deployment_info        | object       | 节点部署信息                  |
| latest_oper_inst_brief_data | object       | 最新操作实例摘要              |

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

| 参数名称                      | 参数类型 | 描述         |
| ----------------------------- | -------- | ------------ |
| life_cycle                    | object   | 生命周期摘要 |
| latest_action_inst_brief_data | object   | 最新动作摘要 |

#### data.operations[n].latest_oper_inst_brief_data.life_cycle

| 参数名称    | 参数类型 | 描述                      |
| ----------- | -------- | ------------------------- |
| state       | string   | 状态值                    |
| create_time | int64    | 创建时间，Unix 毫秒时间戳 |
| start_time  | int64    | 开始时间，Unix 毫秒时间戳 |
| end_time    | int64    | 结束时间，Unix 毫秒时间戳 |
| stop_time   | int64    | 停止时间，Unix 毫秒时间戳 |

#### data.operations[n].latest_oper_inst_brief_data.latest_action_inst_brief_data

| 参数名称 | 参数类型     | 描述         |
| -------- | ------------ | ------------ |
| name     | string       | 最新动作名称 |
| tags     | string array | 最新动作标签 |

### 处理规则说明

- `workflow_id` 必填，不可为空。
- 查询过程先按任务流关联的 `trigger_id` 过滤操作，再根据操作 token 关联节点部署信息。
- 当 `only_count=true` 时，仅返回 `data.total`，`data.operations` 为空数组。
