### 描述

- 该接口提供版本：v3.0.1-alpha.78+
- 该接口所需权限：无。
- 该接口功能描述：根据工作流ID查询插件资源包导出工作流的执行结果及是否结束。工作流执行成功后会返回临时下载地址及过期时间。

### URL

POST /api/v3/package/workflow/export_result

### 输入参数

| 参数名称    | 参数类型 | 必选 | 描述                                   |
| ----------- | -------- | ---- | -------------------------------------- |
| workflow_id | string   | 是   | PackageExportPlugin 接口返回的工作流ID |

### 调用示例

```json
{
  "workflow_id": "workflow-export-abc123def456"
}
```

### 响应示例（执行成功）

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123457",
  "data": {
    "status": "success",
    "is_finish": true,
    "download_url": "https://bk-nodemgr-file.example.com/api/v3/export/download/origin_plugin_package?token=AQ2hR9pK7vN4mT6xQ8sW0yZaBcDeFgHiJkLmNoPqRsTuVw",
    "download_url_expired_at": 1798761600000
  }
}
```

### 响应示例（执行中）

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123458",
  "data": {
    "status": "running",
    "is_finish": false
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
| permission | object   | 权限信息               |
| data       | object   | 响应数据               |

#### data

| 参数名称                | 参数类型 | 必选 | 描述                                                                       |
| ----------------------- | -------- | ---- | -------------------------------------------------------------------------- |
| status                  | string   | 是   | 工作流状态（枚举值：running、success、failed、partial_failed）             |
| is_finish               | bool     | 是   | 工作流是否已结束；`success`、`failed`、`partial_failed` 时为 `true`        |
| download_url            | string   | 否   | 导出成功后返回的临时插件包下载地址，仅在 `status` 为 `success` 时返回      |
| download_url_expired_at | int64    | 否   | 临时下载地址的过期时间（Unix 毫秒时间戳），与 `download_url` 同步返回      |

**status 枚举说明**：

- `running`：工作流执行中，需要继续查询。
- `success`：工作流执行成功，可使用 `download_url` 下载插件包。
- `failed`：工作流执行失败，不返回下载地址。
- `partial_failed`：工作流部分操作执行失败，不返回下载地址。

> 当 `is_finish` 为 `false` 时，需要继续查询工作流结果。`download_url` 已包含下载 token，可直接用于下载；地址过期后不可继续使用。
