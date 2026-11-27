### 描述

- 该接口提供版本：v3.0.1-alpha.121+。
- 该接口所需权限：plugin_view（查看插件）。
- 该接口功能描述：按 UUID 查询依赖文件记录，响应只包含元数据，不含文件正文。

### URL

POST /api/v3/plugin/requirement/get

### 输入参数

| 参数名称            | 参数类型 | 必选 | 描述                |
| ------------------- | -------- | ---- | ------------------- |
| requirement_file_id | string   | 是   | 依赖文件记录的 UUID |

**参数说明**：

- 接口对记录归属插件检查 `plugin_view` 权限。
- `requirement_file_id` 必须是合法 UUID，否则请求校验失败。
- 本接口只返回元数据；如需下载文件正文请使用 download 接口。

### 调用示例

```json
{
  "requirement_file_id": "0b6f0f18-6f1f-4c88-9d3f-2a4f1f3d1a11"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260324-000013",
  "error": null,
  "data": {
    "file": {
      "requirement_file_id": "0b6f0f18-6f1f-4c88-9d3f-2a4f1f3d1a11",
      "name": "agent-config-v2",
      "file_name": "9f3c2a5e-1b7d-4e2f-8a6c-0d1e2f3a4b5c",
      "md5": "6f1ed002ab5595859014ebf0951522d9",
      "size": 18,
      "updated_at": 1774339320,
      "operator": "admin",
      "os": ["linux", "windows"],
      "arch": ["amd64"],
      "version": ["1.2.0", "1.3.0"],
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
