# MCP 工具优先使用指南

本文档详细介绍代码审查过程中各种工具的使用方法，**强调 MCP 工具优先**策略。

---

## 工具优先级原则

在代码审查过程中，工具选择遵循以下优先级：

```
优先级 1：核心 MCP 工具（sequential-thinking）
    ↓
优先级 2：语义分析工具（serena 工具族）
    ↓
优先级 3：文档工具（context7）
    ↓
优先级 4：传统工具（补充和降级方案）
```

**核心理念**：
- ✅ 优先使用语义级工具而非文本级工具
- ✅ 优先使用结构化概览而非全文读取
- ✅ 优先使用 MCP 工具，传统工具作为补充
- ✅ 保持灵活性，根据实际需求选择最合适的工具

---

## 第一部分：MCP 工具详解（优先使用）

### 1. sequential-thinking - 结构化思考工具

#### 用途和优势

**用途**：
- 复杂决策的结构化分析
- 多步骤问题的拆解和规划
- 审查报告的问题分类和优先级判断

**优势**：
- 提供可追溯的推理过程
- 确保审查的系统性和完整性
- 避免遗漏重要检查项
- 生成结构化的审查报告

#### 使用场景

**场景 1：审查开始时的策略规划**

在开始审查之前，使用 sequential-thinking 分析：
- 变更的规模和复杂度
- 需要关注的重点领域
- 审查的优先级和顺序
- 预估的时间和资源

```
示例：
1. 分析 git diff 输出，识别变更文件类型
2. 判断是基础审查、完整审查还是 PR 审查
3. 确定需要加载的规范文档
4. 规划工具使用策略
```

**场景 2：遇到复杂问题时的分析**

当发现潜在问题时，使用 sequential-thinking：
- 分析问题的根本原因
- 评估问题的严重程度
- 探索可能的解决方案
- 判断是否需要进一步调查

```
示例：
1. 发现错误处理不一致
2. 使用 sequential-thinking 分析：
   - 不一致的具体表现
   - 可能的影响范围
   - 正确的处理模式
   - 建议的修改方案
```

**场景 3：生成报告时的问题分类**

在生成最终报告前，使用 sequential-thinking：
- 将发现的问题进行分类（严重/重要/建议）
- 确定问题的优先级
- 组织报告结构
- 确保报告的逻辑性和完整性

```
示例：
1. 收集所有发现的问题
2. 使用 sequential-thinking 分类：
   - 阻塞性问题（编译错误、严重 bug）
   - 重要问题（规范不符、潜在风险）
   - 建议性问题（代码优化、可读性改进）
3. 生成结构化报告
```

#### 最佳实践

1. **审查开始必用**：每次审查都应该先使用 sequential-thinking 规划策略
2. **复杂决策必用**：遇到需要权衡的问题时使用
3. **报告生成必用**：确保报告的结构化和完整性
4. **记录思考过程**：保持推理过程的可追溯性

---

### 2. serena 工具族 - 语义级代码分析

serena 提供了一套强大的语义级代码分析工具，能够理解代码结构和符号关系，显著提高审查效率。

#### 2.1 serena.get_symbols_overview - 快速结构概览

**用途**：快速了解文件的符号结构，避免全文读取

**替代工具**：Read（仅在需要完整内容时才使用）

**使用场景**：

**场景 1：审查开始时快速了解文件结构**

```
目标：在不读取完整文件的情况下，了解文件定义了哪些类、函数、方法

示例调用：
serena.get_symbols_overview(
    relative_path="internal/backend/dpmgr/executor.go",
    depth=1  # 包含顶层符号的子符号
)

返回信息：
- 所有顶层函数的名称和签名
- 所有类/结构体的名称和字段
- 所有方法的名称（如果 depth=1）
- 符号位置（行号）
```

**场景 2：批量了解多个文件的结构**

```
目标：在审查多个文件时，快速建立整体印象

流程：
1. 使用 git diff 获取变更文件列表
2. 对每个文件调用 get_symbols_overview
3. 识别主要的新增/修改符号
4. 确定需要深入审查的部分
```

**优势对比**：

