### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：agent_operate（操作Agent）、proxy_operate（操作Proxy）。
  - 权限说明：根据任务流涉及的节点角色动态收敛。若任务流仅涉及 Agent 节点，则只需 agent_operate；若仅涉及 Proxy 节点，则只需
    proxy_operate；若涉及两者或角色未知，则需要两者。
- 该接口功能描述：按操作 ID 或操作实例 ID 查询节点任务流操作实例列表，支持仅返回总数。

### URL

POST /api/v3/node/workflow/operation/instance/list

### 输入参数

| 参数名称     | 参数类型     | 必选 | 描述                                         |
| ------------ | ------------ | ---- | -------------------------------------------- |
| only_count   | bool         | 否   | 是否仅返回总数，`true` 时仅返回 `data.total` |
| operation_id | string array | 否   | 操作 ID 列表，按操作维度过滤实例             |
| oper_inst_id | string array | 否   | 操作实例 ID 列表，按实例维度过滤             |

#### 查询参数说明

接口调用者可按以下字段组合过滤：

| 参数名称     | 参数类型     | 描述                       |
| ------------ | ------------ | -------------------------- |
| operation_id | string array | 过滤指定操作下的实例       |
| oper_inst_id | string array | 过滤指定操作实例           |
| only_count   | bool         | 仅统计数量，不返回详情列表 |

### 调用示例

查询指定操作下的操作实例，并返回详情。

```json
{
  "only_count": false,
  "operation_id": ["op-20260421-0001"],
  "oper_inst_id": []
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260421-001",
  "data": {
    "total": 1,
    "oper_inst_data": [
      {
        "operation_id": "op-20260421-0001",
        "oper_inst_id": "inst-20260421-0001",
        "oper_inst_status": "running",
        "operation_def_name": "install_agent",
        "parent_operation_id": "",
        "action_names": ["DownloadPackage", "InstallAgent"],
        "life_cycle": {
          "state": "running",
          "create_time": 1718670000000,
          "start_time": 1718670010000,
          "end_time": 0,
          "stop_time": 0
        },
        "latest_action_inst_brief_data": {
          "name": "InstallAgent",
          "tags": ["agent", "linux"]
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

#### data

| 参数名称       | 参数类型 | 描述                         |
| -------------- | -------- | ---------------------------- |
| total          | int64    | 当前过滤条件命中的总记录条数 |
| oper_inst_data | array    | 操作实例简要数据列表         |

#### data.oper_inst_data[n]

| 参数名称                      | 参数类型     | 描述                       |
| ----------------------------- | ------------ | -------------------------- |
| operation_id                  | string       | 操作 ID                    |
| oper_inst_id                  | string       | 操作实例 ID                |
| oper_inst_status              | string       | 操作实例状态               |
| operation_def_name            | string       | 操作定义名称               |
| parent_operation_id           | string       | 父级操作 ID                |
| action_names                  | string array | 实例包含的动作名称列表     |
| life_cycle                    | object       | 实例生命周期信息           |
| latest_action_inst_brief_data | object       | 最近一次动作实例的简要信息 |

#### data.oper_inst_data[n].life_cycle

| 参数名称    | 参数类型 | 描述                      |
| ----------- | -------- | ------------------------- |
| state       | string   | 生命周期状态              |
| create_time | int64    | 创建时间，Unix 毫秒时间戳 |
| start_time  | int64    | 开始时间，Unix 毫秒时间戳 |
| end_time    | int64    | 结束时间，Unix 毫秒时间戳 |
| stop_time   | int64    | 停止时间，Unix 毫秒时间戳 |

#### data.oper_inst_data[n].latest_action_inst_brief_data

| 参数名称 | 参数类型     | 描述             |
| -------- | ------------ | ---------------- |
| name     | string       | 最近动作名称     |
| tags     | string array | 最近动作标签列表 |

### 说明

- 当 `only_count=true` 时，返回 `data.total`，`data.oper_inst_data` 为空或省略。
