### 描述

- 该接口提供版本：v3.0.1-alpha.74+
- 该接口所需权限：无。
- 该接口功能描述：根据工作流ID查询插件资源包导出工作流的执行结果。工作流执行成功后会返回临时下载地址及过期时间。

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

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123457",
  "data": {
    "status": "success",
    "error_message": "",
    "download_url": "https://example.com/download/plugin-export-abc123.tgz",
    "download_url_expired_at": 1767225600000
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
| error_message           | string   | 否   | 导出失败时的错误信息，仅在 `status` 为 `failed` 或 `partial_failed` 时返回 |
| download_url            | string   | 否   | 导出成功后返回的临时插件包下载地址，仅在 `status` 为 `success` 时返回      |
| download_url_expired_at | int64    | 否   | 临时下载地址的过期时间（Unix 毫秒时间戳），与 `download_url` 同步返回      |

**status 枚举说明**：

- `running`：工作流执行中，需要继续查询。
- `success`：工作流执行成功，可使用 `download_url` 下载插件包。
- `failed`：工作流执行失败，可结合 `error_message` 查看失败原因。
- `partial_failed`：工作流部分操作执行失败，可结合 `error_message` 查看失败原因。

> 当 `status` 为 `running` 时，`error_message`、`download_url`、`download_url_expired_at` 字段均不会返回；下载地址过期后不可继续使用。
