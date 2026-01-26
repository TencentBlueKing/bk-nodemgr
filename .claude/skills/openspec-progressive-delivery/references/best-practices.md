# OpenSpec 提案编写最佳实践

本文档提炼渐进式提案编写的关键模式和最佳实践。

## 工具使用说明的写法

### 模板结构

```markdown
## 工具使用说明

本提案的设计、实施和验证过程将使用以下 MCP 工具：

### 1. [工具名] - [用途简述]

**核心工具**：
- `tool_function` - 功能描述
  - 参数：param1（说明），param2（说明）
  - 用途：具体用途

**使用场景**：
- 场景 1 的描述
- 场景 2 的描述

**工具组合使用示例**：
```bash
# 例：完整流程的标题
tool1(param1="value")
tool2(param2="value")
```
```

### 关键要素

1. **分层说明**：按工具分组（serena、context7、sequential-thinking）
2. **参数化**：列出关键参数及其说明
3. **场景驱动**：从实际场景解释工具价值
4. **组合示例**：展示多工具协同的完整流程
5. **注释清晰**：每个步骤都有注释说明

---

## 代码审查要求的描述方式

### 模板结构

```markdown
## 代码审查要求

**重要**：每个任务完成前都必须进行代码审查，审查必须在 sub agent 中进行。

### 审查流程

1. **审查时机**：每个任务完成代码修改后、提交前
2. **审查方式**：在 sub agent 中使用 `/code-review` skill
3. **审查内容**：
   - 语法检查（编译、linter 无错误）
   - 规范一致性（错误处理、日志记录、数据转换、注释、命名规范）
   - 逻辑一致性（函数签名、错误消息、日志格式、数据结构一致性）
   - 代码质量（对比已迁移的代码，确保实现模式一致）
4. **审查通过标准**：所有检查项通过后才能提交代码
```

### 关键要素

1. **强制性**：使用"必须"等强制性语言
2. **流程化**：4 个步骤（时机、方式、内容、标准）
3. **工具化**：指定具体的审查工具（`/code-review` skill）
4. **标准化**：明确的通过标准
5. **嵌入式**：在 tasks.md 每个任务中重复审查要求

---

## 验收标准的定义方法

### 多维度验收模板

```markdown
**验收标准**:
- ✅ **功能验收**: [功能是否正确实现]
- ✅ **性能验收**: [性能是否达标 - 具体数字]
- ✅ **兼容性验收**: [新旧功能是否共存]
- ✅ **完整性验收**: [是否有遗漏]
```

### 示例集合

#### 功能验收
```markdown
- ✅ 方法正确查询并返回预期结果
- ✅ 边界情况处理正确
- ✅ 错误处理符合规范
```

#### 性能验收
```markdown
- ✅ 查询性能 < 10ms
- ✅ 工作流总增加 < 100ms
- ✅ 资源使用在合理范围
```

#### 兼容性验收
```markdown
- ✅ 新 API 调用成功
- ✅ 旧功能正常
- ✅ 无冲突或错误
```

#### 完整性验收
```markdown
- ✅ 所有使用方已迁移
- ✅ 所有测试通过
- ✅ 文档已更新
```

### 关键要素

1. **量化原则**：使用具体数字（< 10ms, < 100ms）
2. **可测试性**：每个标准都可以验证
3. **使用 ✅**：标记可以勾选的检查项
4. **多维度**：功能、性能、兼容性、完整性

---

## 风险管理矩阵模板

### 标准格式

```markdown
## 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| [风险名称] | [高/中/低] | [高/中/低] | [具体缓解措施，包含量化指标] |

**风险控制优势**：渐进式迁移将高风险大规模变更转变为低风险的系列小变更。
```

### 示例

```markdown
| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| 性能影响 | 中 | 低 | 使用优化方案，单次操作 < 10ms，总增加 < 100ms |
| 数据迁移失败 | 中 | 低 | 阶段 6 才执行，提供迁移脚本和回滚计划 |
| 代码变更范围大 | 低 | 低 | 渐进式迁移，每个单元独立提交、可独立回滚 |
| 遗漏使用方 | 中 | 低 | 使用 serena 搜索，1-2 周观察期，监控调用 |
| 回滚困难 | 极低 | 低 | 每个单元独立提交，废弃代码不删除，双轨运行 |
```

### 关键要素

1. **影响等级**：高、中、低、极低
2. **概率等级**：高、中、低
3. **量化指标**：具体的数字和时间
4. **缓解具体**：不是泛泛而谈，而是具体措施
5. **对比优势**：突出渐进式迁移的优势

---

## 实时进度追踪模式

### 模板结构

```markdown
## 当前完成情况

### ✅ 已完成（阶段 X）
- [x] 任务 1
- [x] 任务 2

### ⚠️ 进行中（阶段 Y）
- [x] 任务 3（部分完成）
- [ ] 任务 4

**进度**: X/Y 单元已完成迁移（Z%）

### 📝 commit [hash] 的改进（日期）
- ✅ 完成了 ...
- ✅ 优化了 ...

### ❌ 未开始（阶段 Z）
- [ ] 任务 5
- [ ] 任务 6
```

### 关键要素

1. **实时更新**：随实施进度更新文档
2. **可视化**：使用 ✅ ⚠️ ❌ 符号
3. **百分比**：显示进度百分比
4. **commit 追踪**：记录关键 commit 的贡献
5. **活文档**：提案持续更新，不是一次性文档

---

## 任务 MCP 工具使用说明模式

### 模板结构

```markdown
**MCP 工具使用**：
```bash
# 步骤 1: [步骤描述]
tool1(param1="value")

