### 描述

- 该接口提供版本：v3.0.1-alpha.1+。
- 版本变更：`v3.0.1-alpha.13+` 增加权限响应字段；`v3.0.1-alpha.19+` 增加 `plugin_operate` 鉴权。
- 该接口所需权限：plugin_operate（操作插件）。
- 该接口功能描述：根据操作 ID 列表查询插件任务流操作实例，支持仅返回总数。

### URL

POST /api/v3/plugin/workflow/operation/instance/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| only_count | bool | 否 | 是否只返回总数，不返回操作实例详情 |
| operation_id | string array | 是 | 操作 ID 列表，至少包含一个元素，且首个元素不可为空 |

### 调用示例

```json
{
  "only_count": false,
  "operation_id": ["op-plugin-0001"]
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
        "operation_id": "op-plugin-0001",
        "oper_inst_id": "inst-plugin-0001",
        "oper_inst_status": "running",
        "operation_def_name": "upgrade_plugin",
        "parent_operation_id": "",
        "action_names": ["download_package", "upgrade_plugin"],
        "life_cycle": {
          "state": "running",
          "create_time": 1719820000000,
          "start_time": 1719820010000,
          "end_time": 0,
          "stop_time": 0
        }
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| code | int32 | 状态码，`0` 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时通常为空 |
| permission | object | 权限信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| total | int64 | 当前查询命中的操作实例总数 |
| oper_inst_data | array | 操作实例简要数据列表 |

#### data.oper_inst_data[n]

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| operation_id | string | 操作 ID |
| oper_inst_id | string | 操作实例 ID |
| oper_inst_status | string | 操作实例状态，可选值：`init`、`launched`、`running`、`success`、`failed`、`timeout`、`terminated` |
| operation_def_name | string | 操作定义名称 |
| parent_operation_id | string | 父操作 ID |
| action_names | string array | 操作实例包含的动作名称列表 |
| life_cycle | object | 操作实例生命周期信息 |
| latest_action_inst_brief_data | object | 最新动作实例摘要；协议已声明，但当前插件任务流响应转换器不填充 |

#### data.oper_inst_data[n].life_cycle

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| state | string | 生命周期状态，可选值同 `oper_inst_status` |
| create_time | int64 | 创建时间，Unix 毫秒时间戳 |
| start_time | int64 | 开始时间，Unix 毫秒时间戳 |
| end_time | int64 | 结束时间，Unix 毫秒时间戳 |
| stop_time | int64 | 停止时间，Unix 毫秒时间戳 |

#### data.oper_inst_data[n].latest_action_inst_brief_data

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| name | string | 最新动作名称 |
| tags | string array | 最新动作标签 |

### 处理规则说明

- 请求校验要求 `operation_id` 至少包含一个元素，并校验首个元素不可为空。
- 后端根据首个操作 ID 定位所属任务流并校验 `plugin_operate` 权限，然后查询请求中的全部操作 ID。
- 当 `only_count=true` 时，返回 `data.total`，`data.oper_inst_data` 为空数组。
- `latest_action_inst_brief_data` 已在共享 Proto 中声明，但当前插件任务流响应转换器未设置该字段，因此成功示例不包含它。
