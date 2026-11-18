### 描述

- 该接口提供版本：v3.0.1-alpha.114+。
- 该接口所需权限：system_access（系统访问）、agent_history_view（查看 Agent 历史）、proxy_history_view（查看 Proxy 历史）。历史权限按目标任务流记录的业务和节点角色校验：`blank`、`agent` 需要 Agent 历史权限，`proxy` 需要 Proxy 历史权限；涉及两者、角色缺失或未知时需要两者。
- 该接口功能描述：按节点操作 ID 查询执行实例及历史执行状态，可进一步按实例 ID 过滤，支持仅返回总数。
- 用户身份由 API 网关认证并传递，请求体不需要用户名或租户参数。

### URL

POST /api/v3/open/node/workflow/operation/instance/list

### 输入参数

| 参数名称     | 参数类型     | 必选 | 描述                                                              |
| ------------ | ------------ | ---- | ----------------------------------------------------------------- |
| only_count   | bool         | 否   | 是否仅返回总数，默认为 `false`                                    |
| operation_id | string array | 是   | 非空操作 ID 列表；排障时每次传一个操作 ID                         |
| oper_inst_id | string array | 否   | 在所传操作范围内进一步筛选的实例 ID 列表，不能代替 `operation_id` |

### 使用限制

- 排障时每次传一个 `operation_id`：从操作列表取得操作 ID，再查询该操作的实例。`operation_id` 的参数类型仍为数组。
- 不支持跨任务流批量排障，不要在一次请求中混入不同任务流的操作。
- 接口没有分页参数；`only_count` 仅控制是否返回实例详情，不提供分页或排序控制。

### 调用示例

查询单个操作的执行实例。

```json
{
  "only_count": false,
  "operation_id": ["op-8a3d"]
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
    "oper_inst_data": [
      {
        "operation_id": "op-8a3d",
        "oper_inst_id": "inst-001",
        "oper_inst_status": "running",
        "operation_def_name": "install_agent",
        "parent_operation_id": "",
        "action_names": ["DownloadPackage", "InstallAgent"],
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

| 参数名称       | 参数类型 | 描述                                       |
| -------------- | -------- | ------------------------------------------ |
| total          | int64    | 匹配的操作实例总数                         |
| oper_inst_data | array    | 操作实例摘要列表；`only_count=true` 时为空 |

#### data.oper_inst_data[n]

| 参数名称                      | 参数类型     | 描述                                              |
| ----------------------------- | ------------ | ------------------------------------------------- |
| operation_id                  | string       | 所属操作 ID                                       |
| oper_inst_id                  | string       | 操作实例 ID，用于查询步骤日志                     |
| oper_inst_status              | string       | 操作实例状态，例如 `running`、`success`、`failed` |
| operation_def_name            | string       | 操作定义名称                                      |
| parent_operation_id           | string       | 父操作 ID                                         |
| action_names                  | string array | 实例中的动作名称列表                              |
| life_cycle                    | object       | 操作实例生命周期                                  |
| latest_action_inst_brief_data | object       | 最新动作摘要，可能为空对象                        |

#### data.oper_inst_data[n].life_cycle

| 参数名称    | 参数类型 | 描述                      |
| ----------- | -------- | ------------------------- |
| state       | string   | 操作实例生命周期状态      |
| create_time | int64    | 创建时间，Unix 毫秒时间戳 |
| start_time  | int64    | 开始时间，Unix 毫秒时间戳 |
| end_time    | int64    | 结束时间，Unix 毫秒时间戳 |
| stop_time   | int64    | 停止时间，Unix 毫秒时间戳 |

尚未记录的生命周期时间为 `0`。

#### data.oper_inst_data[n].latest_action_inst_brief_data

| 参数名称 | 参数类型     | 描述     |
| -------- | ------------ | -------- |
| name     | string       | 动作名称 |
| tags     | string array | 动作标签 |
