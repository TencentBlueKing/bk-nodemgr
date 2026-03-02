## Requirements

### Requirement: Design 审查步骤（阻塞性门禁）

审查流程 SHALL 在"确定范围"之后、"语法检查"之前包含一个独立的 Design 审查步骤，作为阻塞性门禁。该步骤 SHALL 检查代码放置位置（`pkg` vs `internal`）、模块依赖合理性和抽象层级。该步骤 SHALL 对每次审查都执行，不区分变更类型。

#### Scenario: 无设计问题时正常继续

- **WHEN** Design 审查未发现问题
- **THEN** 在 `.review/checklist.md` 中标记步骤完成，继续执行后续步骤

#### Scenario: 发现设计问题时触发门禁

- **WHEN** Design 审查发现设计问题（如 service 逻辑放在了 `pkg`、不合理的跨模块依赖）
- **THEN** 生成早期报告（仅包含设计问题），提示用户"发现设计层面问题，建议先修正后重新提交审查。是否继续后续检查？"

#### Scenario: 用户选择终止审查

- **WHEN** 门禁触发后用户选择终止
- **THEN** 输出早期报告并结束审查。在 `.review/checklist.md` 中记录步骤状态为"⛔ 发现设计问题，用户选择终止"，Issues Found 中记录设计问题

#### Scenario: 用户选择继续审查

- **WHEN** 门禁触发后用户选择继续
- **THEN** 设计问题记录到 Issues Found，照常执行后续步骤 3-7

### Requirement: Complexity / Over-engineering 参考提示

质量检查步骤 SHALL 包含一个复杂度参考提示块，与硬性检查项视觉区分（使用 `> 💭` 格式）。提示内容仅覆盖 linter 无法检测的设计层面复杂度：YAGNI 倾向、过度泛化、冗余抽象（如透传 wrapper）。该提示 SHALL 定位为思考参考，不作为必须报告的检查项。

#### Scenario: 质量检查步骤中展示提示

- **WHEN** 审查执行到质量检查步骤
- **THEN** checklist 中包含 `> 💭 审查时留意` 块，列出复杂度关注模式（YAGNI、过度泛化、冗余抽象），agent 按需判断是否报告

### Requirement: Tests 质量审查（条件触发）

当 PR 包含 `_test.go` 文件时，审查 SHALL 对测试代码执行质量检查。检查项 SHALL 包括：assertion 有效性、假阳性风险、mock 合理性、测试可维护性。审查 SHALL NOT 检查测试完备性（测试完备由 PR 作者负责）。

#### Scenario: PR 包含测试文件

- **WHEN** 变更文件中包含 `_test.go` 文件
- **THEN** checklist 中激活"四、测试质量检查"大节，对测试文件执行 assertion 有效性、假阳性风险、mock 合理性、可维护性检查

#### Scenario: PR 不包含测试文件

- **WHEN** 变更文件中不包含 `_test.go` 文件
- **THEN** 跳过"四、测试质量检查"大节，不在报告中提及

### Requirement: Edge Cases / API consumer 视角

质量检查步骤 SHALL 包含一条边界情况检查项，要求以 API 消费者视角审视代码：nil 输入、空集合、零值、超大输入是否被合理处理。

#### Scenario: 质量检查中审视边界情况

- **WHEN** 审查执行到质量检查步骤
- **THEN** checklist 中包含 `[ ] **边界情况**：以 API 消费者视角审视——nil 输入、空集合、零值、超大输入是否被合理处理` 检查项

### Requirement: 系统健康审查原则

审查原则列表 SHALL 包含第 6 条"系统健康"原则，内容为："评估变更对系统整体的影响，不接受降低代码健康度的变更。小复杂度会累积。"

#### Scenario: 审查原则展示

- **WHEN** 读取 SKILL.md 的审查原则部分
- **THEN** 原则列表包含 6 条，第 6 条为"系统健康"

### Requirement: Documentation 提醒式检查

报告生成步骤 SHALL 检查变更是否涉及需要同步更新的关联文档（`docs/`、`proto/` README、`AGENTS.md`）。若相关文档未随代码变更，SHALL 在报告总结中提醒用户。该提醒 SHALL NOT 作为阻塞性问题。

#### Scenario: 变更涉及文档关联但文档未更新

- **WHEN** 代码变更涉及 proto 定义或 API 行为变化，但 `docs/` 或 proto README 未包含在变更文件中
- **THEN** 报告审查总结中包含 `**文档同步**: 建议更新 X 文档`

#### Scenario: 变更不涉及文档关联

- **WHEN** 代码变更不涉及 API、proto 或公共接口
- **THEN** 报告审查总结中包含 `**文档同步**: 无需更新`

### Requirement: 审查状态持久化（`.review/` 文件夹）

审查 SHALL 在 worktree 下生成 `.review/` 文件夹，包含 `checklist.md`（进度 + 已发现问题）和 `meta.json`（元数据）。`.review/` SHALL 被 `.gitignore` 忽略，不提交到 git。

#### Scenario: 首次审查

- **WHEN** worktree 下不存在 `.review/` 文件夹
- **THEN** 创建 `.review/checklist.md` 和 `.review/meta.json`，从步骤 1 开始

#### Scenario: 续审（branch 和 commit 均匹配）

- **WHEN** `.review/` 存在且 `meta.json` 中的 branch 名和 commit hash 均与当前一致
- **THEN** 提示用户"检测到上次审查进度，是否从步骤 N 继续？"

#### Scenario: 续审（branch 匹配但 commit 变化）

- **WHEN** `.review/` 存在且 branch 名匹配，但 commit hash 不同（如 amend）
- **THEN** 提示用户"检测到同分支的上次审查进度，但 commit 已变更。是否从步骤 N 继续，还是重新开始？"

#### Scenario: 不匹配时重新开始

- **WHEN** `.review/` 存在但 branch 名不匹配
- **THEN** 清空 `.review/` 文件夹，从步骤 1 重新开始

#### Scenario: 实时记录 issues

- **WHEN** 审查过程中任一步骤发现问题
- **THEN** 立即将问题追加到 `.review/checklist.md` 的 Issues Found 部分，更新 `meta.json` 的 currentStep

#### Scenario: 最终报告从 checklist 汇总

- **WHEN** 审查执行到报告生成步骤（步骤 7）
- **THEN** 从 `.review/checklist.md` 的 Issues Found 汇总生成最终报告，而非重新收集

### Requirement: 审查流程扩展为 7 步

审查流程 SHALL 从 6 步扩展为 7 步。新步骤顺序：1. 确定范围 → 2. Design 审查（门禁）→ 3. 语法检查（门禁）→ 4. 规范检查 → 5. 一致性检查 → 6. 质量检查 → 7. 生成报告。

#### Scenario: SKILL.md 流程展示

- **WHEN** 读取 SKILL.md 的审查流程部分
- **THEN** 流程列表包含 7 个步骤，步骤 2 为"🏗️ Design 审查"，标注为阻塞性门禁
