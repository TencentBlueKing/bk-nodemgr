---
name: code-review
description: 审查代码和 commit，检查语法错误、逻辑一致性和代码质量。当用户说 review、检查错误、审查 commit 时使用。
allowed-tools:
  - Read
  - Grep
  - Glob
  - SemanticSearch
  - ReadLints
  - Shell
---

# Code Review Skill

## 审查流程

按以下顺序进行代码审查：

### 1. 确定审查范围

首先确定需要审查的文件：
```bash
git status -uno  # 查看已暂存的提交文件
```

识别文件类型和模块，判断适用的规范和检查重点。

### 2. 语法错误检查（最高优先级）

对于 Go 项目：
```bash
go build ./path/to/package/...
go vet ./path/to/package/...
```

使用 ReadLints 检查文件的 linter 错误。

### 3. 项目规范检查

**根据文件类型和模块，参考相应的规范文档：**

#### API 接口开发
- 文件路径包含 `proto/` 或 API 相关代码
- **参考文档**: `docs/api/API接口开发流程.md`
- **检查项**: 
  - API 接口命名和设计规范
  - 请求/响应结构定义
  - 错误码和错误处理

#### Proto 文件
- 文件后缀为 `.proto`
- **参考文档**: `proto/README.md`
- **检查项**:
  - 命名规范（service、message、field）
  - 字段编号管理
  - 注释完整性

#### 开发规范
- **参考文档**: `docs/developer/README.md`
- **检查项**:
  - 代码风格和命名规范
  - 错误处理模式
  - 日志记录规范

### 4. 逻辑一致性检查

**模块特定规范：**
- `internal/backend/dpmgr/executor.go` → [patterns/dpmgr-executor.md](patterns/dpmgr-executor.md)
- 其他模块规范参见 [patterns/README.md](patterns/README.md)

**一般检查项：**
- 函数签名是否与相似函数一致
- 错误处理模式是否统一
- 日志格式是否一致
- 数据结构构建是否符合模式

### 5. 代码质量检查

**核心规范：**
- 遵循 `.golangci.yml` 规则
- 遵循 `.cursor/rules/conv.mdc` 数据转换规则
- 参考 `AGENTS.md` 和 `CLAUDE.md` 项目规范

**检查清单：**
- [ ] 错误处理是否完整（使用 `fmt.Errorf` 和 `%w`）
- [ ] 是否遵循项目代码风格（导入别名、命名、注释）
- [ ] 是否使用了 `pkg/runtime/conv` 进行数据转换
- [ ] 日志记录是否使用结构化日志
- [ ] 公共函数是否有完整注释

### 6. 潜在问题检查

关注以下问题：
- **性能问题**: 不必要的循环、重复计算、资源浪费
- **安全问题**: SQL 注入、XSS、敏感信息泄露
- **并发问题**: 竞态条件、死锁、goroutine 泄露
- **资源泄露**: 文件句柄、数据库连接、内存泄露

## 报告格式

审查报告应按优先级分类问题，每个问题提供具体位置和修复建议：

### 严重问题（必须修复）
- **语法错误**: 编译错误、go vet 警告、linter 错误
- **安全问题**: 安全漏洞、敏感信息泄露
- **逻辑错误**: 会导致功能失败或数据错误的问题

示例：
```
❌ 严重：internal/backend/dpmgr/executor.go:123
   - 问题：未检查错误返回值
   - 影响：可能导致程序崩溃
   - 修复：添加 if err != nil { return err }
```

### 重要问题（强烈建议修复）
- **项目规范违反**: 不符合 `.golangci.yml`、`AGENTS.md` 的规范
- **逻辑一致性问题**: 与相似函数不一致的实现
- **性能问题**: 明显的性能瓶颈
- **资源泄露**: 文件句柄、连接等未正确关闭

示例：
```
⚠️ 重要：internal/backend/dpmgr/executor.go:156
   - 问题：日志格式不符合项目规范
   - 规范：应使用两行格式（参考 patterns/dpmgr-executor.md）
   - 当前：logger.G.Sys().Info("message")
   - 应改为：
     logger.G.Sys().With("key", value).
         Info("message")
```

### 建议改进（可选）
- **代码优化**: 可读性、可维护性改进
- **注释补充**: 缺少注释但不影响功能
- **最佳实践**: 更好的实现方式

示例：
```
💡 建议：internal/backend/dpmgr/executor.go:178
   - 建议：使用 conv.MapKeyToSlice() 代替手动循环
   - 参考：.cursor/rules/conv.mdc
   - 好处：代码更简洁，自动排序
```

### 规范文档引用

