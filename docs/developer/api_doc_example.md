# API 文档示例

本文档提供三个完整的 API 文档示例，展示如何使用模板编写实际的 API 文档。

**注意**:
- 文档文件名应基于 swagger 的 operationId（如 `NodeAgent_NodeAgentInstall.md`）
- 版本号应根据接口实际情况填写，新增接口使用当前最新版本（示例使用 v3.0.1+）

---

## 示例 1：批量重启节点 (基础模板)

### 描述

- 该接口提供版本：v3.0.1+
- 该接口所需权限：节点管理-节点操作。
- 该接口功能描述：批量重启指定的节点Agent。

### URL

POST /api/v3/backend/node/batch_restart

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| node_ids | string array | 是 | 节点ID列表 |
| force | bool | 否 | 是否强制重启，默认false |
| timeout | int32 | 否 | 超时时间（秒），默认300 |

### 调用示例

```json
{
  "node_ids": [
    "node-abc123",
    "node-def456"
  ],
  "force": true,
  "timeout": 600
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "task_id": "task-xyz789",
    "success_count": 2,
    "failed_count": 0
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0表示成功 |
| message | string | 请求信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| task_id | string | 任务ID，可用于查询任务状态 |
| success_count | int32 | 成功重启的节点数 |
| failed_count | int32 | 重启失败的节点数 |

---

## 示例 2：查询插件列表 (复杂查询模板)

### 描述

- 该接口提供版本：v3.1.0+。
- 该接口所需权限：插件管理-插件查看。
- 该接口功能描述：查询插件列表，支持分页和条件过滤。

### URL

GET /api/v3/backend/plugin/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| page | object | 否 | 分页配置 |
| filter | object | 否 | 查询过滤条件 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| count | bool | 是 | 是否返回总记录条数 |
| start | uint32 | 否 | 记录开始位置，起始值为0 |
| limit | uint32 | 否 | 每页限制条数，最大500 |
| sort | string | 否 | 排序字段 |
| order | string | 否 | 排序顺序（ASC、DESC） |

#### filter

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| op | string | 是 | 操作符（枚举值：and、or） |
| rules | array | 是 | 过滤规则，最多设置5个rules |

#### rules[n]

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| field | string | 是 | 查询条件字段名称 |
| op | string | 是 | 操作符（枚举值：eq、neq、gt、gte、lt、lte、in、nin、cs、cis） |
| value | 可变类型 | 是 | 查询条件值 |

##### 操作符说明

| 操作符 | 描述 | value支持的数据类型 |
|-------|------|-------------------|
| eq | 等于 | boolean, numeric, string |
| neq | 不等于 | boolean, numeric, string |
| gt | 大于 | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| gte | 大于等于 | numeric，时间类型为字符串 |
| lt | 小于 | numeric，时间类型为字符串 |
| lte | 小于等于 | numeric，时间类型为字符串 |
| in | 在给定的数组范围中 | boolean, numeric, string（数组最多100个元素） |
| nin | 不在给定的数组范围中 | boolean, numeric, string（数组最多100个元素） |
| cs | 模糊查询，区分大小写 | string |
| cis | 模糊查询，不区分大小写 | string |

#### 查询参数说明

接口调用者可以根据以下参数设置查询规则：

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| plugin_id | string | 插件ID |
| plugin_name | string | 插件名称 |
| plugin_type | string | 插件类型（枚举值：official、custom） |
| status | string | 插件状态（枚举值：enabled、disabled、error） |
| version | string | 插件版本 |
| creator | string | 创建者 |
| created_at | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at | string | 更新时间，标准格式：2006-01-02T15:04:05Z |

### 调用示例

#### 示例 1：查询所有启用状态的插件

```json
{
  "page": {
    "count": true,
    "start": 0,
    "limit": 10,
    "sort": "created_at",
    "order": "DESC"
  },
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "status",
        "op": "eq",
        "value": "enabled"
      }
    ]
  }
}
```

#### 示例 2：查询名称包含"monitor"的官方插件

```json
{
  "page": {
    "count": true,
    "limit": 20
  },
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "plugin_name",
        "op": "cis",
        "value": "monitor"
      },
      {
        "field": "plugin_type",
        "op": "eq",
        "value": "official"
      }
    ]
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "count": 25,
    "details": [
      {
        "plugin_id": "plugin-001",
        "plugin_name": "bk-monitor-agent",
        "plugin_type": "official",
        "status": "enabled",
        "version": "2.1.3",
        "description": "蓝鲸监控采集器",
        "creator": "admin",
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-03-20T14:25:00Z"
      },
      {
        "plugin_id": "plugin-002",
        "plugin_name": "bk-log-collector",
        "plugin_type": "official",
        "status": "enabled",
        "version": "1.5.0",
        "description": "日志采集插件",
        "creator": "admin",
        "created_at": "2024-01-10T08:00:00Z",
        "updated_at": "2024-02-28T16:45:00Z"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0表示成功 |
| message | string | 请求信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| count | uint64 | 当前规则能匹配到的总记录条数 |
| details | array | 查询返回的数据 |

#### data.details[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| plugin_id | string | 插件ID |
| plugin_name | string | 插件名称 |
| plugin_type | string | 插件类型（枚举值：official、custom） |
| status | string | 插件状态（枚举值：enabled、disabled、error） |
| version | string | 插件版本 |
| description | string | 插件描述 |
| creator | string | 创建者 |
| created_at | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at | string | 更新时间，标准格式：2006-01-02T15:04:05Z |

---

## 示例 3：上传插件包 (文件操作模板)

### 描述

- 该接口提供版本：v3.0.1+
- 该接口所需权限：插件管理-插件上传。
- 该接口功能描述：上传插件包文件到服务器。

### URL

POST /api/v3/file/plugin/upload

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| file | file | 是 | 插件包文件 |
| plugin_name | string | 是 | 插件名称 |
| version | string | 是 | 插件版本 |
| description | string | 否 | 插件描述 |

**文件要求**:
- 最大文件大小：500MB
- 支持的文件格式：.tar.gz, .tgz

### 调用示例

```
POST /api/v3/file/plugin/upload HTTP/1.1
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW

------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="file"; filename="bk-monitor-agent-2.1.3.tar.gz"
Content-Type: application/gzip

[binary data]
------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="plugin_name"

bk-monitor-agent
------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="version"

2.1.3
------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="description"

蓝鲸监控采集器v2.1.3版本
------WebKitFormBoundary7MA4YWxkTrZu0gW--
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "file_id": "file-abc123def456",
    "file_name": "bk-monitor-agent-2.1.3.tar.gz",
    "file_size": 52428800,
    "file_md5": "5d41402abc4b2a76b9719d911017c592",
    "upload_time": "2024-03-25T15:30:45Z",
    "storage_path": "/data/plugins/bk-monitor-agent-2.1.3.tar.gz"
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0表示成功 |
| message | string | 请求信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| file_id | string | 文件ID |
| file_name | string | 文件名 |
| file_size | int64 | 文件大小（字节） |
| file_md5 | string | 文件MD5校验值 |
| upload_time | string | 上传时间，标准格式：2006-01-02T15:04:05Z |
| storage_path | string | 存储路径 |

---

## 使用这些示例

这些示例展示了：

1. **基础模板的应用**（示例1）:
   - 简单的参数结构
   - 数组类型参数的表示
   - 清晰的响应数据说明

2. **复杂查询模板的应用**（示例2）:
   - 完整的 page 和 filter 参数
   - 多个调用示例展示不同场景
   - 详细的查询参数说明
   - 嵌套的响应数据结构

3. **文件操作模板的应用**（示例3）:
   - multipart/form-data 格式
   - 文件要求说明
   - 文件相关的响应字段

编写新的 API 文档时，可以参考这些示例，根据实际 API 的特点进行调整。

---

**示例版本**: v1.0
**最后更新**: 2026-01
