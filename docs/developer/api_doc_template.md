# API 文档模板

**版本**: v1.0
**适用项目**: bk-nodemgr
**最后更新**: 2026-01

## 使用说明

本模板用于编写 bk-nodemgr 项目的 API 参考文档。请根据实际 API 特点选择合适的模板变体。

**前置准备**:
- **文件命名**: 基于 swagger 定义的 `operationId`
  - 在 `docs/api/swagger/` 目录查找对应 API 的 operationId
  - 示例: `/api/v3/node/agent/install` → `NodeAgent_NodeAgentInstall.md`
- **版本号**: 查询项目 git 标签获取最新正式版本号
  - 新增接口使用当前版本（如 `v3.0.1+`）
  - 功能变动更新版本号（如 `v3.1.0+`）

**模板变体**:
- [基础模板](#基础模板) - 适用于简单的 CRUD 操作
- [复杂查询模板](#复杂查询模板) - 适用于带分页和过滤的列表查询
- [文件操作模板](#文件操作模板) - 适用于文件上传/下载操作

**填写原则**:
1. 所有 `[占位符]` 需要替换为实际内容
2. `{可选}` 标记的章节根据实际需要保留或删除
3. 参数类型使用 protobuf 类型（string, int64, bool等）
4. 示例使用真实的JSON格式，保持格式正确

---

## 基础模板

适用于简单的增删改查操作，如：创建节点、删除插件、更新配置等。

```markdown
### 描述

- 该接口提供版本：v3.0.0+。
- 该接口所需权限：[权限名称，如"节点管理-节点操作"，或留空]。
- 该接口功能描述：[一句话说明接口功能，如"批量启动节点Agent"]。

### URL

[HTTP_METHOD] /api/v3/[service]/[resource]/[method]

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| [param1] | [type] | 是 | [参数描述] |
| [param2] | [type] | 否 | [参数描述] |

{可选：如果有路径参数，在描述中说明}

### 调用示例

\```json
{
  "[param1]": "[value]",
  "[param2]": "[value]"
}
\```

### 响应示例

\```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "[field1]": "[value]",
    "[field2]": "[value]"
  }
}
\```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0表示成功 |
| message | string | 请求信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| [field1] | [type] | [字段描述] |
| [field2] | [type] | [字段描述] |

```

---

## 复杂查询模板

适用于列表查询，支持分页、排序、过滤等高级功能。

```markdown
### 描述

- 该接口提供版本：v3.0.0+。
- 该接口所需权限：[权限名称或留空]。
- 该接口功能描述：[如"查询节点列表，支持分页和条件过滤"]。

### URL

GET /api/v3/[service]/[resource]/list

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
| [field1] | [type] | [可查询字段描述] |
| [field2] | [type] | [可查询字段描述] |
| created_at | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at | string | 更新时间，标准格式：2006-01-02T15:04:05Z |

### 调用示例

查询[具体条件描述]的数据。

\```json
{
  "page": {
    "count": true,
    "start": 0,
    "limit": 10,
    "sort": "[sort_field]",
    "order": "DESC"
  },
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "[field_name]",
        "op": "eq",
        "value": "[value]"
      }
    ]
  }
}
\```

### 响应示例

\```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "count": 100,
    "details": [
      {
        "[field1]": "[value]",
        "[field2]": "[value]"
      }
    ]
  }
}
\```

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
| [field1] | [type] | [字段描述] |
| [field2] | [type] | [字段描述] |
| created_at | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at | string | 更新时间，标准格式：2006-01-02T15:04:05Z |

```

---

## 文件操作模板

适用于文件上传、下载等操作。

```markdown
### 描述

- 该接口提供版本：v3.0.0+。
- 该接口所需权限：[权限名称或留空]。
- 该接口功能描述：[如"上传插件包文件"]。

### URL

[POST|GET] /api/v3/file/[resource]/[method]

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| file | file | 是 | 上传的文件 |
| [param1] | [type] | 否 | [其他参数描述] |

{可选：文件相关的特殊说明，如大小限制、格式要求}

**文件要求**:
- 最大文件大小：[如 100MB]
- 支持的文件格式：[如 .tar.gz, .zip]

### 调用示例

{说明：文件上传通常使用 multipart/form-data}

\```
POST /api/v3/file/upload HTTP/1.1
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="file"; filename="plugin.tar.gz"

[binary data]
------WebKitFormBoundary--
\```

### 响应示例

\```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "file_id": "file-123456",
    "file_name": "plugin.tar.gz",
    "file_size": 1048576,
    "upload_time": "2024-01-01T12:00:00Z"
  }
}
\```

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
| upload_time | string | 上传时间 |

```

---

## 通用规范

### 参数类型

| 类型 | 说明 | 示例 |
|------|------|------|
| string | 字符串 | "example" |
| int32 | 32位整数 | 100 |
| int64 | 64位整数 | 1234567890 |
| uint32 | 32位无符号整数 | 100 |
| uint64 | 64位无符号整数 | 1234567890 |
| bool | 布尔值 | true, false |
| float64 | 64位浮点数 | 3.14 |
| array | 数组 | ["item1", "item2"] |
| object | 对象 | {"key": "value"} |

### HTTP 方法

| 方法 | 用途 |
|------|------|
| GET | 查询资源 |
| POST | 创建资源或执行操作 |
| PUT | 更新资源 |
| DELETE | 删除资源 |
| PATCH | 部分更新资源 |

### URL 路径规范

标准格式：`/api/{version}/{service}/{resource}/{method}`

**示例**:
- `/api/v3/backend/node/list` - 后端服务的节点列表查询
- `/api/v3/application/plugin/get` - 应用服务的插件查询
- `/api/v3/file/upload` - 文件上传

### 响应状态码

| code | 含义 |
|------|------|
| 0 | 成功 |
| 非0 | 错误，具体错误信息在 message 字段 |

### 时间格式

统一使用 ISO 8601 格式：`2006-01-02T15:04:05Z`

### 命名约定

#### 字段命名
- 使用小写字母和下划线：`node_id`, `created_at`
- 时间字段使用 `_at` 后缀：`created_at`, `updated_at`
- ID 字段使用 `_id` 后缀：`node_id`, `plugin_id`
- 布尔字段使用 `is_` 或 `enable_` 前缀：`is_enabled`, `enable_auto_start`

#### URL 命名
- 使用小写字母和下划线
- 资源使用名词：`/node`, `/plugin`
- 操作使用动词：`/list`, `/create`, `/update`

---

## 填写清单

完成一个 API 文档需要填写：

- [ ] 接口版本号
- [ ] 接口所需权限（可暂时留空）
- [ ] 功能描述（一句话）
- [ ] HTTP 方法和 URL 路径
- [ ] 输入参数表格（参数名、类型、必选、描述）
- [ ] 嵌套对象的参数说明（如有）
- [ ] 调用示例（JSON 格式）
- [ ] 响应示例（JSON 格式）
- [ ] 响应参数说明表格
- [ ] 嵌套响应对象的说明（如有）

---

## 常见问题

**Q: 如何处理枚举类型？**
A: 在"描述"列中说明，如："状态（枚举值：running、stopped、error）"

**Q: 权限字段应该填什么？**
A: 可以暂时留空或填写"待补充"，后续统一完善。

**Q: 响应的 data 字段为空怎么办？**
A: 如果没有返回数据，data 可以为 null 或空对象 {}，在响应参数说明中注明。

**Q: 如何描述数组类型的参数？**
A: 类型写为 "array" 或 "string array"，并在下方单独列出数组元素的结构。

---

**模板版本**: v1.0
**维护人**: bk-nodemgr
**更新日期**: 2026-01
