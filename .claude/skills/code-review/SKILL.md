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

### 审查状态持久化（`.review/`）

审查过程中在 worktree 下生成 `.review/` 文件夹，保存审查进度和已发现的问题，支持跨会话续审。

**文件结构：**
```
.review/
├── checklist.md        # 审查进度：各步骤 [x]/[ ] 状态 + 已发现问题摘要
└── meta.json           # 元数据：审查目标（commit/branch）、开始时间、当前步骤
```

- **`checklist.md`**：实时记录审查进度和问题。每个步骤完成后更新，最终报告（步骤 7）从此文件的 Issues Found 汇总生成。门禁终止时，门禁步骤标记 `[x]` 并附注"⛔"，Issues Found 中记录发现的问题。
- **`meta.json`**：记录 target branch、commit hash、startedAt、currentStep、totalSteps。

**续审检测规则：**
- `.review/` 不存在 → 从步骤 1 开始
- branch 名不匹配 → 清空 `.review/` 重新开始
- branch + commit 均匹配 → 提示用户"检测到上次审查进度，是否从步骤 N 继续？"
- branch 匹配但 commit 变化 → 提示用户"commit 已变更，是否继续还是重新开始？"

`.review/` 已加入 `.gitignore`，不会提交到 git。

详细状态管理流程参见：[references/workflow-guide.md](references/workflow-guide.md)

### 审查流程

执行 7 步审查流程：

1. **🔍 确定范围** - 使用 `🧠 sequential-thinking` 分析策略，用 `🔍 serena.get_symbols_overview` 识别文件和模块类型
2. **🏗️ Design 审查 + AGENTS.md 门禁**（阻塞性）- 读取 `AGENTS.md`，用 `🧠 sequential-thinking` 评估设计合理性和 AGENTS.md Anti-patterns 合规性，用 `🔍 serena.get_symbols_overview` 检查代码放置位置和依赖方向。违反 Anti-patterns 或设计问题 → 生成早期报告，提示用户是否终止
3. **⚠️ 语法检查**（阻塞性门禁）- 使用 `make lint`（最高优先级，阻塞性错误）
4. **📋 规范检查 + AGENTS.md Conventions** - 检查 AGENTS.md Conventions 合规性，使用 `🔍 serena.find_symbol` 精确定位，用 `📚 context7` 查询规范（可选）
5. **🔗 一致性检查** - 使用 `🔍 serena.find_symbol` 和 `find_referencing_symbols` 查找并对比相似代码；参考 `cases/` 案例库和 [architecture-principles.md](references/architecture-principles.md)
6. **💡 质量检查** - 使用 `🧠 sequential-thinking` 分析架构原则和质量问题，用 `🔍 serena.search_for_pattern` 查找特定模式
7. **📄 生成报告** - 使用 `🧠 sequential-thinking` 分类问题，按优先级生成结构化报告

**详细流程和工具使用**：参见 [references/workflow-guide.md](references/workflow-guide.md)

### Architecture Judgment Handoff

When a diff affects architecture boundaries, dependency direction, shared contracts, or materially departs from the intended design, load `bk-nodemgr-architecture-judgment` in `post-flight` mode as a focused self-review lens. Map its `Stop` verdict to the existing serious/blocking finding level, `Warn` to an important finding, and `Continue` to no architecture-judgment finding. Do not automatically launch a deepening scan during ordinary reviews.

## 核心检查项

代码审查关注四大类检查项：

### AGENTS.md 合规性（强制）
- Anti-patterns 零容忍：不手编 pb.go、proto 仅作 boundary type、不引入重复 helper、service 逻辑不放 pkg
- Conventions 必须遵循：extending existing code paths、pkg/logger、公共类型英文注释
- Where to look 约定：代码放置位置和依赖方向必须正确

### 语法检查（阻塞性）
- 编译通过、静态分析无错误、Linter 无错误

### 规范一致性
- 错误处理、日志记录、数据转换、注释、命名规范

### 架构与设计原则
- DRY、正交性、单层级函数、显式处理、控制熵
- 详见 [references/architecture-principles.md](references/architecture-principles.md)

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

1. **AGENTS.md 即法律** - AGENTS.md 的 Anti-patterns 和 Conventions 是强制规则，违反即为阻塞性或重要问题，不可降级为建议
2. **效率优先** - 先快速检查（linter、编译），再深入审查
3. **基于规范** - 所有问题都应引用具体规范文档
4. **保持一致** - 对比相似代码，确保实现模式一致
5. **建设性** - 提供具体修复建议，而非仅指出问题
6. **优先级明确** - 区分严重、重要和建议性问题
7. **系统健康** - 评估变更对系统整体的影响，不接受降低代码健康度的变更。小复杂度会累积

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

## Review 案例库

`cases/` 目录逐步积累代码审查中发现的典型案例（好/坏代码对比 + 违反原则标签），供 reviewer 参考。

**审查中的使用方式**：
- 步骤 5（一致性检查）和步骤 6（质量检查）时，若发现的问题与已有案例相似，在报告中引用案例
- 审查结束后，若发现了新的典型问题，可补充为新案例

**添加新案例**：参见 [cases/README.md](cases/README.md) 和 [cases/_TEMPLATE.md](cases/_TEMPLATE.md)

## Common Mistakes

| 错误 | 后果 | 修复 |
|------|------|------|
| 直接使用 Read/Grep 而跳过 serena | Token 浪费 4-5x，结构理解不准确 | 总是先尝试 `serena.get_symbols_overview` 或 `find_symbol` |
| 不限制搜索路径 | 全库扫描，速度慢且结果噪声大 | 所有 serena/Grep 调用必须指定 `relative_path` |
| 跳过 sequential-thinking 直接开始审查 | 遗漏重要检查项，审查不系统 | 审查开始和报告生成时必用 sequential-thinking |
| 不先运行 lint/build 就深入逻辑审查 | 在有语法错误的代码上浪费时间 | 步骤 3（语法检查）必须先于后续所有步骤 |
| 审查时不创建 worktree | 污染当前工作区 | 遵循 `using-git-worktrees` sub-skill |

## 参考索引

| 类别 | 文档 |
|------|------|
| 工作流 | [workflow-guide.md](references/workflow-guide.md)、[quick-reference.md](references/quick-reference.md) |
| 工具 | [tools-usage.md](references/tools-usage.md) |
| 检查 | [checklist.md](references/checklist.md)、[file-type-mapping.md](references/file-type-mapping.md) |
| 架构原则 | [architecture-principles.md](references/architecture-principles.md) |
| 架构判断 | `bk-nodemgr-architecture-judgment` |
| 报告 | [report-template.md](references/report-template.md) |
| Go 规范 | [go-standards.md](references/go-standards.md) |
| 模块 Pattern | [patterns/README.md](patterns/README.md)、[dpmgr-executor.md](patterns/dpmgr-executor.md) |
| Review 案例 | [cases/README.md](cases/README.md) |
| 项目规范 | `AGENTS.md`（**强制**）、`docs/api/API接口开发流程.md`、`.golangci.yml` |
