---
name: code-review
description: 审查代码和 commit，检查语法错误、逻辑一致性和代码质量。当用户说 review、检查错误、审查 commit 时使用。
allowed-tools:
  - sequential-thinking  # MCP 工具优先！
  - mcp__serena__*       # serena 工具族
  - context7             # 文档查询
  - ReadLints            # Linter
  - Shell                # 编译、测试
  - Read                 # 降级方案
  - Grep                 # 降级方案
  - Glob                 # 文件查找
---

# Code Review Skill

高效审查 Go 代码，确保代码质量、规范一致性和功能正确性。

## MCP 工具优先原则

本 Skill 优先使用 MCP 工具进行代码分析和审查：

**核心 MCP 工具**：
- **🧠 sequential-thinking**: 结构化思考和问题分析
- **🔍 serena 工具族**: 精确的符号级代码分析（`get_symbols_overview`、`find_symbol`、`find_referencing_symbols`、`search_for_pattern`）
- **📚 context7**: 文档查询和规范验证

**传统工具作为补充**：Read/Grep/Glob 仅在 MCP 工具不可用或不适用时使用，ReadLints 和 Shell 用于语法检查。

详细工具选择策略参见：[references/tools-usage.md](references/tools-usage.md) 和 [references/quick-reference.md](references/quick-reference.md)

## 快速开始

### PR 审查（推荐）

审查远程 PR 时，使用 [scripts/pr-worktree.sh](scripts/pr-worktree.sh) 创建独立工作区：

```bash
# 根据 PR URL 自动创建 worktree
.claude/skills/code-review/scripts/pr-worktree.sh <PR_URL>

# 支持: GitHub、GitLab、腾讯工蜂、Gitee
```

脚本自动解析 PR URL、fetch 分支、创建 worktree。创建后进入 worktree 目录进行审查。

### 代码审查流程

执行 6 步审查流程：

1. **🔍 确定范围** - 使用 `🧠 sequential-thinking` 分析策略，用 `🔍 serena.get_symbols_overview` 识别文件和模块类型
2. **⚠️ 语法检查** - 使用 [scripts/go-lint.sh](scripts/go-lint.sh)（最高优先级，阻塞性错误）
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

## 工具使用策略

优先使用 MCP 工具（sequential-thinking、serena 工具族、context7）进行代码分析，它们提供精确的符号级分析、语义级理解和结构化思考能力。传统工具（Read/Grep/Glob）仅在 MCP 工具不可用或不适用时使用。

**详细工具选择指南**：参见 [references/tools-usage.md](references/tools-usage.md) 和 [references/quick-reference.md](references/quick-reference.md)

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

## 相关文档

**Skill 核心文档**：
- [references/workflow-guide.md](references/workflow-guide.md) - 详细工作流指南
- [references/tools-usage.md](references/tools-usage.md) - 工具使用策略
- [references/quick-reference.md](references/quick-reference.md) - 快速参考
- [references/checklist.md](references/checklist.md) - 完整检查清单
- [references/file-type-mapping.md](references/file-type-mapping.md) - 文件类型映射表
- [references/report-template.md](references/report-template.md) - 报告模板
- [references/go-standards.md](references/go-standards.md) - Go 规范速查

**模块规范**：
- [patterns/README.md](patterns/README.md) - Pattern 系统说明
- [patterns/dpmgr-executor.md](patterns/dpmgr-executor.md) - Deploy Policy Manager Executor 规范

**项目规范**：
- `AGENTS.md` - AI 协作规范
- `CLAUDE.md` - 项目指令
- `docs/api/API接口开发流程.md` - API 开发规范
- `docs/developer/README.md` - 开发者指南
- `.golangci.yml` - Linter 配置
