# OpenSpec 变更管理系统

OpenSpec 是用于管理项目变更和规范的框架。本文档提供快速入门指南。

## 快速开始

### 1. 了解项目上下文

阅读 `project.md` 了解项目的技术栈、结构和规范。

### 2. 查看现有变更和规范

```bash
# 列出所有变更
openspec list

# 列出所有规范
openspec list --specs
```

### 3. 创建变更提案

使用 `/openspec-proposal` 命令创建新的变更提案。提案阶段只创建设计文档，不编写代码。

### 4. 实施变更

提案批准后，使用 `/openspec-apply` 命令实施变更。

### 5. 归档变更

变更部署后，使用 `/openspec-archive` 命令归档变更并更新规范。

## 文档结构

```
openspec/
├── AGENTS.md          # 详细工作流指南
├── project.md         # 项目上下文和技术栈信息
├── changes/           # 变更提案目录
│   ├── <change-id>/   # 每个变更的目录
│   │   ├── proposal.md
│   │   ├── tasks.md
│   │   ├── design.md  # 可选
│   │   └── specs/     # 规范变更
│   └── archive/       # 已归档的变更
└── specs/             # 当前规范文档
```

## 更多信息

- 详细工作流：查看 `AGENTS.md`
- 项目信息：查看 `project.md`
- 命令帮助：查看 `.cursor/commands/` 目录下的命令文档