| 维度 | serena.get_symbols_overview | Read |
|------|----------------------------|------|
| Token 消耗 | ~500 tokens/文件 | ~2,000 tokens/文件 |
| 获取信息 | 结构化符号列表 | 完整文本内容 |
| 适用场景 | 快速了解结构 | 需要查看实现细节 |
| 推荐优先级 | ✅ 优先使用 | ⚠️ 降级方案 |

**最佳实践**：
1. 总是先使用 get_symbols_overview，再决定是否需要 Read
2. 使用 depth 参数控制返回详细程度（0=顶层，1=包含子符号）
3. 结合 find_symbol 深入查看特定符号的实现

**何时降级到 Read**：
- 需要查看完整的注释和文档
- 需要理解导入依赖关系
- 需要查看符号之间的完整上下文
- get_symbols_overview 返回的信息不足以做出判断

---

#### 2.2 serena.find_symbol - 精确符号查找

**用途**：基于符号名称精确查找函数、类、方法的定义

**替代工具**：Grep（语义级搜索更准确，Grep 作为降级方案）

**使用场景**：

**场景 1：查找特定函数的实现**

```
目标：找到某个函数的完整定义，包括签名和实现

示例调用：
serena.find_symbol(
    name_path_pattern="CreateDeployment",
    relative_path="internal/backend/dpmgr/",
    include_body=true,  # 包含函数体
    depth=0  # 只返回函数本身
)

返回信息：
- 函数的完整定义（包括签名和函数体）
- 函数的位置（文件和行号）
- 函数的文档注释
```

**场景 2：查找相似函数进行一致性检查**

```
目标：查找同模块中的相似函数，对比实现模式

示例调用：
serena.find_symbol(
    name_path_pattern="Create",
    substring_matching=true,  # 匹配所有以 Create 开头的函数
    relative_path="internal/backend/dpmgr/",
    include_body=true
)

用途：
- 对比错误处理模式
- 对比日志记录格式
- 对比函数签名设计
- 对比数据结构构建方式
```

**场景 3：查找类的所有方法**

```
目标：了解某个类的完整方法列表

示例调用：
serena.find_symbol(
    name_path_pattern="DeploymentManager",
    relative_path="internal/backend/dpmgr/",
    depth=1,  # 包含所有方法
    include_body=false  # 只要方法签名，不要实现
)

返回信息：
- 类的定义
- 所有方法的名称和签名
- 方法的文档注释
```

**参数说明**：

| 参数 | 说明 | 建议 |
|------|------|------|
| name_path_pattern | 符号名称或路径模式 | 使用简单名称即可，如 "CreateDeployment" |
| relative_path | 限制搜索范围 | 总是指定，提高搜索速度 |
| include_body | 是否包含完整实现 | 一致性检查时使用 true，快速了解时使用 false |
| depth | 子符号深度 | 0=只返回符号本身，1=包含子符号（如方法） |
| substring_matching | 子串匹配 | 查找相似函数时使用 true |

**优势对比**：

| 维度 | serena.find_symbol | Grep |
|------|-------------------|------|
| 搜索精度 | 语义级，精确匹配符号 | 文本级，可能误匹配注释或字符串 |
| 结果结构 | 结构化符号信息 | 文本行 |
| 上下文理解 | 理解符号作用域和层次 | 仅匹配文本模式 |
| Token 消耗 | ~300 tokens/查询 | ~500 tokens/查询 |
| 推荐优先级 | ✅ 优先使用 | ⚠️ 降级方案 |

**最佳实践**：
1. 总是指定 relative_path 限制搜索范围
2. 使用 substring_matching=true 查找相似函数系列
3. 一致性检查时使用 include_body=true，快速浏览时使用 false
4. 结合 depth 参数控制返回详细程度

**何时降级到 Grep**：
- 需要搜索跨语言的模式（serena 主要支持主流语言）
- 需要搜索注释或文档字符串中的内容
- 需要使用复杂的正则表达式
- serena 无法理解你的查询意图

---

#### 2.3 serena.find_referencing_symbols - 依赖关系分析

**用途**：查找引用某个符号的所有位置，理解依赖关系和影响范围

**替代工具**：无（传统工具无法提供此能力）

