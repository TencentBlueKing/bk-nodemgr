### 描述

- 该接口提供版本：v3.0.1-alpha.119+。
- 该接口所需权限：plugin_operate（操作插件）。
- 该接口功能描述：替换依赖文件的名称与正文，保留记录 ID、归属插件和适用范围。

### URL

POST /api/v3/plugin/requirement/update

### 输入参数

| 参数名称            | 参数类型 | 必选 | 描述                               |
| ------------------- | -------- | ---- | ---------------------------------- |
| requirement_file_id | string   | 是   | 依赖文件记录的 UUID                |
| name                | string   | 是   | 替换后的依赖业务名称               |
| content_base64      | string   | 是   | 替换后的文件正文，标准 Base64 编码 |
| target_file_name    | string   | 是   | 替换后的下发文件名                 |

**参数说明**：

- 接口对记录归属插件检查 `plugin_operate` 权限。
- `requirement_file_id` 必须是合法 UUID。
- `name` 不能为空、不能包含控制字符；同一租户、同一插件内唯一。
- `target_file_name` 不能为空，不能为 `.` 或 `..`，不能包含 `/ \ : < > " | ? *` 等字符，不能以空格或 `.` 结尾，且不能使用保留设备名。
- 本接口只替换名称、正文和下发文件名；归属插件与适用范围保持不变，如需调整适用范围请使用 update_scope 接口。

### 调用示例

```json
{
  "requirement_file_id": "0b6f0f18-6f1f-4c88-9d3f-2a4f1f3d1a11",
  "name": "agent-config-v2",
  "content_base64": "IyBhZ2VudCBjb25maWcgdjIK",
  "target_file_name": "agent.conf"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260324-000011",
  "error": null,
  "data": {
    "file": {
      "requirement_file_id": "0b6f0f18-6f1f-4c88-9d3f-2a4f1f3d1a11",
      "name": "agent-config-v2",
      "file_name": "9f3c2a5e-1b7d-4e2f-8a6c-0d1e2f3a4b5c",
      "md5": "6f1ed002ab5595859014ebf0951522d9",
      "size": 18,
      "updated_at": 1774339260,
      "operator": "admin",
      "os": ["linux"],
      "arch": ["amd64", "arm64"],
      "version": [],
      "plugin_name": "bk-monitor-agent",
      "target_file_name": "agent.conf"
    }
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

| 参数名称 | 参数类型 | 描述         |
| -------- | -------- | ------------ |
| file     | object   | 依赖文件记录 |

#### data.file

| 参数名称            | 参数类型     | 描述                                |
| ------------------- | ------------ | ----------------------------------- |
| requirement_file_id | string       | 依赖文件记录的稳定 UUID             |
| name                | string       | 依赖业务名称，同租户同插件内唯一    |
| file_name           | string       | 文件在存储中的 FileGroup 对象名称（UUID），区别于下发文件名              |
| md5                 | string       | 文件正文的 MD5                      |
| size                | int64        | 解码后的文件大小，单位为字节        |
| updated_at          | int64        | 最近更新时间，Unix 时间戳，单位为秒 |
| operator            | string       | 最近操作人                          |
| os                  | string array | 适用操作系统；空数组表示不限        |
| arch                | string array | 适用 CPU 架构；空数组表示不限       |
| version             | string array | 适用插件包版本；空数组表示不限      |
| plugin_name         | string       | 依赖文件归属的插件名称              |
| target_file_name    | string       | 下发到目标主机的文件名，不含目录    |

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
