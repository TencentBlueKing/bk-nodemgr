# 代码审查快速参考手册

本文档提供代码审查过程中常用的工具、命令和模式的快速查找。

## MCP 工具快速参考

### sequential-thinking

**用途**：复杂决策、多步骤分析、问题分类

**使用场景**：
- 审查开始时规划策略
- 遇到复杂问题需要结构化分析
- 生成报告时进行问题分类和优先级判断

**优势**：结构化思考、可追溯的推理过程

### serena 工具命令速查

#### serena.get_symbols_overview
```
用途：快速了解文件结构，避免全文读取
时机：审查开始时，需要了解文件整体结构
替代：Read（在已知需要完整内容时才用）
```

#### serena.find_symbol
```
用途：精确查找函数/类/方法定义
时机：需要查找特定符号的实现
替代：Grep（语义级搜索更准确）
参数：使用 depth 参数获取子符号（如类的方法）
```

#### serena.find_referencing_symbols
```
用途：理解代码依赖关系和影响范围
时机：需要分析代码修改的影响范围
替代：无（传统工具无法提供此能力）
```

#### serena.search_for_pattern
```
用途：语义级模式搜索
时机：需要查找特定的代码模式
替代：Grep（在 serena 无法满足时降级使用）
```

### context7

**用途**：获取最新的第三方文档和最佳实践

**使用场景**：
- 需要验证 API 使用正确性
- 查阅框架和标准库文档
- 确保规范准确性，减少过时信息

## 常用命令

### 编译和静态检查
```bash
# 编译检查
go build ./path/to/package/...

# 静态分析
go vet ./path/to/package/...
```

### Git 操作
```bash
# 查看暂存的变更
git diff --staged

# 查看变更的文件
git status -uno

# 查看特定 commit
git show <commit-hash>
```

## 常用 Grep 模式

### 查找函数
```bash
# 查找相似函数
Grep pattern="func Create.*Deployment"
```

### 查找错误处理
```bash
# 查找错误包装
Grep pattern="fmt\\.Errorf.*%w"
```

### 查找日志使用
```bash
# 查找结构化日志
Grep pattern="logger\\.G\\.Sys\\(\\)\\.With"
```

## 工具选择速查

### 语法检查
- **ReadLints** - 检查 linter 错误（最高优先级）
- **Shell: go build** - 编译检查
- **Shell: go vet** - 静态分析

### 代码理解
- **serena.get_symbols_overview** - 快速了解文件结构（优先）
- **serena.find_symbol** - 精确查找符号定义（优先）
- **Read** - 读取完整文件内容（降级方案）

### 代码搜索
- **serena.find_symbol** - 语义级符号搜索（优先）
- **serena.search_for_pattern** - 语义级模式搜索（优先）
- **Grep** - 文本级模式搜索（降级方案）

### 依赖分析
- **serena.find_referencing_symbols** - 查找符号引用和依赖（无可替代）

### 文档查询
- **context7** - 第三方文档和最佳实践（优先）
- **Read** - 项目内规范文档（补充）

### 文件查找
- **Glob** - 按文件名模式查找

## 工具组合示例

### 基础审查（单文件小修改）
```
1. sequential-thinking - 规划审查策略
2. serena.get_symbols_overview - 了解文件结构
3. ReadLints + Shell - 语法检查
4. serena.find_symbol - 查找相似代码，检查一致性
5. sequential-thinking - 生成报告
```

### 完整审查（新功能或复杂变更）
```
1. sequential-thinking - 分析变更范围和策略
2. Shell: git diff - 了解变更文件
3. serena.get_symbols_overview - 批量了解文件结构
4. ReadLints + Shell - 语法检查
5. serena.find_symbol - 查找需要审查的符号
6. serena.find_referencing_symbols - 理解影响范围
7. context7（可选）- 查阅相关规范
8. sequential-thinking - 生成详细报告
```

### Commit/PR 审查
```
1. sequential-thinking - 规划审查策略
2. Shell: git show - 查看变更
3. 执行完整审查流程
4. OpenSpec 提案一致性检查
5. sequential-thinking - 生成 PR 审查报告
```

## 快速决策树

```
审查任务
    │
    ├─ 需要语法检查？
    │   └─ 是 → ReadLints + Shell (go build/vet)
    │
    ├─ 需要理解代码结构？
    │   └─ 是 → sequential-thinking（规划）
    │           → serena.get_symbols_overview（结构）
    │
    ├─ 需要查找特定符号？
    │   └─ 是 → serena.find_symbol
    │           （如果失败 → 降级到 Grep）
    │
    ├─ 需要理解依赖关系？
    │   └─ 是 → serena.find_referencing_symbols
    │
    ├─ 需要查阅规范文档？
    │   └─ 是 → context7（第三方文档）
    │           → Read（项目内规范）
    │
    ├─ 需要模式搜索？
    │   └─ 是 → serena.search_for_pattern
    │           （如果失败 → 降级到 Grep）
    │
    └─ 需要完整文件内容？
        └─ 是 → Read
```

## 优先级原则

1. **MCP 工具优先**：优先使用 sequential-thinking 和 serena 工具族
2. **语义理解优先**：使用语义级工具（serena）而非文本级工具（Grep）
3. **效率优先**：使用 get_symbols_overview 而非全文 Read
4. **降级策略**：当 MCP 工具无法满足时，降级到传统工具
5. **实用主义**：保持工具选择的灵活性，根据实际需求调整

## 常见问题

### Q: 什么时候使用 Read 而不是 serena.get_symbols_overview？
A: 当你需要查看完整的文件内容，包括注释、导入语句、完整的代码实现时使用 Read。get_symbols_overview 只提供结构概览。

### Q: 什么时候使用 Grep 而不是 serena.search_for_pattern？
A: 当你需要进行简单的文本模式匹配，或者 serena 无法理解你的查询意图时，降级使用 Grep。

### Q: 如何决定是否使用 context7？
A: 当你需要验证第三方库的 API 使用、查阅最新的框架文档、或者需要确认某个实践是否符合最新标准时，使用 context7。

### Q: sequential-thinking 在什么时候必须使用？
A: 在审查开始时（规划策略）、遇到复杂问题需要分析时、以及生成最终报告时（问题分类）都应该使用。
