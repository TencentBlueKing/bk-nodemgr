---
name: openspec-progressive-delivery
description: "Progressive delivery strategy for large-scale OpenSpec changes. Use as a reference when /openspec:proposal, /openspec:apply, or /openspec:archive commands handle changes affecting >5 files, breaking APIs, architecture refactoring, or data migrations. Provides 6-phase migration patterns, verification strategies, and rollback procedures."
---

# OpenSpec 渐进式交付策略

本 skill 为大规模 OpenSpec 变更提供渐进式交付的参考策略。

**与现有命令的配合**：
- `/openspec:proposal` - 创建提案时，参考本 skill 设计 6 阶段迁移路径
- `/openspec:apply` - 实施变更时，参考本 skill 的独立提交和验证策略
- `/openspec:archive` - 归档时，确认所有阶段完成

## 快速导航

- 📖 **[工具使用指南](references/tool-usage.md)** - serena 工具的详细用法
- ✅ **[代码审查要求](references/code-review-requirements.md)** - 审查时机和场景
- 📋 **[6阶段迁移模板](references/progressive-migration-template.md)** - 详细的阶段指导
- 💡 **[最佳实践](references/best-practices.md)** - 从实际案例提炼的经验
- 📚 **[实际示例](references/examples.md)** - 5个完整示例
- 🤖 **[Agent 协作指南](references/agent-collaboration.md)** - 多 Agent 协作模式

---

## 适用场景

### 必须使用渐进式策略

- 影响 > 5 个文件或模块
- 破坏性 API 变更
- 数据结构变更
- 架构重构

### 可选使用

- 影响 2-5 个文件
- 向后兼容的 API 添加
- 功能增强

### 不需要

- 单文件修改
- 配置参数添加
- 文档更新

---

## 核心原则

### 1. 小步快跑
```
大规模变更 = 系列小变更 × 独立验证 × 可回滚
```

### 2. 快速验证
每个阶段完成后立即使用 `/code-review` skill 验证。

### 3. 独立提交
每个功能单元一个 commit，每个模块迁移一个 commit。

---

## 6 阶段迁移模式

```
阶段 1: 准备新能力 (无破坏性)
  → 新增代码，不影响现有功能

阶段 2: 双轨运行 (新旧共存)
  → 新旧实现共存，互不干扰

阶段 3: 逐个迁移 (独立切换)
  → 每个使用方独立迁移
  → 每个单元一个 commit

阶段 4: 标记废弃 (保留代码)
  → 添加 Deprecated 注释

阶段 5: 观察期 (监控稳定)
  → 生产环境运行 1-2 周

阶段 6: 清理删除 (安全移除)
  → 确认无使用方后删除
```

详见 [progressive-migration-template.md](references/progressive-migration-template.md)。

---

## 与 /openspec:proposal 的配合

创建提案时，在 `proposal.md` 中：

1. **What Changes 章节** - 描述 6 阶段迁移策略
2. **风险与缓解** - 使用风险管理矩阵
3. **当前完成情况** - 实时更新进度

在 `tasks.md` 中：

- 按阶段组织任务
- 每个任务包含验收标准
- 标记代码审查检查点

---

## 与 /openspec:apply 的配合

实施变更时：

### 执行流程

```bash
# 1. 分析代码 (使用 serena)
mcp__serena__find_symbol(...)
mcp__serena__find_referencing_symbols(...)

# 2. 修改代码
mcp__serena__replace_symbol_body(...)

# 3. 验证 (使用 /code-review)
# 在 sub agent 中执行 /code-review

# 4. 提交
git commit -m "refactor: migrate X to Y"

# 5. 更新 tasks.md
# 标记 - [x] 完成
```

### 验证清单

每个任务完成后，使用 `/code-review` skill 验证：

- [ ] 语法检查通过
- [ ] 规范一致性
- [ ] 逻辑一致性
- [ ] 实现模式与已迁移代码一致

---

## 与 /openspec:archive 的配合

归档前确认：

- [ ] 所有阶段完成 (阶段 6 清理完成)
- [ ] 旧代码已删除
- [ ] 所有测试通过
- [ ] tasks.md 全部标记 `[x]`

---

## 提交信息格式

```
<type>: <description>

类型:
- feat: 新功能 (阶段 1-2)
- refactor: 重构 (阶段 3)
- deprecate: 废弃标记 (阶段 4)
- cleanup: 清理删除 (阶段 6)
```

---

## 参考资源

### 详细指南

- [progressive-migration-template.md](references/progressive-migration-template.md) - 6 阶段详细模板
- [tool-usage.md](references/tool-usage.md) - serena 工具使用
- [code-review-requirements.md](references/code-review-requirements.md) - 审查时机
- [best-practices.md](references/best-practices.md) - 最佳实践
- [examples.md](references/examples.md) - 完整示例
- [agent-collaboration.md](references/agent-collaboration.md) - Agent 协作

