### 描述

- 该接口提供版本：v3.0.1-alpha.72+
- 该接口所需权限：无。
- 该接口功能描述：查询插件V3资源包导入工作流的执行结果。通过 PackageImport 接口返回的 workflow_id 查询导入进度、状态以及各操作（operation）的执行日志。

### URL

POST /api/v3/package/workflow/import_result

### 输入参数

| 参数名称    | 参数类型 | 必选 | 描述                                |
| ----------- | -------- | ---- | ----------------------------------- |
| workflow_id | string   | 是   | 工作流ID，由 PackageImport 接口返回 |

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
  "request_id": "req-123456",
  "data": {
    "status": "success",
    "is_finish": true,
    "operations": [
      {
        "operation_id": "package_plugin_v3_import",
        "last_instance_id": "oper-inst-xxx",
        "oper_inst_logs": {
          "package_import_plugin_v3_pkg_fetch_and_upload": {
            "life_cycle": {
              "state": "success",
              "create_time": 1738896000000,
              "start_time": 1738896001000,
              "end_time": 1738896050000,
              "stop_time": 0
            },
            "message": {
              "logs": [
                {
                  "time": 1738896001000,
                  "level": "INFO",
                  "text_zh": "开始下载并上传插件V3资源包",
                  "text_en": "Start to fetch and upload plugin V3 package"
                }
              ]
            },
            "display_name_zh": "获取并上传插件V3资源包",
            "display_name_en": "Fetch And Upload Plugin V3 Package",
            "sub_workflow_refs": []
          },
          "package_publish_plugin_v3_pkg": {
            "life_cycle": {
              "state": "success",
              "create_time": 1738896051000,
              "start_time": 1738896052000,
              "end_time": 1738896070000,
              "stop_time": 0
            },
            "message": {
              "logs": []
            },
            "display_name_zh": "发布插件V3资源包",
            "display_name_en": "Publish Plugin V3 Package",
            "sub_workflow_refs": []
          },
          "package_release_plugin_enable": {
            "life_cycle": {
              "state": "success",
              "create_time": 1738896071000,
              "start_time": 1738896072000,
              "end_time": 1738896080000,
              "stop_time": 0
            },
            "message": {
              "logs": []
            },
            "display_name_zh": "启用插件资源包",
            "display_name_en": "Enable Release Plugin Package",
            "sub_workflow_refs": []
          },
          "package_release_plugin_hidden": {
            "life_cycle": {
              "state": "success",
              "create_time": 1738896081000,
              "start_time": 1738896082000,
              "end_time": 1738896090000,
              "stop_time": 0
            },
            "message": {
              "logs": []
            },
            "display_name_zh": "隐藏插件资源包",
            "display_name_en": "Hide Release Plugin Package",
            "sub_workflow_refs": []
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
| error      | object   | 错误信息（成功时为空） |
| data       | object   | 响应数据               |

#### data

| 参数名称   | 参数类型 | 描述                                                                      |
| ---------- | -------- | ------------------------------------------------------------------------- |
| status     | string   | 工作流状态（枚举值：running、success、failed、partial_failed）            |
| is_finish  | bool     | 工作流是否执行完成（status 为 success、failed、partial_failed 时为 true） |
| operations | array    | 导入操作执行结果列表                                                      |

**status 枚举说明**:

- `running`: 工作流执行中
- `success`: 工作流执行成功
- `failed`: 工作流执行失败
- `partial_failed`: 工作流部分操作执行失败

#### data.operations[n]

| 参数名称             | 参数类型 | 描述                                               |
| -------------------- | -------- | -------------------------------------------------- |
| operation_id         | string   | 操作ID，本次导入为 package_plugin_v3_import        |
| last_instance_id     | string   | 操作实例ID                                         |
| oper_inst_logs       | object   | 操作实例中各 action 的执行日志，key 为 action 名称 |
| extra_execution_logs | object   | 操作实例的额外执行日志                             |

**oper_inst_logs 的 action 说明**（key 为 action 名称）:

| action 名称                                   | 说明                   |
| --------------------------------------------- | ---------------------- |
| package_import_plugin_v3_pkg_fetch_and_upload | 获取并上传插件V3资源包 |
| package_publish_plugin_v3_pkg                 | 发布插件V3资源包       |
| package_release_plugin_enable                 | 启用插件资源包         |
| package_release_plugin_hidden                 | 隐藏插件资源包         |

#### data.operations[n].oper_inst_logs[action]

| 参数名称          | 参数类型 | 描述                |
| ----------------- | -------- | ------------------- |
| life_cycle        | object   | action 生命周期信息 |
| message           | object   | action 日志信息     |
| display_name_zh   | string   | action 中文显示名称 |
| display_name_en   | string   | action 英文显示名称 |
| sub_workflow_refs | array    | 子工作流引用列表    |

#### data.operations[n].oper_inst_logs[action].sub_workflow_refs[m]

| 参数名称        | 参数类型 | 描述             |
| --------------- | -------- | ---------------- |
| workflow_id     | string   | 子工作流ID       |
| workflow_domain | string   | 子工作流所属领域 |

#### data.operations[n].oper_inst_logs[action].life_cycle

| 参数名称    | 参数类型 | 描述                                                                                 |
| ----------- | -------- | ------------------------------------------------------------------------------------ |
| state       | string   | action 状态（枚举值：init、launched、running、success、failed、timeout、terminated） |
| create_time | int64    | 创建时间（Unix 毫秒时间戳）                                                          |
| start_time  | int64    | 开始时间（Unix 毫秒时间戳）                                                          |
| end_time    | int64    | 结束时间（Unix 毫秒时间戳）                                                          |
| stop_time   | int64    | 停止时间（Unix 毫秒时间戳）                                                          |

#### data.operations[n].oper_inst_logs[action].message

| 参数名称 | 参数类型 | 描述         |
| -------- | -------- | ------------ |
| logs     | array    | 日志消息列表 |

#### data.operations[n].oper_inst_logs[action].message.logs[m]

| 参数名称 | 参数类型 | 描述                                  |
| -------- | -------- | ------------------------------------- |
| time     | int64    | 日志时间（Unix 毫秒时间戳）           |
| level    | string   | 日志级别（枚举值：INFO、WARN、ERROR） |
| text_zh  | string   | 中文日志内容                          |
| text_en  | string   | 英文日志内容                          |

#### data.operations[n].extra_execution_logs

| 参数名称 | 参数类型 | 描述                                                                        |
| -------- | -------- | --------------------------------------------------------------------------- |
| logs     | array    | 日志消息列表，结构同 data.operations[n].oper_inst_logs[action].message.logs |