**使用场景**：

**场景 1：分析函数修改的影响范围**

```
目标：了解修改某个函数会影响哪些调用方

示例调用：
serena.find_referencing_symbols(
    name_path="CreateDeployment",
    relative_path="internal/backend/dpmgr/executor.go"
)

返回信息：
- 所有调用该函数的位置
- 调用方的函数名称和上下文
- 调用代码片段
```

**场景 2：理解数据结构的使用模式**

```
目标：了解某个结构体在代码库中如何被使用

示例调用：
serena.find_referencing_symbols(
    name_path="DeploymentConfig",
    relative_path="internal/backend/dpmgr/types.go"
)

用途：
- 发现所有构建该结构体的位置
- 验证字段使用的一致性
- 评估结构体修改的影响
```

**场景 3：检查接口实现**

```
目标：查找实现某个接口的所有类型

示例调用：
serena.find_referencing_symbols(
    name_path="Executor",
    relative_path="internal/backend/dpmgr/interface.go"
)

用途：
- 验证所有实现的一致性
- 评估接口修改的影响范围
```

**最佳实践**：
1. 修改函数签名前，先使用此工具了解影响范围
2. 重构代码时，使用此工具确保所有引用都已更新
3. 结合 include_info=true 获取引用位置的详细上下文

**独特价值**：
- ✅ 无可替代：传统工具无法提供语义级的引用分析
- ✅ 精确度高：理解代码结构，避免误报
- ✅ 影响分析：评估代码修改的波及范围

---

#### 2.4 serena.search_for_pattern - 语义级模式搜索

**用途**：基于正则表达式进行语义级的代码模式搜索

**替代工具**：Grep（在 serena 无法满足时降级使用）

**使用场景**：

**场景 1：查找特定的代码模式**

```
目标：查找所有使用特定模式的代码片段

示例调用：
serena.search_for_pattern(
    substring_pattern="fmt\\.Errorf.*%w",
    relative_path="internal/backend/dpmgr/",
    restrict_search_to_code_files=true
)

用途：
- 检查错误包装模式的一致性
- 查找特定 API 的使用方式
```

**场景 2：查找日志记录模式**

```
目标：验证日志记录的一致性

示例调用：
serena.search_for_pattern(
    substring_pattern="logger\\.G\\.Sys\\(\\)\\.With",
    relative_path="internal/",
    context_lines_after=2  # 显示后续 2 行
)

用途：
- 检查日志格式一致性
- 验证结构化字段使用
```

**场景 3：查找安全相关模式**

```
目标：查找可能的安全问题

示例调用：
serena.search_for_pattern(
    substring_pattern="exec\\.Command|os\\.Exec",
    relative_path=".",
    context_lines_before=1,
    context_lines_after=3
)

用途：
- 发现潜在的命令注入风险
- 检查输入验证
```

**参数说明**：

| 参数 | 说明 | 建议 |
|------|------|------|
| substring_pattern | 正则表达式模式 | 使用 Python re 模块语法 |
| relative_path | 搜索路径 | 总是指定，避免全局搜索 |
| restrict_search_to_code_files | 只搜索代码文件 | 通常使用 true |
| context_lines_before/after | 上下文行数 | 根据需要调整，默认 0 |
| paths_include_glob | 包含文件模式 | 进一步限制搜索范围 |
| paths_exclude_glob | 排除文件模式 | 排除测试文件等 |

**优势对比**：

| 维度 | serena.search_for_pattern | Grep |
|------|--------------------------|------|
| 搜索范围 | 可灵活限制（代码文件、glob） | 基于路径和文件类型 |
| 模式匹配 | 支持正则表达式 | 支持正则表达式 |
| 结果结构 | 结构化输出 | 文本行 |
| 上下文控制 | 精确控制上下文行数 | 使用 -A/-B/-C 参数 |
| 推荐优先级 | ✅ 优先使用 | ⚠️ 降级方案 |

**最佳实践**：
1. 总是使用 relative_path 限制搜索范围
2. 使用 restrict_search_to_code_files=true 避免搜索非代码文件
3. 使用 paths_exclude_glob 排除测试文件（如 "*_test.go"）
4. 根据需要调整上下文行数

