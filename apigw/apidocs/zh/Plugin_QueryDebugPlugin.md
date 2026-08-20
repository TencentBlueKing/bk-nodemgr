### 描述

- 该接口提供版本：v3.0.1-alpha.72+。
- 该接口所需权限：⚠️ 待确认（接口尚未实现 handler，权限待补充）。
- 该接口功能描述：查询插件调试工作流的聚合状态和执行结果，包括各操作实例的生命周期与执行日志。

### URL

POST /api/v3/plugin/query_debug

### 输入参数

| 参数名称    | 参数类型 | 必选 | 描述                                     |
| ----------- | -------- | ---- | ---------------------------------------- |
| workflow_id | string   | 是   | 要查询的调试工作流ID，由启动调试接口返回 |

### 调用示例

```json
{
  "workflow_id": "workflow-abc123def456"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {
    "status": "running",
    "is_finish": false,
    "operations": [
      {
        "operation_id": "op-1001",
        "last_instance_id": "inst-5001",
        "oper_inst_logs": {
          "inst-5001": {
            "life_cycle": {
              "state": "running",
              "create_time": 1735000000,
              "start_time": 1735000010
            },
            "message": {
              "logs": [
                {
                  "time": 1735000010,
                  "level": "info",
                  "text_zh": "开始安装插件",
                  "text_en": "start installing plugin"
                }
              ]
            },
            "display_name_zh": "安装插件",
            "display_name_en": "Install Plugin"
          }
        },
        "extra_execution_logs": {
          "logs": []
        }
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                   |
| ---------- | -------- | ---------------------- |
| code       | int32    | 状态码，0表示成功      |
| message    | string   | 请求信息               |
| request_id | string   | 请求ID                 |
| error      | object   | 错误信息，成功时为null |
| data       | object   | 响应数据               |

#### data

| 参数名称   | 参数类型 | 描述                                                               |
| ---------- | -------- | ------------------------------------------------------------------ |
| status     | string   | 调试工作流状态（枚举值：running、success、failed、partial_failed） |
| is_finish  | bool     | 调试工作流是否已结束                                               |
| operations | array    | 调试工作流中的操作列表                                             |

**status 枚举说明**:

- `running`: 工作流执行中
- `success`: 工作流执行成功
- `failed`: 工作流执行失败
- `partial_failed`: 工作流部分操作执行失败

#### data.operations[n]

| 参数名称             | 参数类型 | 描述                                               |
| -------------------- | -------- | -------------------------------------------------- |
| operation_id         | string   | 操作ID                                             |
| last_instance_id     | string   | 最近一次执行的操作实例ID                           |
| oper_inst_logs       | object   | 各执行实例ID到操作数据（WorkflowActionData）的映射 |
| extra_execution_logs | object   | 额外执行日志（WorkflowActionMessage）              |

#### data.operations[n].oper_inst_logs.<key>

| 参数名称          | 参数类型 | 描述             |
| ----------------- | -------- | ---------------- |
| life_cycle        | object   | 生命周期信息     |
| message           | object   | 操作日志消息     |
| display_name_zh   | string   | 操作中文显示名   |
| display_name_en   | string   | 操作英文显示名   |
| sub_workflow_refs | array    | 子工作流引用列表 |

#### data.operations[n].oper_inst_logs.<key>.life_cycle

| 参数名称    | 参数类型 | 描述                    |
| ----------- | -------- | ----------------------- |
| state       | string   | 生命周期状态            |
| create_time | int64    | 创建时间（Unix 时间戳） |
| start_time  | int64    | 开始时间（Unix 时间戳） |
| end_time    | int64    | 结束时间（Unix 时间戳） |
| stop_time   | int64    | 停止时间（Unix 时间戳） |

#### data.operations[n].oper_inst_logs.<key>.message

| 参数名称 | 参数类型 | 描述         |
| -------- | -------- | ------------ |
| logs     | array    | 日志消息列表 |

#### data.operations[n].oper_inst_logs.<key>.message.logs[n]

| 参数名称 | 参数类型 | 描述                    |
| -------- | -------- | ----------------------- |
| time     | int64    | 日志时间（Unix 时间戳） |
| level    | string   | 日志级别                |
| text_zh  | string   | 中文日志内容            |
| text_en  | string   | 英文日志内容            |

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
