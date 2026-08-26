### 描述

- 该接口提供版本：v3.0.1-alpha.74+
- 该接口所需权限：无。
- 该接口名称：PackageExportPlugin。
- 该接口功能描述：根据已发布的插件包名称与版本，发起异步插件资源包导出工作流（将目标插件包打包为可下载文件），并返回工作流ID。可通过 PackageExportResult 接口查询导出结果及临时下载地址。

### URL

POST /api/v3/package/workflow/export/plugin

### 输入参数

| 参数名称           | 参数类型 | 必选 | 描述             |
| ------------------ | -------- | ---- | ---------------- |
| plugin_pkg_name    | string   | 是   | 目标插件包名称   |
| plugin_pkg_version | string   | 是   | 目标插件包版本号 |

### 调用示例

```json
{
  "plugin_pkg_name": "bkmonitorbeat",
  "plugin_pkg_version": "2.1.3"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "workflow_id": "workflow-export-abc123def456"
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

| 参数名称    | 参数类型 | 描述                                              |
| ----------- | -------- | ------------------------------------------------- |
| workflow_id | string   | 导出工作流ID，可用于 PackageExportResult 查询结果 |