**何时降级到 Grep**：
- 需要更灵活的输出格式控制
- 需要搜索非代码文件（如文档、配置）
- 需要使用 Grep 特有的功能（如 -o 只输出匹配部分）
- serena 的性能不满足需求

---

### 3. context7 - 第三方文档查询

**用途**：获取最新的第三方库文档、API 规范和最佳实践

**使用场景**：

**场景 1：验证 API 使用正确性**

```
目标：确认某个 API 的正确使用方式

示例：
- 验证 Go 标准库函数的签名
- 检查第三方库的最新 API
- 确认框架的推荐用法
```

**场景 2：查阅最新的最佳实践**

```
目标：确保代码符合最新的标准和规范

示例：
- 查询 Go 错误处理的最佳实践
- 了解并发编程的推荐模式
- 验证性能优化的建议
```

**场景 3：了解框架文档**

```
目标：查阅框架的官方文档

示例：
- 查询 gRPC 的使用规范
- 了解 ORM 框架的最佳实践
- 验证中间件的正确用法
```

**优势**：
- ✅ 确保规范准确性，减少过时信息
- ✅ 获取官方文档，避免误导性信息
- ✅ 了解最新版本的变更和建议

**最佳实践**：
1. 当不确定 API 使用是否正确时使用
2. 优先查询官方文档源
3. 结合项目内规范（使用 Read）进行综合判断

**何时使用**：
- ✅ 审查使用了不熟悉的第三方库
- ✅ 需要验证 API 使用的正确性
- ✅ 需要确认最新的最佳实践
- ❌ 不要用于查询项目内部规范（使用 Read）

---

## 第二部分：传统工具（补充和降级方案）

传统工具在某些场景下仍然必要，或作为 MCP 工具的降级方案。

### 1. ReadLints - Linter 错误检查（无可替代）

**用途**：检查文件的 linter 错误和警告

**特点**：✅ 无可替代，必须使用

**使用场景**：
- 审查开始时首先执行（最高优先级）
- 验证代码符合 `.golangci.yml` 配置
- 发现明显的代码规范问题

**最佳实践**：
```
1. 在审查开始时立即使用
2. 只对变更的文件运行 ReadLints
3. 发现 linter 错误立即报告（阻塞性问题）
```

**示例**：
```
ReadLints file_path=/path/to/file.go
```

---

### 2. Shell/Bash - 命令执行（部分无可替代）

**用途**：执行编译、测试、Git 操作等命令

**特点**：✅ 部分场景无可替代（编译、测试、Git）

**使用场景**：

**场景 1：编译检查（无可替代）**

```bash
# 编译检查
go build ./internal/backend/dpmgr/...

# 静态分析
go vet ./internal/backend/dpmgr/...
```

**场景 2：测试执行（无可替代）**

```bash
# 运行测试
go test ./internal/backend/dpmgr/...

# 运行特定测试
go test -run TestCreateDeployment ./internal/backend/dpmgr/
```

**场景 3：Git 操作（无可替代）**

```bash
# 查看暂存的变更
git diff --staged

# 查看变更的文件
git status -uno

# 查看特定 commit
git show <commit-hash>

# 查看 commit 列表
git log --oneline -10
```

**最佳实践**：
1. 总是指定具体的包路径，避免全局检查浪费时间
2. 优先运行快速检查（build、vet），再运行慢检查（test）
3. Git 操作用于理解变更范围和历史

---

### 3. Read - 读取完整文件（降级方案）

**用途**：读取文件的完整内容

**特点**：⚠️ 降级方案，优先使用 serena.get_symbols_overview

**何时使用**：
- ✅ 需要查看完整的注释和文档
- ✅ 需要理解导入依赖关系
- ✅ 需要查看文件的完整上下文
- ✅ serena.get_symbols_overview 返回的信息不足

**何时避免**：
- ❌ 只是想快速了解文件结构（使用 serena.get_symbols_overview）
- ❌ 只是想查找特定符号（使用 serena.find_symbol）
- ❌ 文件很大但只关心部分内容（使用 serena 工具）