如涉及规范问题，引用相关文档路径：
- 项目规范：`AGENTS.md`、`CLAUDE.md`
- API 规范：`docs/api/API接口开发流程.md`
- Proto 规范：`proto/README.md`
- 开发规范：`docs/developer/README.md`
- 概念文档：`docs/concepts/README.md`
- 代码转换：`.cursor/rules/conv.mdc`
- Linter 配置：`.golangci.yml`

### 审查总结

最后提供总体评估：
- **审查结果**: 通过 / 需要修复 / 不建议合并
- **必须修复的问题数**: X 个严重问题
- **建议修复的问题数**: Y 个重要问题
- **改进建议数**: Z 个建议
- **整体评价**: 简要总结代码质量和主要问题

## 审查 Commit

审查 commit 时：
1. 使用 `git show <commit-hash>` 查看变更
2. 识别变更的文件类型和模块
3. 读取修改的文件（完整内容，了解上下文）
4. 执行编译和静态检查
5. 对比相似代码的模式
6. 参考相关规范文档
7. 生成审查报告

## 审查步骤详解

### 步骤 1: 识别文件和模块

根据文件路径判断：
- `proto/*.proto` → Proto 文件规范
- `internal/backend/dpmgr/` → dpmgr-executor 模式
- `pkg/dao/` → 数据访问层规范
- `pkg/rest/` → API 框架规范
- API 相关代码 → API 开发流程

### 步骤 2: 加载相关规范

根据识别的模块，读取相应的规范文档：
- 模块特定规范：`patterns/*.md`
- 项目规范：`AGENTS.md`、`CLAUDE.md`
- 领域规范：`docs/api/`、`docs/developer/`、`proto/README.md`

### 步骤 3: 执行静态检查

```bash
# 编译检查
go build ./path/to/package/...

# 静态分析
go vet ./path/to/package/...

# Linter 检查
# 使用 ReadLints 工具检查修改的文件
```

### 步骤 4: 对比相似代码

- 使用 Grep 查找相似函数
- 使用 SemanticSearch 查找相关模式
- 比对函数签名、错误处理、日志格式
- 检查数据结构构建模式

### 步骤 5: 检查项目规范

按检查清单逐项验证：
- [ ] 错误处理完整性
- [ ] 代码风格一致性
- [ ] 数据转换使用 conv 包
- [ ] 日志记录结构化
- [ ] 注释完整性
- [ ] 特定模块规范遵循情况

### 步骤 6: 生成报告

按"报告格式"章节的要求，生成结构化的审查报告。

## 项目规范速查

### 数据转换规范（.cursor/rules/conv.mdc）

**优先使用 `pkg/runtime/conv` 包：**

**基本类型转换：**
- 转换为 int64: `conv.ToInt64()` 或 `conv.ToInt64Default()`
- 转换为 string: `conv.ToString()` 或 `conv.ToStringDefault()`
- 字符串转布尔值: `conv.StringToBool()`
- 数字转布尔值: `conv.NumberToBool()`

**结构体和映射转换：**
- map 转 struct: `conv.MapToStruct()`
- struct 转 map: `conv.StructToMap()` 或 `conv.StructToMapIgnoreError()`

**切片操作：**
- 切片去重: `conv.SliceUnique()`
- 切片转换: `conv.SliceToSlice()` 或 `conv.SliceToSliceWithError()`
- 切片转 map: `conv.SliceToMap()`
- map 值转切片: `conv.MapValueToSlice()`
- **map 键转切片（自动排序）**: `conv.MapKeyToSlice()`

**工具函数：**
- 检查值是否为空: `conv.IsEmpty()`

**使用原则：**
1. ✅ 优先使用：当 conv 包提供了合适的转换函数时
2. ✅ 可以使用标准库：如果标准库更合适、更高效
3. ❌ 不要强行使用：如果需要特殊的转换逻辑
4. ⚠️ 注意错误处理：处理转换函数返回的错误

### 错误处理规范

**标准模式：** 使用 `fmt.Errorf("...: %w", err)` 包装错误

示例：
```go
// ✅ 正确：包装错误并提供上下文
if err := someFunc(); err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// ❌ 错误：直接返回错误，丢失上下文
if err := someFunc(); err != nil {
    return err
}

// ❌ 错误：忽略错误
_ = someFunc()
```

### 日志记录规范

**标准格式（两行）：**
```go
logger.G.Sys().With("key1", value1).With("key2", value2).
    Info("message")
```

**检查项：**
- 使用结构化日志（`With` 方法添加上下文）
- `With` 和 `Info`/`Error` 等方法分开在不同行
- 日志消息清晰描述操作结果

