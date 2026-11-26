### 描述

- 该接口提供版本：v3.0.1-alpha.119+。
- 该接口所需权限：plugin_view（查看插件）。
- 该接口功能描述：按元数据分页列举当前租户的依赖文件，用于管理文件记录。

### URL

POST /api/v3/plugin/requirement/list

### 输入参数

| 参数名称   | 参数类型 | 必选 | 描述         |
| ---------- | -------- | ---- | ------------ |
| page       | object   | 是   | 分页参数     |
| conditions | object   | 否   | 精确匹配条件 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                          |
| -------- | -------- | ---- | ----------------------------- |
| offset   | int64    | 否   | 分页偏移量，从 0 开始，默认 0 |
| limit    | int64    | 是   | 每页条数，必须大于 0          |
| sort     | string   | 否   | 可选排序表达式                |

#### conditions.exact_include

| 参数名称            | 参数类型     | 必选 | 描述                     |
| ------------------- | ------------ | ---- | ------------------------ |
| requirement_file_id | string array | 否   | 按记录 UUID 精确匹配     |
| name                | string array | 否   | 按依赖业务名称精确匹配   |
| os                  | string array | 否   | 按适用操作系统精确匹配   |
| arch                | string array | 否   | 按适用 CPU 架构精确匹配  |
| version             | string array | 否   | 按适用插件包版本精确匹配 |
| plugin_name         | string array | 否   | 按归属插件精确匹配       |
| target_file_name    | string array | 否   | 按下发文件名精确匹配     |

**参数说明**：

- `page` 必填；`offset` 不能为负数，`limit` 必须大于 0。
- 按记录中保存的值精确筛选；同一字段内的取值为或，不同字段之间为且。
- 适用范围为空的记录不会展开为具体的操作系统、架构或版本，因此不会匹配具体的 `os`/`arch`/`version` 条件。
- 接口对调用方检查 `plugin_view` 权限，并将结果收敛到其授权可见的插件范围：仅返回归属插件对调用方可见的记录；若授权范围内没有可见插件，返回空列表与总数 0。

### 调用示例

```json
{
  "page": {
    "offset": 0,
    "limit": 20
  },
  "conditions": {
    "exact_include": {
      "plugin_name": ["bk-monitor-agent"],
      "os": ["linux"]
    }
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260324-000014",
  "error": null,
  "data": {
    "total": 1,
    "items": [
      {
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

| 参数名称 | 参数类型     | 描述                     |
| -------- | ------------ | ------------------------ |
| total    | int64        | 匹配条件的记录总数       |
| items    | object array | 当前页的依赖文件记录列表 |

#### data.items[n]

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