**Token 消耗对比**：
- Read 全文：~2,000 tokens/文件
- serena.get_symbols_overview：~500 tokens/文件
- **节省：75%**

**最佳实践**：
1. 先使用 serena.get_symbols_overview，再决定是否需要 Read
2. 如果文件很大，考虑使用 offset 和 limit 参数分段读取
3. 只读取确实需要审查的文件

**示例**：
```
Read file_path=/path/to/file.go
```

---

### 4. Grep - 文本模式搜索（降级方案）

**用途**：在代码库中搜索文本模式

**特点**：⚠️ 降级方案，优先使用 serena.find_symbol 或 serena.search_for_pattern

**何时使用**：
- ✅ 需要搜索注释或文档字符串
- ✅ 需要使用复杂的正则表达式
- ✅ serena 工具无法满足需求
- ✅ 需要 Grep 特有的功能

**何时避免**：
- ❌ 查找特定函数或类（使用 serena.find_symbol）
- ❌ 查找代码模式（使用 serena.search_for_pattern）

**最佳实践**：
1. 总是指定 path 参数限制搜索范围
2. 使用 type 参数限制文件类型（如 type="go"）
3. 使用 output_mode 控制输出格式
4. 结合 -A/-B/-C 参数查看上下文

**示例**：

```bash
# 查找相似函数
Grep pattern="func Create.*Deployment" path=internal/backend/dpmgr/

# 查找错误处理模式
Grep pattern="fmt\\.Errorf.*%w" path=internal/backend/ output_mode="content"

# 查找日志使用
Grep pattern="logger\\.G\\.Sys\\(\\)\\.With" path=internal/ -A=2
```

---

### 5. Glob - 文件查找

**用途**：根据文件名模式查找文件

**特点**：✅ 无可替代（文件查找）

**使用场景**：
- 查找特定类型的文件
- 定位测试文件
- 查找配置文件

**最佳实践**：
1. 使用通配符提高匹配灵活性
2. 限制搜索路径以加快速度

**示例**：

```bash
# 查找所有 executor 相关文件
Glob pattern="**/executor*.go"

# 查找测试文件
Glob pattern="**/*_test.go" path=internal/backend/dpmgr/
```

---

## 第三部分：工具组合策略和最佳实践

### 工具选择决策树

```
审查任务
    │
    ├─ 需要规划审查策略？
    │   └─ 是 → sequential-thinking
    │
    ├─ 需要语法检查？
    │   └─ 是 → ReadLints + Shell (go build/vet)
    │
    ├─ 需要理解代码结构？
    │   └─ 是 → serena.get_symbols_overview
    │           （如果信息不足 → 降级到 Read）
    │
    ├─ 需要查找特定符号？
    │   └─ 是 → serena.find_symbol
    │           （如果失败 → 降级到 Grep）
    │
    ├─ 需要理解依赖关系？
    │   └─ 是 → serena.find_referencing_symbols
    │
    ├─ 需要查阅第三方文档？
    │   └─ 是 → context7
    │
    ├─ 需要模式搜索？
    │   └─ 是 → serena.search_for_pattern
    │           （如果失败 → 降级到 Grep）
    │
    ├─ 需要完整文件内容？
    │   └─ 是 → Read
    │
    └─ 需要生成报告？
        └─ 是 → sequential-thinking
```

---

### 不同审查场景的工具组合

#### 场景 1：基础审查（单文件小修改）

**目标**：快速检查，10分钟内完成

**工具组合**：
```
1. sequential-thinking（30秒）
   - 快速规划审查策略

2. serena.get_symbols_overview（30秒）
   - 了解文件结构

3. ReadLints + Shell（2分钟）
   - 语法检查

4. serena.find_symbol（3分钟）
   - 查找相似代码，检查一致性

5. sequential-thinking（3分钟）
   - 生成简洁报告
```

**预期 Token 消耗**：~5,000 tokens

**效率优化**：
- ✅ 使用 serena.get_symbols_overview 而非 Read
- ✅ 使用 serena.find_symbol 精确查找
- ✅ 避免全局搜索，总是指定路径

---

#### 场景 2：完整审查（新功能或复杂变更）

