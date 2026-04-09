---
name: api-doc
description: Writes standardized Markdown reference documentation for bk-nodemgr Proto APIs.
allowed-tools:
  # MCP 工具（优先使用）
  - mcp__sequential-thinking__sequentialthinking  # 结构化思考
  - mcp__serena__find_symbol                      # 查找符号定义
  - mcp__serena__search_for_pattern               # 搜索代码模式
  - mcp__serena__get_symbols_overview             # 获取文件符号概览
  - mcp__serena__find_referencing_symbols         # 查找引用
  - mcp__serena__read_memory                      # 读取项目记忆
  - mcp__serena__list_memories                    # 列出可用记忆
  # 传统工具（补充）
  - Read                 # 读取文件
  - Glob                 # 文件查找
  - Grep                 # 内容搜索
  - Write                # 写入文档
---

# API 文档编写 Skill

为 bk-nodemgr 项目的 Proto API 生成标准化参考文档。

## MCP 工具优先原则

优先使用 MCP 工具：
- **serena 工具族**: 符号级代码分析（`find_symbol`、`search_for_pattern`）
- **read_memory**: 读取项目记忆

传统工具（Read/Glob/Grep/Write）作为补充。

## 快速开始

执行 7 步流程：

1. **确定范围** - 使用 `sequential-thinking` 明确要编写的 API 和模板类型
2. **收集资料** - 使用 Serena 理解代码结构, 阅读 proto、类型定义、概念文档
3. **选择模板** - 根据 API 类型选择合适模板
4. **编写文档** - 按模板结构填写内容
5. **补充信息** - 使用 Serena 添加概述、场景、FAQ（可选）
6. **交叉验证** - 使用 Serena 验证字段语义（**关键**）
7. **审查检查** - 使用 `sequential-thinking` 验证准确性和完整性

**详细流程**: [references/workflow.md](references/workflow.md)

## 输出规范

- **一个 API 一个文档**
- **中英文分离**: `apigw/apidocs/zh/` 和 `apigw/apidocs/en/`
- **命名规范**: 基于 swagger 的 `operationId`
  - 运行 `scripts/get_doc_filename.py <url> [method]` 获取正确文件名
  - 示例: `scripts/get_doc_filename.py /api/v3/node/agent/install POST`
  - 输出: `NodeAgent_NodeAgentInstall.md`

## 版本号管理

文档"该接口提供版本"字段规则：

**获取当前版本**: 运行 `scripts/get_version.sh` 获取最新正式版本号

**填写规则**:
- 新增接口 → 使用当前版本 (如 `v3.0.1+`)
- 功能变动 → 更新到变动版本 (如 `v3.1.0+`)
- 重大变更可单独说明 (如 "v3.1.0+ 新增 xxx 参数")

**不确定时**: 询问用户确认接口首次出现或变动的版本

## 权限格式规范

文档"该接口所需权限"字段必须使用标准格式：

**格式**: `action_id（中文名）`
- 单个权限: `networkunit_view（查看管控单元）`
- 多个权限: `agent_operate（操作Agent）、networkunit_use_for_agent（使用网络单元部署Agent）`
- 无权限: `无`

**查找方法**:
1. 在 handler 代码中搜索 `h.authorizer.Check` 或 `auth.ActionXxx`
2. 在 `internal/backend/auth/action.go` 中查找对应常量定义
3. 在 `ActionDisplayName` 函数中查找中文名

**英文版格式**: `action_id (English Name)`
- 示例: `networkunit_view (View Network Unit)`

## 关键文件路径

| 文件类型 | 路径 |
|---------|------|
| Proto 定义 | `proto/` |
| 类型定义 | `pkg/types/*.go` |
| 概念文档 | `docs/concepts/` |
| 模板 | `docs/developer/api_doc_template.md` |
| 示例 | `docs/developer/api_doc_example*.md` |
| 中文输出 | `apigw/apidocs/zh/` |
| 英文输出 | `apigw/apidocs/en/` |

## 模板选择

| API 类型 | 模板 | 典型场景 |
|---------|------|---------|
| 简单 CRUD | 基础模板 | create, update, delete, get |
| 列表查询 | 复杂查询模板 | list (带分页过滤) |
| 文件操作 | 文件操作模板 | upload, download |

## 资料收集顺序

1. **Proto 定义** (必读) → Service/RPC、Request/Response、字段类型
2. **类型定义** (必读) → 业务模型、枚举常量、验证规则
3. **概念文档** (推荐) → 业务含义、使用场景
4. **实现代码** (可选) → 参数校验、业务逻辑

## 交叉验证（关键）

**Proto 只是起点，不是终点。** 必须用 Serena 验证字段语义。

| 字段类型 | 风险 | 验证方法 |
|---------|------|---------|
| `repeated string` | 可能是预定义枚举 | `search_for_pattern` 查找 Go 类型 |
| `string` 枚举 | 枚举值可能不完整 | 在 `pkg/types/` 查找 `const` |
| `google.protobuf.Struct` | 多态结构 | 找实际 Go 类型定义 |

**典型案例**: `proxy_tags` 在 Proto 中看似自定义标签，实际是只有 4 个有效值的预定义枚举。

**详细验证流程和案例**: [references/workflow.md](references/workflow.md) 阶段 6

## 不明确时的处理

**核心原则**: 宁可少写、标记待确认，也不要写错误的业务描述。

使用告警格式：
```
⚠️ **待确认**: [问题描述]
   - 来源: [proto/类型定义]
   - 需要确认: [具体内容]
```

## 关键技巧

| 技巧 | 参考位置 |
|------|---------|
| 处理多态字段 | `docs/developer/api_doc_template.md` |
| 处理嵌套对象 | `docs/developer/api_doc_template.md` |
| 枚举值说明格式 | `docs/developer/api_doc_template.md` |

## 审查检查清单

- [ ] 文件名正确（运行 `scripts/get_doc_filename.py` 确认）
- [ ] 版本号正确（运行 `scripts/get_version.sh` 确认）
- [ ] 字段名称/类型与 proto 一致
- [ ] 枚举值完整正确
- [ ] URL 路径和 HTTP 方法正确
- [ ] JSON 示例格式正确
- [ ] 嵌套对象和多态参数结构完整

## 参考文档

- **详细流程**: [references/workflow.md](references/workflow.md)
- **模板**: `docs/developer/api_doc_template.md`
- **示例**: `docs/developer/api_doc_example*.md`
