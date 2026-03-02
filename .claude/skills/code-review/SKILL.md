---
name: code-review
description: Use when user asks to review code, check for errors, audit commits, or review a PR. Also use when reviewing Go changes for lint failures, inconsistent patterns, or logic bugs.
---

# Code Review Skill

高效审查 Go 代码，确保代码质量、规范一致性和功能正确性。

## When to Use

**使用场景：**
- 用户要求审查代码、检查错误、审计提交或审查 PR
- 审查 Go 代码中的 lint 失败、不一致模式或逻辑错误
- 合并前的代码质量验证

**不适用场景：**
- 仅需理解代码含义（使用普通代码阅读，不启动完整审查流程）
- 快速单行修复确认（无需创建 worktree 和生成报告）

## 工具优先级

优先使用 MCP 工具进行代码分析：

| 优先级 | 工具 | 用途 |
|--------|------|------|
| 1 | sequential-thinking | 结构化思考、问题分析 |
| 2 | serena 工具族 | 符号级代码分析（`get_symbols_overview`、`find_symbol`、`find_referencing_symbols`、`search_for_pattern`） |
| 3 | context7 | 文档查询、规范验证 |
| 降级 | Read/Grep/Glob | 仅在 MCP 工具不可用时使用 |

详细工具选择策略参见：[references/tools-usage.md](references/tools-usage.md)

## 快速开始

**隔离工作区**：审查代码前应创建独立 worktree，避免污染当前工作区。

**REQUIRED SUB-SKILL:** Use `using-git-worktrees` for worktree lifecycle (directory selection, gitignore verification, baseline tests).

**PR 审查额外步骤**：先用 [scripts/pr-fetch.sh](scripts/pr-fetch.sh) 将 PR 分支 fetch 到本地（不切换当前分支），再按 `using-git-worktrees` 创建 worktree：

```bash
BRANCH=$(.claude/skills/code-review/scripts/pr-fetch.sh <PR_URL>)
# 然后按 using-git-worktrees 流程创建 worktree
git worktree add .worktrees/"$BRANCH" "$BRANCH"
```

### 审查流程

执行 6 步审查流程：

1. **🔍 确定范围** - 使用 `🧠 sequential-thinking` 分析策略，用 `🔍 serena.get_symbols_overview` 识别文件和模块类型
2. **⚠️ 语法检查** - 使用 `make lint`（最高优先级，阻塞性错误）
3. **📋 规范检查** - 使用 `🔍 serena.find_symbol` 精确定位，用 `📚 context7` 查询规范（可选）
4. **🔗 一致性检查** - 使用 `🔍 serena.find_symbol` 和 `find_referencing_symbols` 查找并对比相似代码
5. **💡 质量检查** - 使用 `🧠 sequential-thinking` 分析质量问题，用 `🔍 serena.search_for_pattern` 查找特定模式
6. **📄 生成报告** - 使用 `🧠 sequential-thinking` 分类问题，按优先级生成结构化报告

**详细流程和工具使用**：参见 [references/workflow-guide.md](references/workflow-guide.md)

## 核心检查项

代码审查关注三大类检查项：

### 语法检查（阻塞性）
- 编译通过、静态分析无错误、Linter 无错误

### 规范一致性
- 错误处理、日志记录、数据转换、注释、命名规范

### 逻辑一致性
- 函数签名、错误消息、日志格式、数据结构一致性

**详细检查清单**：参见 [references/checklist.md](references/checklist.md)

## 文件类型识别

根据文件路径特征自动识别模块类型，加载相应的规范文档：

- **API 接口** (`proto/`, API 相关) → API 接口开发流程
- **Proto 文件** (`*.proto`) → Proto 规范
- **Executor 模式** (`internal/backend/dpmgr/executor.go`) → Executor 规范
- **DAO 层** (`pkg/dao/`) → 数据访问规范
- **REST 框架** (`pkg/rest/`) → REST 错误处理规范

**完整文件类型映射表**：参见 [references/file-type-mapping.md](references/file-type-mapping.md)

## 报告格式

生成结构化的审查报告，按三级优先级分类问题：❌ 严重问题（必须修复）、⚠️ 重要问题（强烈建议修复）、💡 建议改进（可选）。

**完整报告模板和示例**：参见 [references/report-template.md](references/report-template.md)

## 审查原则

1. **效率优先** - 先快速检查（linter、编译），再深入审查
2. **基于规范** - 所有问题都应引用具体规范文档
3. **保持一致** - 对比相似代码，确保实现模式一致
4. **建设性** - 提供具体修复建议，而非仅指出问题
5. **优先级明确** - 区分严重、重要和建议性问题

## 模块规范（Patterns）

`patterns/` 目录存储模块特定的审查规范。当审查特定模块的代码时，读取相应的 pattern 文档。

**如何使用 Patterns**：
1. 识别被修改的文件属于哪个模块
2. 读取对应的 pattern 文档
3. 使用 `serena.find_symbol` 查找同系列的其他函数
4. 检查一致性：函数签名、错误处理、日志格式、数据结构

**当前支持的模块**：
- **Deploy Policy Manager Executor**: [patterns/dpmgr-executor.md](patterns/dpmgr-executor.md)

**添加新 Pattern**：参见 [patterns/README.md](patterns/README.md) 和 [patterns/_TEMPLATE.md](patterns/_TEMPLATE.md)

## Common Mistakes

| 错误 | 后果 | 修复 |
|------|------|------|
| 直接使用 Read/Grep 而跳过 serena | Token 浪费 4-5x，结构理解不准确 | 总是先尝试 `serena.get_symbols_overview` 或 `find_symbol` |
| 不限制搜索路径 | 全库扫描，速度慢且结果噪声大 | 所有 serena/Grep 调用必须指定 `relative_path` |
| 跳过 sequential-thinking 直接开始审查 | 遗漏重要检查项，审查不系统 | 审查开始和报告生成时必用 sequential-thinking |
| 不先运行 lint/build 就深入逻辑审查 | 在有语法错误的代码上浪费时间 | 步骤 2（语法检查）必须先于其他所有步骤 |
| 审查时不创建 worktree | 污染当前工作区 | 遵循 `using-git-worktrees` sub-skill |

## 参考索引

| 类别 | 文档 |
|------|------|
| 工作流 | [workflow-guide.md](references/workflow-guide.md)、[quick-reference.md](references/quick-reference.md) |
| 工具 | [tools-usage.md](references/tools-usage.md) |
| 检查 | [checklist.md](references/checklist.md)、[file-type-mapping.md](references/file-type-mapping.md) |
| 报告 | [report-template.md](references/report-template.md) |
| Go 规范 | [go-standards.md](references/go-standards.md) |
| 模块 Pattern | [patterns/README.md](patterns/README.md)、[dpmgr-executor.md](patterns/dpmgr-executor.md) |
| 项目规范 | `AGENTS.md`、`docs/api/API接口开发流程.md`、`.golangci.yml` |
