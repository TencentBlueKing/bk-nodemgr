### 描述

- 该接口提供版本：v3.0.1-alpha.72+
- 该接口所需权限：无。
- 该接口功能描述：通过下载地址导入插件V3资源包。接口会异步执行导入工作流（下载并上传插件包、发布资源包、更新资源包可见性），并返回工作流ID，可通过 PackageImportResult 接口查询导入结果。

### URL

POST /api/v3/package/workflow/import/v3/plugin

### 输入参数

| 参数名称     | 参数类型 | 必选 | 描述                |
| ------------ | -------- | ---- | ------------------- |
| filename     | string   | 是   | 插件包文件名        |
| download_url | string   | 是   | 插件包下载地址      |
| md5          | string   | 是   | 插件包文件MD5校验值 |

### 调用示例

```json
{
  "filename": "bkmonitorbeat-2.1.3.tgz",
  "download_url": "https://example.com/packages/bkmonitorbeat-2.1.3.tgz",
  "md5": "5d41402abc4b2a76b9719d911017c592"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "workflow_id": "workflow-abc123def456"
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

| 参数名称    | 参数类型 | 描述                                                  |
| ----------- | -------- | ----------------------------------------------------- |
| workflow_id | string   | 工作流ID，可用于 PackageImportResult 接口查询导入结果 |
