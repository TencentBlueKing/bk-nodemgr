### Description

- API Version: v3.0.1-alpha.72+
- Required Permission: None.
- Function: Query the execution result of a plugin V3 package import workflow. Use the workflow_id returned by the PackageImport API to query the import progress, status, and execution logs of each operation.

### URL

POST /api/v3/package/workflow/import_result

### Request Parameters

| Parameter   | Type   | Required | Description                                    |
| ----------- | ------ | -------- | ---------------------------------------------- |
| workflow_id | string | Yes      | Workflow ID, returned by the PackageImport API |

### Request Example

```json
{
  "workflow_id": "workflow-abc123def456"
}
```

### Response Example

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

### Response Parameters

| Parameter  | Type   | Description                          |
| ---------- | ------ | ------------------------------------ |
| code       | int32  | Status code, 0 for success           |
| message    | string | Response message                     |
| request_id | string | Request ID                           |
| error      | object | Error information (empty on success) |
| data       | object | Response data                        |

#### data

| Parameter  | Type   | Description                                                                               |
| ---------- | ------ | ----------------------------------------------------------------------------------------- |
| status     | string | Workflow status (enum: running, success, failed, partial_failed)                          |
| is_finish  | bool   | Whether the workflow has finished (true when status is success, failed or partial_failed) |
| operations | array  | List of operation execution results                                                       |

**status enum**:

- `running`: The workflow is running
- `success`: The workflow succeeded
- `failed`: The workflow failed
- `partial_failed`: Part of the workflow operations failed

#### data.operations[n]

| Parameter            | Type   | Description                                                                   |
| -------------------- | ------ | ----------------------------------------------------------------------------- |
| operation_id         | string | Operation ID, which is package_plugin_v3_import for this import               |
| last_instance_id     | string | Operation instance ID                                                         |
| oper_inst_logs       | object | Execution logs of each action in the operation instance, keyed by action name |
| extra_execution_logs | object | Extra execution logs of the operation instance                                |

**oper_inst_logs action names** (keys):

| Action name                                   | Description                        |
| --------------------------------------------- | ---------------------------------- |
| package_import_plugin_v3_pkg_fetch_and_upload | Fetch and upload plugin V3 package |
| package_publish_plugin_v3_pkg                 | Publish plugin V3 package          |
| package_release_plugin_enable                 | Enable release plugin package      |
| package_release_plugin_hidden                 | Hide release plugin package        |

#### data.operations[n].oper_inst_logs[action]

| Parameter         | Type   | Description                        |
| ----------------- | ------ | ---------------------------------- |
| life_cycle        | object | Action lifecycle information       |
| message           | object | Action log information             |
| display_name_zh   | string | Chinese display name of the action |
| display_name_en   | string | English display name of the action |
| sub_workflow_refs | array  | List of sub-workflow references    |

#### data.operations[n].oper_inst_logs[action].sub_workflow_refs[m]

| Parameter       | Type   | Description                |
| --------------- | ------ | -------------------------- |
| workflow_id     | string | Sub-workflow ID            |
| workflow_domain | string | Domain of the sub-workflow |

#### data.operations[n].oper_inst_logs[action].life_cycle

| Parameter   | Type   | Description                                                                        |
| ----------- | ------ | ---------------------------------------------------------------------------------- |
| state       | string | Action state (enum: init, launched, running, success, failed, timeout, terminated) |
| create_time | int64  | Create time (Unix millisecond timestamp)                                           |
| start_time  | int64  | Start time (Unix millisecond timestamp)                                            |
| end_time    | int64  | End time (Unix millisecond timestamp)                                              |
| stop_time   | int64  | Stop time (Unix millisecond timestamp)                                             |

#### data.operations[n].oper_inst_logs[action].message

| Parameter | Type  | Description          |
| --------- | ----- | -------------------- |
| logs      | array | List of log messages |

#### data.operations[n].oper_inst_logs[action].message.logs[m]

| Parameter | Type   | Description                           |
| --------- | ------ | ------------------------------------- |
| time      | int64  | Log time (Unix millisecond timestamp) |
| level     | string | Log level (enum: INFO, WARN, ERROR)   |
| text_zh   | string | Chinese log content                   |
| text_en   | string | English log content                   |

#### data.operations[n].extra_execution_logs

| Parameter | Type  | Description                                                                                    |
| --------- | ----- | ---------------------------------------------------------------------------------------------- |
| logs      | array | List of log messages, same structure as data.operations[n].oper_inst_logs[action].message.logs |
