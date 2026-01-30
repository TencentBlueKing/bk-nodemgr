---
name: OPSX: Analyze
description: Interactive Q&A to analyze and confirm an OpenSpec proposal before implementation.
category: OpenSpec
tags: [openspec, analyze, interactive]
---
<!-- OPENSPEC:START -->
**Purpose**
通过多轮问答分析 OpenSpec 提案，确保所有需求、任务和设计在实施前都已确认。

**Guardrails**
- 在用户明确确认前，不要进入实施阶段
- 每轮只问 2-3 个聚焦问题，避免信息过载
- 跟踪已确认项和待解决问题

**Analysis Phases**

按顺序完成以下阶段，每阶段确认后再进入下一阶段：

1. **范围确认** - 问题边界、影响范围、成功标准
2. **需求验证** - 场景覆盖、边界情况、向后兼容
3. **设计审查** - 技术方案、风险、权衡取舍
4. **任务检查** - 顺序、依赖、验证标准

**Steps**

1. 读取提案文件 (`proposal.md`, `tasks.md`, `design.md`, specs)
2. 用 3-5 个要点总结提案
3. 逐阶段问答，每轮等待用户响应
4. 根据分析结果更新提案文件
5. 运行 `openspec validate <id> --strict` 验证
6. 用户确认后，提案即可进入 `/openspec-apply`

**Question Types**

- 澄清: "X 具体指什么？"
- 边界: "当 X 为空/无效时如何处理？"
- 集成: "这与现有的 Y 功能如何交互？"
- 验证: "如何确认 X 正常工作？"
<!-- OPENSPEC:END -->
