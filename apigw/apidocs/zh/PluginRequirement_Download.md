### 描述

- 该接口提供版本：v3.0.1-alpha.119+。
- 该接口所需权限：plugin_view（查看插件）。
- 该接口功能描述：按 UUID 下载依赖文件正文，成功响应为文件二进制内容。

### URL

GET /api/v3/plugin/requirement/download

### 输入参数

| 参数名称            | 参数类型 | 必选 | 描述                                  |
| ------------------- | -------- | ---- | ------------------------------------- |
| requirement_file_id | string   | 是   | 依赖文件记录的 UUID，通过查询参数传递 |

**参数说明**：

- 接口对记录归属插件检查 `plugin_view` 权限。
- `requirement_file_id` 必须是合法 UUID，否则请求校验失败。
- 成功时返回文件二进制正文；失败时返回标准 JSON 错误响应。

### 调用示例

```text
GET /api/v3/plugin/requirement/download?requirement_file_id=0b6f0f18-6f1f-4c88-9d3f-2a4f1f3d1a11
```

### 响应示例

成功时响应状态为 200，响应体为文件二进制正文（非 JSON 包裹）：

```text
<binary file content>
```

失败时返回标准 JSON 错误响应：

```json
{
  "code": 500,
  "message": "failed to download plugin requirement",
  "request_id": "req-20260324-000015",
  "error": {
    "system": "backend",
    "message": "failed to get requirement file",
    "details": []
  }
}
```

### 响应参数说明

- 成功：HTTP 200，响应体为文件二进制正文，无 JSON 包裹。
- 失败：返回标准错误响应，字段如下。

| 参数名称   | 参数类型 | 描述                   |
| ---------- | -------- | ---------------------- |
| code       | int32    | 状态码，0表示成功      |
| message    | string   | 请求信息               |
| request_id | string   | 请求ID                 |
| error      | object   | 错误信息，成功时为null |

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