示例：
```go
// ✅ 正确
logger.G.Sys().With("workflow-id", workflowID).With("task-count", len(tasks)).
    Info("successful to execute change action")

// ❌ 错误：单行格式
logger.G.Sys().With("workflow-id", workflowID).Info("successful to execute change action")
```

### 命名规范

根据 `.golangci.yml`：
- 错误类型以 `Error` 结尾
- 错误变量以 `Err` 开头（如 `ErrNotFound`）
- 接口命名简洁明了
- 使用统一的导入别名

### 注释规范

- 公共函数和类型必须有注释
- 注释以句号结尾
- 使用中文注释（项目主要使用中文）
- 注释应解释"为什么"而非"做什么"

### API 开发规范

参考 `docs/api/API接口开发流程.md`：
- 遵循 REST API 设计原则
- 使用统一的错误响应格式（`pkg/rest/errf`）
- 正确映射 HTTP 状态码
- 实现认证和授权检查

### Proto 文件规范

参考 `proto/README.md`：
- 遵循命名规范（service、message、field）
- 合理管理字段编号
- 添加完整的注释
- 正确使用 proto3 特性

## 扩展模块规范

本 skill 支持针对不同模块添加特定审查规范。所有模块规范文档位于 `patterns/` 目录。

**添加新模块规范：**
1. 在 `patterns/` 目录创建模块文档（如 `patterns/workflow-operation.md`）
2. 使用 `patterns/_TEMPLATE.md` 作为模板
3. 在本文件的"逻辑一致性检查"部分添加引用
4. 参考 `patterns/README.md` 了解文档结构建议

**当前支持的模块：**
- Deploy Policy Manager Executor - `patterns/dpmgr-executor.md`

## 与项目工作流集成

### 与 .cursor/agents 的关系

本 skill 与 `.cursor/agents/code-reviewer.md` 协同工作：
- `.cursor/agents/code-reviewer.md` 提供简化的审查工作流和检查清单
- 本 skill 提供详细的审查步骤、规范速查和模块特定规范
- 两者共享相同的规范文档引用体系

### 与 OpenSpec 工作流的关系

参考 `openspec/AGENTS.md`：
- 审查变更提案（`/openspec-proposal`）时，关注设计文档的完整性
- 审查变更实施（`/openspec-apply`）时，确保实现符合提案要求
- 审查归档前（`/openspec-archive`），验证所有任务已完成

### PR 审查建议

在 PR 审查时：
1. 检查 PR 标题是否符合格式：`<type>: <description> --issue=#<number>`
2. 验证实现是否与相关的 OpenSpec 变更提案一致
3. 确认所有 linter 错误已修复
4. 检查是否遵循项目规范和模块特定规范
5. 提供结构化的审查报告

## 审查技巧

### 高效使用工具

- **Grep**: 快速查找相似函数、错误处理模式
- **SemanticSearch**: 理解模块整体设计和代码模式
- **ReadLints**: 优先检查 linter 错误，避免重复工作
- **Shell**: 执行编译和静态检查，获取准确的错误信息

### 审查优先级

1. **语法错误** - 阻止编译的问题，最高优先级
2. **安全问题** - 可能导致安全漏洞的问题
3. **逻辑错误** - 影响功能正确性的问题
4. **规范违反** - 不符合项目规范的问题
5. **代码质量** - 可读性、可维护性改进

### 常见审查场景

#### 场景 1: 审查新增的 API 接口

1. 读取 `docs/api/API接口开发流程.md`
2. 检查 proto 文件定义（如有）
3. 验证错误码定义和 HTTP 状态码映射
4. 检查认证和授权实现
5. 验证请求/响应结构

#### 场景 2: 审查 executor 模式代码

1. 读取 `patterns/dpmgr-executor.md`
2. 识别函数系列（Agent/Plugin/PluginPkg）
3. 对比同系列的其他函数
4. 检查日志格式、错误消息、数据结构构建
5. 验证 Manager 调用是否正确

#### 场景 3: 审查数据转换代码

1. 检查是否使用了 `pkg/runtime/conv` 包
2. 如果手动实现转换，评估是否必要
3. 验证错误处理是否完整
4. 建议使用 conv 包的合适函数

## 注意事项

- **保持客观**: 基于规范和最佳实践提供反馈，避免主观判断
- **提供上下文**: 引用具体的规范文档和行号
- **建设性反馈**: 不仅指出问题，还提供解决方案
- **优先级明确**: 区分必须修复和建议改进
- **简洁明了**: 避免冗长的解释，直接指出问题和修复方法
