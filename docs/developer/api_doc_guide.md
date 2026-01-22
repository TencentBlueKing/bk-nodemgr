# API 文档编写指南

本指南说明如何使用 `api_doc_template.md` 模板编写 bk-nodemgr 项目的 API 参考文档。

## 目录

1. [快速开始](#快速开始)
2. [模板选择](#模板选择)
3. [填写步骤](#填写步骤)
4. [示例演示](#示例演示)
5. [最佳实践](#最佳实践)
6. [常见错误](#常见错误)

---

## 快速开始

### 第零步：准备工作

**确定文件名**:

文档文件名基于 swagger 定义中的 `operationId`，不是手动命名。

- 在 `docs/api/swagger/` 目录下查找对应 API 的 swagger.json 文件
- 找到对应 URL 和 HTTP 方法的 `operationId`
- 直接使用 operationId 作为文件名（如 `NodeAgent_NodeAgentInstall.md`）

示例：`/api/v3/node/agent/install` 的 operationId 是 `NodeAgent_NodeAgentInstall`，文档名为 `NodeAgent_NodeAgentInstall.md`

**确定版本号**:

版本号遵循以下规则：

- **新增接口**: 使用当前最新正式版本号（如 `v3.0.1+`）
  - 查询项目 git 标签获取最新正式版本，排除 alpha、beta、test 等预发布版本
- **功能变动**: 更新到变动时的版本号（如 `v3.1.0+`）
  - 功能变动包括：新增参数、修改行为、调整响应结构
- **不确定时**: 咨询团队或查看 git 历史记录确认接口首次出现的版本

### 第一步：选择模板

根据 API 的类型选择合适的模板变体：

| API 类型 | 使用模板 | 示例 |
|---------|---------|------|
| 简单 CRUD | 基础模板 | 创建节点、删除插件 |
| 列表查询 | 复杂查询模板 | 查询节点列表、查询工作流列表 |
| 文件操作 | 文件操作模板 | 上传插件包、下载日志文件 |

### 第二步：复制模板

从 `api_doc_template.md` 中复制对应的模板到新文件。

### 第三步：替换占位符

将所有 `[占位符]` 替换为实际内容。

### 第四步：删除可选内容

删除标记为 `{可选}` 但不需要的章节。

### 第五步：填写版本号和文件名

- 使用第零步获取的版本号填写"该接口提供版本"字段
- 使用第零步获取的文件名保存文档

---

## 模板选择

### 基础模板

**适用场景**:
- 单个资源的增删改查
- 简单的操作请求
- 参数结构扁平（无复杂嵌套）

**典型 API**:
- `POST /api/v3/backend/node/create` - 创建节点
- `DELETE /api/v3/backend/plugin/delete` - 删除插件
- `POST /api/v3/backend/process/restart` - 重启进程

**特点**:
- 参数表格简单明了
- 一个调用示例
- 一个响应示例

---

### 复杂查询模板

**适用场景**:
- 列表查询 API
- 需要分页功能
- 需要条件过滤
- 需要排序

**典型 API**:
- `GET /api/v3/backend/node/list` - 查询节点列表
- `GET /api/v3/backend/plugin/list` - 查询插件列表
- `GET /api/v3/backend/workflow/list` - 查询工作流列表

**特点**:
- 包含 page 和 filter 参数
- 详细的操作符说明
- 可查询字段列表
- 支持复杂的查询条件

---

### 文件操作模板

**适用场景**:
- 文件上传
- 文件下载
- 文件传输

**典型 API**:
- `POST /api/v3/file/upload` - 上传文件
- `GET /api/v3/file/download` - 下载文件
- `POST /api/v3/file/transfer` - 文件传输

**特点**:
- 包含文件要求说明
- multipart/form-data 示例
- 文件信息的响应结构

---

## 填写步骤

### 1. 填写描述部分

```markdown
### 描述

- 该接口提供版本：v3.0.0+。
- 该接口所需权限：节点管理-节点查看。
- 该接口功能描述：查询指定业务下的节点列表，支持分页和条件过滤。
```

**注意事项**:
- 版本号格式：v{major}.{minor}.{patch}+
- 权限可暂时留空，后续统一填写
- 功能描述要简洁明确，一句话说清楚

### 2. 填写 URL

```markdown
### URL

GET /api/v3/backend/node/list
```

**URL 构成**:
```
/api/{version}/{service}/{resource}/{method}
```

**各部分说明**:
- `version`: 当前统一使用 v3
- `service`: backend, application, file 之一
- `resource`: 资源名称（node, plugin, process等）
- `method`: 操作方法（list, get, create, update, delete等）

### 3. 填写输入参数

#### 简单参数

```markdown
### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| node_id | string | 是 | 节点ID |
| force | bool | 否 | 是否强制执行，默认false |
```

#### 复杂参数（嵌套对象）

```markdown
### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| node_info | object | 是 | 节点信息 |

#### node_info

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| node_id | string | 是 | 节点ID |
| node_name | string | 是 | 节点名称 |
| tags | string array | 否 | 节点标签 |
```

**参数类型对照表**（基于 protobuf）:

| Protobuf 类型 | 文档类型 |
|--------------|---------|
| string | string |
| int32 | int32 |
| int64 | int64 |
| uint32 | uint32 |
| uint64 | uint64 |
| bool | bool |
| double | float64 |
| repeated string | string array |
| repeated int64 | int64 array |
| message | object |

### 4. 填写调用示例

```markdown
### 调用示例

\```json
{
  "node_id": "node-123456",
  "force": true
}
\```
```

**注意事项**:
- 使用真实的示例值，不要用 `xxx` 或 `...`
- JSON 格式必须正确（注意逗号、引号）
- 字符串值加引号，数字和布尔值不加引号
- GET 请求无 body 时可以留空：`\```json\n\```\`

### 5. 填写响应示例

```markdown
### 响应示例

\```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "node_id": "node-123456",
    "status": "running"
  }
}
\```
```

**标准响应结构**:
- `code`: 0 表示成功，非 0 表示错误
- `message`: 请求信息，成功时通常为 "ok"
- `data`: 响应数据，根据接口不同而不同

### 6. 填写响应参数说明

```markdown
### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0表示成功 |
| message | string | 请求信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| node_id | string | 节点ID |
| status | string | 节点状态 |
```

---

## 示例演示

完整示例请参考：`api_doc_example.md`

### 简单 CRUD 示例

见 `api_doc_example.md` 中的"示例 1：批量重启节点"

### 复杂查询示例

见 `api_doc_example.md` 中的"示例 2：查询插件列表"

### 文件操作示例

见 `api_doc_example.md` 中的"示例 3：上传插件包"

---

## 最佳实践

### 1. 保持一致性

- 所有文档使用相同的模板
- 参数命名风格统一
- 示例格式统一

### 2. 详细但简洁

- 描述要准确但不冗长
- 示例要真实但不复杂
- 避免不必要的重复

### 3. 用户视角

- 从 API 使用者的角度编写
- 提供常见使用场景的示例
- 说明重要的注意事项

### 4. 及时更新

- API 变更时同步更新文档
- 版本号与实际版本对应
- 标注废弃的 API

### 5. 参数说明清晰

- 枚举值要列举完整
- 范围限制要明确
- 格式要求要具体
- 默认值要说明

**好的参数描述示例**:
```
| status | string | 是 | 节点状态（枚举值：running、stopped、error） |
| port | int32 | 否 | 端口号，范围1-65535，默认8080 |
| created_at | string | 否 | 创建时间，格式：2006-01-02T15:04:05Z |
```

### 6. 示例要真实

**不好的示例**:
```json
{
  "id": "xxx",
  "name": "xxx"
}
```

**好的示例**:
```json
{
  "id": "node-abc123",
  "name": "node-prod-01"
}
```

### 7. 处理复杂场景

对于复杂的 filter 查询，提供多个示例：

```markdown
### 调用示例

#### 示例 1：查询单个字段

查询状态为 running 的节点。

\```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "status",
        "op": "eq",
        "value": "running"
      }
    ]
  }
}
\```

#### 示例 2：组合查询

查询状态为 running 且创建时间在指定范围内的节点。

\```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "status",
        "op": "eq",
        "value": "running"
      },
      {
        "field": "created_at",
        "op": "gte",
        "value": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
\```
```

---

## 常见错误

### 错误 1：JSON 格式错误

**错误示例**:
```json
{
  "name": "test",    // 不要写注释
  "id": 123,
}                    // 最后一个逗号要删除
```

**正确示例**:
```json
{
  "name": "test",
  "id": 123
}
```

### 错误 2：参数类型不一致

**错误示例**:
```markdown
| id | string | 是 | 节点ID |

调用示例：
{
  "id": 123    // 示例中是数字，但类型标注是 string
}
```

**正确示例**:
```markdown
| id | string | 是 | 节点ID |

调用示例：
{
  "id": "node-123"
}
```

### 错误 3：遗漏必选参数

**错误示例**:
```markdown
输入参数：
| name | string | 是 | 名称 |
| desc | string | 是 | 描述 |

调用示例：
{
  "name": "test"
  // 缺少 desc 参数
}
```

### 错误 4：响应参数说明不完整

**错误示例**:
```markdown
响应示例：
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "123",
    "name": "test",
    "created_at": "2024-01-01"
  }
}

响应参数说明：
| code | int32 | 状态码 |
| message | string | 请求信息 |
// 缺少 data 内部字段的说明
```

**正确示例**: 应该包含完整的 data 字段说明（见模板）。

### 错误 5：枚举值不明确

**错误示例**:
```markdown
| status | string | 是 | 状态 |
```

**正确示例**:
```markdown
| status | string | 是 | 状态（枚举值：running、stopped、error） |
```

---

## 文档组织

建议的文档目录结构：

```
docs/api-docs/
├── README.md                    # API 文档索引
├── backend/
│   ├── node_agent.md           # 节点 Agent 相关 API
│   ├── node_proxy.md           # 节点 Proxy 相关 API
│   ├── plugin.md               # 插件管理相关 API
│   ├── process.md              # 进程管理相关 API
│   └── workflow.md             # 工作流相关 API
├── application/
│   └── ...                     # 应用服务 API
└── file/
    ├── upload.md               # 文件上传 API
    ├── download.md             # 文件下载 API
    └── transfer.md             # 文件传输 API
```

---

## 检查清单

文档完成后，使用此清单检查：

- [ ] 版本号正确
- [ ] URL 路径符合规范
- [ ] 所有参数都有类型和描述
- [ ] 必选/可选标注正确
- [ ] 枚举值已列举
- [ ] 调用示例 JSON 格式正确
- [ ] 响应示例 JSON 格式正确
- [ ] 响应参数说明完整
- [ ] 嵌套对象都有详细说明
- [ ] 时间格式统一使用 ISO 8601
- [ ] 无拼写错误
- [ ] 表格对齐整齐

---

## 获取帮助

- 查看 `api_doc_template.md` 了解模板详情
- 查看 `api_doc_example.md` 参考完整示例
- 查看现有 proto 文件获取准确的参数定义
- 如有疑问，联系团队负责人

---

**文档版本**: v1.0
**最后更新**: 2026-01