**目标**：全面审查，30-60分钟完成

**工具组合**：
```
1. sequential-thinking（2分钟）
   - 分析变更范围和策略

2. Shell: git diff（1分钟）
   - 了解变更文件

3. serena.get_symbols_overview（5分钟）
   - 批量了解文件结构

4. ReadLints + Shell（5分钟）
   - 语法检查

5. serena.find_symbol（10分钟）
   - 查找需要审查的符号
   - 深度分析实现

6. serena.find_referencing_symbols（10分钟）
   - 理解影响范围
   - 检查依赖关系

7. context7（5分钟，可选）
   - 查阅相关规范

8. sequential-thinking（10分钟）
   - 生成详细报告
```

**预期 Token 消耗**：~20,000 tokens

**效率优化**：
- ✅ 并行执行语法检查
- ✅ 使用 serena 工具避免全文读取
- ✅ 渐进式审查，先快速检查再深入

---

#### 场景 3：Commit/PR 审查

**目标**：全面审查 + 提案一致性检查

**工具组合**：
```
1. sequential-thinking（3分钟）
   - 规划审查策略

2. Shell: git show（2分钟）
   - 查看变更

3. 执行完整审查流程（40分钟）
   - 参考场景 2

4. OpenSpec 提案一致性检查（10分钟）
   - 验证实现符合提案要求
   - 检查所有任务已完成

5. sequential-thinking（5分钟）
   - 生成 PR 审查报告
```

**预期 Token 消耗**：~25,000 tokens

---

### 效率优化建议

#### 优化 1：并行执行

某些检查可以并行执行：

```
同时执行：
- ReadLints file1.go
- ReadLints file2.go
- ReadLints file3.go

同时执行：
- serena.get_symbols_overview file1.go
- serena.get_symbols_overview file2.go
```

#### 优化 2：限制搜索范围

总是尽可能限制工具的搜索范围：

```
✅ 推荐：
serena.find_symbol(
    name_path_pattern="CreateDeployment",
    relative_path="internal/backend/dpmgr/"
)

❌ 避免：
serena.find_symbol(
    name_path_pattern="CreateDeployment",
    relative_path="."
)
```

#### 优化 3：渐进式审查

从快到慢，从全局到局部：

```
1. ReadLints (最快) - 发现明显问题
2. go build (快) - 编译检查
3. go vet (快) - 静态分析
4. serena.get_symbols_overview (快) - 结构概览
5. serena.find_symbol (中速) - 符号查找
6. serena.search_for_pattern (中速) - 模式搜索
7. Read (慢) - 全文读取（降级方案）
```

#### 优化 4：智能降级

MCP 工具无法满足时才降级：

```
serena.find_symbol
    ↓ 失败
Grep（降级方案）

serena.get_symbols_overview
    ↓ 信息不足
Read（降级方案）
```

---

### 常见错误和陷阱

#### 错误 1：过度使用 Read

**问题**：读取大量不相关的文件，浪费 Token

**解决**：
- ✅ 优先使用 serena.get_symbols_overview 了解结构
- ✅ 使用 serena.find_symbol 精确查找
- ✅ 只在确实需要完整内容时才使用 Read

**对比**：
```
❌ 错误做法：
Read file1.go
Read file2.go
Read file3.go
总 Token：~6,000

✅ 正确做法：
serena.get_symbols_overview file1.go
serena.get_symbols_overview file2.go
serena.get_symbols_overview file3.go
总 Token：~1,500
节省：75%
```

---

#### 错误 2：忘记限制搜索范围

**问题**：在整个代码库搜索，浪费时间和资源

**解决**：
- ✅ 总是指定 relative_path 参数
- ✅ 根据 git diff 确定相关路径
- ✅ 使用 paths_include_glob 进一步限制

**对比**：
```
❌ 错误做法：
serena.find_symbol(
    name_path_pattern="CreateDeployment"
)

✅ 正确做法：
serena.find_symbol(
    name_path_pattern="CreateDeployment",
    relative_path="internal/backend/dpmgr/"
)
```

---

#### 错误 3：忽略 MCP 工具优先策略

**问题**：直接使用传统工具，错过 MCP 工具的优势