# 步骤 2: [步骤描述]
tool2(param2="value")

# 步骤 3: [步骤描述]
tool3(...)
```
```

### 关键要素

1. **步骤化**：从查看到替换到验证的完整流程
2. **参数具体化**：给出实际的参数值
3. **注释说明**：每个步骤都有注释
4. **可复制性**：执行者可以直接复制使用
5. **完整性**：包含验证步骤

---

## 文档组织最佳实践

### 1. proposal.md 的章节顺序

建议顺序：

1. 变更 ID
2. 工具使用说明 ⭐（降低执行者学习成本）
3. 代码审查要求 ⭐（嵌入质量保证流程）
4. Why（问题陈述）
5. What Changes（渐进式迁移策略）
6. 目标
7. 影响范围
8. 依赖
9. 风险与缓解
10. 实际实现改进 ⭐（记录反馈循环）
11. 当前完成情况 ⭐（实时进度追踪）
12. 参考

### 2. tasks.md 的任务结构

每个任务包含（5 要素）：
1. **任务内容**（子任务列表）
2. **MCP 工具使用说明**（具体步骤）
3. **验收标准**（可量化）
4. **代码审查检查点**（使用 /code-review）
5. **提交信息模板**

### 3. design.md 的设计层次

1. **设计目标**（高层）
2. **设计理念**（模式和原则）
3. **当前架构分析**（问题）
4. **新架构设计**（方案）
5. **核心组件设计**（细节，包含正反示例）
6. **性能/容错/回滚**（工程实践）

### 4. spec.md 的场景驱动

每个需求至少 1 个场景，使用 Given-When-Then：
- **Given**: 前提条件
- **When**: 操作步骤
- **Then**: 期望结果（列表，使用 **强调** 关键要求）

---

## 渐进式迁移的核心组件

### 1. 废弃标记（Deprecation Markers）

**Go 语言标准格式**：
```go
// Deprecated: OldFunction is deprecated, use NewFunction instead.
// This function will be removed in v2.0.0.
func OldFunction() {
    // ... 保留原有实现
}
```

**关键要素**：
- 使用 `// Deprecated:` 标记
- 说明替代方案
- 说明移除版本

### 2. 双轨运行架构

**设计模式**：
```
工作流开始
    │
    ├─ [旧方式: 可用但废弃]
    │   └─ 缓存模式
    │
    └─ [新方式: 推荐]
        └─ 按需查询模式
    │
    ↓
Action 执行（可选择旧或新）
```

**关键要素**：
- 新旧实现同时存在
- 互不干扰
- 可以独立测试

### 3. 容错设计

**多实例顺序重试模式**：
```go
for i, instance := range instances {
    log.Info("trying instance", i)
    err := tryInstance(instance)
    if err == nil {
        log.Info("success with instance", i)
        return nil // 成功即返回
    }
    log.Warn("failed with instance", i, err)
}
return fmt.Errorf("all instances failed")
```

**关键要素**：
- 顺序尝试，不并发
- 成功即返回
- 详细日志记录

---

## 关键成功因素总结

10 个成功因素：

1. ✅ **小步快跑**：每次只改一点，快速验证
2. ✅ **独立提交**：每个变更单元一个 commit
3. ✅ **充分测试**：每个变更都有测试覆盖
4. ✅ **及时回滚**：发现问题立即回滚
5. ✅ **工具化执行**：使用 MCP 工具自动化操作
6. ✅ **质量保证**：内置代码审查流程
7. ✅ **监控观察**：有足够的观察期
8. ✅ **活文档**：文档随实施进度更新
9. ✅ **量化指标**：所有标准都可量化和测试
10. ✅ **追溯性**：每个决策都可追溯到设计文档

---

## 常用代码片段

### 查找所有使用方（serena）
```bash
mcp__serena__find_referencing_symbols(
    name_path="TargetSymbol",
    relative_path="pkg/module/file.go"
)
```

### 搜索代码模式（serena）
```bash
mcp__serena__search_for_pattern(
    pattern="OldAPI|deprecated_function",
    relative_path=".",
    output_mode="files_with_matches"
)
```

### 验证所有使用方已迁移
```bash
# 搜索结果应该 = 0
mcp__serena__find_referencing_symbols(
    name_path="DeprecatedAPI",
    relative_path="."
)
```

### 数据库字段移除（MongoDB）
```javascript
// 移除字段
db.collection.updateMany(
    {},
    { $unset: { "deprecated_field": "" } }
)

// 验证
db.collection.find({ "deprecated_field": { $exists: true } }).count()
// 应该返回 0
```

---

## 提交信息模板

### 阶段 1-2: 新功能
```
feat: add new capability for XXX

- Implement YYY method
- Add ZZZ utility function
- Write unit tests

Related to openspec/changes/<change-id>
```

### 阶段 3: 迁移
```
refactor(module): migrate to new implementation

- Replace old API with new API
- Update tests
- Verify functionality

Progress: X/Y modules migrated (Z%)
Related to openspec/changes/<change-id>
```

### 阶段 4: 废弃
```
deprecate: mark OldAPI as deprecated

- Add Deprecated comment
- Update documentation
- Provide migration guide
- Will be removed in vX.Y.Z

Related to openspec/changes/<change-id>
```

### 阶段 6: 清理
```
cleanup: remove deprecated OldAPI

- Delete implementation
- Remove database field
- Update documentation

Completed openspec/changes/<change-id>
```

---

通过遵循这些最佳实践，你可以编写出高质量、可快速验证、风险可控的渐进式提案。
