# OpenSpec 工作流指南

本文档说明如何在 BlueKing Node Manager (bk-nodemgr) 项目中使用 OpenSpec 进行变更管理和规范制定。

## 什么是 OpenSpec？

OpenSpec 是一个基于规范的变更管理系统，通过以下方式帮助团队协作：

1. **变更提案 (Change Proposals)**: 在实施前记录和审查变更
2. **规范文档 (Specifications)**: 维护系统能力的正式规范
3. **任务跟踪 (Tasks)**: 将变更分解为可验证的任务
4. **设计文档 (Design Documents)**: 记录架构决策和权衡

## 工作流概览

### 1. 创建变更提案 (`/openspec-proposal`)

当需要引入新功能、重大变更或架构调整时：

1. **审查现有状态**
   - 阅读 `openspec/project.md` 了解项目上下文
   - 运行 `openspec list` 查看现有变更
   - 运行 `openspec list --specs` 查看现有规范
   - 使用 `rg` 或代码搜索了解当前实现

2. **创建提案结构**
   - 选择唯一的动词引导的 `change-id`（如 `add-node-health-monitoring`）
   - 在 `openspec/changes/<id>/` 下创建：
     - `proposal.md` - 变更概述和动机
     - `tasks.md` - 可验证的任务清单
     - `design.md` - 架构设计文档（如需要）
     - `specs/<capability>/spec.md` - 规范变更

3. **编写规范变更**
   - 在 `changes/<id>/specs/<capability>/spec.md` 中使用：
     - `## ADDED Requirements` - 新增需求
     - `## MODIFIED Requirements` - 修改的需求
     - `## REMOVED Requirements` - 移除的需求
   - 每个需求至少包含一个 `#### Scenario:` 场景说明
   - 交叉引用相关能力

4. **验证提案**
   - 运行 `openspec validate <id> --strict`
   - 解决所有验证问题
   - **注意**: 提案阶段不编写代码，只创建设计文档

### 2. 实施变更 (`/openspec-apply`)

提案批准后：

1. **阅读提案文档**
   - 阅读 `changes/<id>/proposal.md`
   - 阅读 `design.md`（如果存在）
   - 阅读 `tasks.md` 确认范围和验收标准

2. **按任务实施**
   - 按顺序完成 `tasks.md` 中的任务
   - 保持更改最小化和聚焦
   - 确保每个任务都完成后再标记

3. **更新任务状态**
   - 完成所有工作后，将 `tasks.md` 中的每个任务标记为 `- [x]`
   - 确保状态反映实际情况

### 3. 归档变更 (`/openspec-archive`)

变更部署后：

1. **确认变更 ID**
   - 使用 `openspec list` 查找变更
   - 确认变更已完成并准备归档

2. **执行归档**
   - 运行 `openspec archive <id> --yes`
   - CLI 会将变更移动到 `changes/archive/` 并更新规范

3. **验证结果**
   - 运行 `openspec validate --strict` 确保规范更新正确
   - 使用 `openspec show <id>` 检查归档结果

## 规范格式

### 需求格式

每个规范文件应遵循以下格式：

```markdown
## ADDED Requirements

### Requirement: <需求名称>

<需求描述>

#### Scenario: <场景名称>
<场景描述>

#### Scenario: <另一个场景>
<场景描述>
```

### 变更类型

- **ADDED**: 新增功能或能力
- **MODIFIED**: 修改现有功能
- **REMOVED**: 移除功能（通常需要迁移计划）

## 最佳实践

### 提案阶段

1. **保持范围聚焦**: 每个提案应该解决一个明确的问题
2. **提供充分上下文**: 解释为什么需要这个变更
3. **识别依赖**: 明确与其他变更或规范的关系
4. **考虑权衡**: 在 `design.md` 中记录架构决策

### 实施阶段

1. **遵循任务顺序**: 按 `tasks.md` 中的顺序实施
2. **最小化更改**: 只实现提案中明确要求的内容
3. **保持测试**: 确保每个任务都有验证步骤
4. **及时更新状态**: 完成任务后立即更新检查清单

### 规范编写

1. **使用场景驱动**: 每个需求至少有一个场景
2. **明确验收标准**: 场景应该可测试和可验证
3. **避免实现细节**: 规范关注"做什么"而非"怎么做"
4. **交叉引用**: 引用相关能力和需求

## 常见命令

```bash
# 列出所有变更
openspec list

# 列出所有规范
openspec list --specs

# 查看变更详情
openspec show <change-id>

# 查看规范详情
openspec show <spec-name> --type spec

# 验证变更
openspec validate <change-id> --strict

# 验证所有内容
openspec validate --strict

# 归档变更
openspec archive <change-id> --yes
```

## 与项目工作流的集成

### PR 标题规范

项目要求 PR 标题遵循格式：
```
<type>: <description> --issue=#<number>
```

OpenSpec 变更提案可以关联到 GitHub Issue，在 PR 中引用变更 ID。

### 代码审查

在代码审查时：
1. 检查实现是否与提案一致
2. 验证所有任务是否完成
3. 确认规范已正确更新

### 文档更新

- 规范变更会自动反映在 `openspec/specs/` 中
- 设计决策记录在 `changes/<id>/design.md`
- 实施细节通过代码注释和 README 补充

## 故障排除

### 验证失败

如果 `openspec validate` 失败：
1. 检查规范格式是否正确
2. 确认所有必需的文件存在
3. 使用 `openspec show <id> --json --deltas-only` 查看详细信息

### 找不到变更

如果找不到变更：
1. 确认变更 ID 拼写正确
2. 检查变更是否已归档（查看 `changes/archive/`）
3. 使用 `openspec list` 查看所有可用变更

### 规范冲突

如果规范之间存在冲突：
1. 检查是否有多个变更修改同一规范
2. 确认变更的顺序和依赖关系
3. 在 `design.md` 中记录解决冲突的决策

## 参考资源

- `openspec/project.md` - 项目上下文和技术栈信息
- `.cursor/commands/openspec-proposal.md` - 提案命令详细说明
- `.cursor/commands/openspec-apply.md` - 实施命令详细说明
- `.cursor/commands/openspec-archive.md` - 归档命令详细说明