**解决**：
- ✅ 记住工具优先级：sequential-thinking > serena > context7 > 传统工具
- ✅ 只在 MCP 工具无法满足时才降级
- ✅ 定期回顾 MCP 工具的使用方式

**对比**：
```
❌ 错误做法：
直接使用 Grep 查找函数

✅ 正确做法：
先尝试 serena.find_symbol
如果失败，再降级到 Grep
```

---

#### 错误 4：不使用 sequential-thinking 规划

**问题**：缺乏系统性，容易遗漏重要检查项

**解决**：
- ✅ 审查开始时必用 sequential-thinking
- ✅ 遇到复杂问题时使用 sequential-thinking
- ✅ 生成报告时使用 sequential-thinking

---

#### 错误 5：不查找相似代码

**问题**：无法判断代码是否符合项目模式

**解决**：
- ✅ 使用 serena.find_symbol 查找相似函数
- ✅ 对比实现方式、错误处理、日志格式
- ✅ 确保新代码与现有代码一致

---

## 工具速查表

| 工具 | 优先级 | 速度 | 用途 | 何时使用 | 替代/降级 |
|------|-------|------|------|---------|-----------|
| sequential-thinking | ⭐⭐⭐ | 快 | 结构化思考 | 规划、分析、报告 | 无 |
| serena.get_symbols_overview | ⭐⭐⭐ | 快 | 结构概览 | 了解文件结构 | Read |
| serena.find_symbol | ⭐⭐⭐ | 中 | 符号查找 | 查找函数/类/方法 | Grep |
| serena.find_referencing_symbols | ⭐⭐⭐ | 中 | 依赖分析 | 理解影响范围 | 无 |
| serena.search_for_pattern | ⭐⭐⭐ | 中 | 模式搜索 | 查找代码模式 | Grep |
| context7 | ⭐⭐ | 中 | 文档查询 | 查阅第三方文档 | Read |
| ReadLints | ⭐ | 快 | Linter 检查 | 审查开始时 | 无 |
| Shell | ⭐ | 快 | 命令执行 | 编译/测试/Git | 无 |
| Read | 降级 | 中 | 读取文件 | 需要完整内容时 | - |
| Grep | 降级 | 中 | 文本搜索 | MCP 工具无法满足时 | - |
| Glob | 补充 | 中 | 文件查找 | 查找文件 | - |

**优先级说明**：
- ⭐⭐⭐：优先使用（MCP 工具）
- ⭐⭐：推荐使用（文档工具）
- ⭐：必要使用（无可替代的传统工具）
- 降级：作为降级方案
- 补充：作为补充工具

---

## 总结

**高效审查的关键原则**：

1. **MCP 工具优先**
   - 优先使用 sequential-thinking、serena、context7
   - 传统工具作为补充和降级方案

2. **语义级理解优先**
   - 使用 serena 工具进行语义级分析
   - 避免过度依赖文本级工具

3. **效率优先**
   - 使用 get_symbols_overview 而非全文 Read
   - 使用 find_symbol 而非 Grep
   - 总是限制搜索范围

4. **渐进式审查**
   - 先快速检查（linter、编译）
   - 再语义分析（serena 工具）
   - 最后深入理解（Read 降级方案）

5. **结构化思考**
   - 审查开始使用 sequential-thinking 规划
   - 复杂问题使用 sequential-thinking 分析
   - 报告生成使用 sequential-thinking 分类

**记住**：
- ✅ 优先使用 MCP 工具
- ✅ 总是限制搜索范围
- ✅ 并行执行独立检查
- ✅ 智能降级，保持灵活性

---

## 相关文档

- [SKILL.md](../SKILL.md) - Code Review Skill 主文档
- [workflow-guide.md](./workflow-guide.md) - 详细的审查工作流程
- [go-standards.md](./go-standards.md) - Go 代码规范速查表
- [checklist.md](./checklist.md) - 详细检查清单
- [file-type-mapping.md](./file-type-mapping.md) - 文件类型识别规则
- [quick-reference.md](./quick-reference.md) - 快速参考手册
- [report-template.md](./report-template.md) - 报告模板和示例
