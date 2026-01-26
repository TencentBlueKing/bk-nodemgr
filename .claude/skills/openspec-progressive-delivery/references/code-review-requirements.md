# 代码审查要求

本文档定义渐进式提案实施过程中的代码审查时机和场景。

**实际审查流程**: 使用 `/code-review` skill，它包含完整的语法检查、规范检查和质量检查能力。

---

## 核心原则

**重要**: 每个任务完成前都**必须**进行代码审查，审查必须在 sub agent 中进行。

代码审查确保:
- ✅ 代码质量一致性
- ✅ 规范遵守完整性
- ✅ 逻辑正确性
- ✅ 无回归风险

---

## 审查时机

### 触发点

每个任务完成代码修改后、提交前。

### 具体场景

| 阶段 | 审查时机 | 重点关注 |
|------|----------|----------|
| 阶段 1 | 新增代码完成后 | 新功能符合规范 |
| 阶段 2 | 双轨逻辑实现后 | 新旧功能互不干扰 |
| 阶段 3 | 每个单元迁移完成后 | 实现模式与已迁移代码一致 |
| 阶段 6 | 清理代码完成后 | 无遗留引用、测试通过 |

**关键**: 不允许跳过审查直接提交代码。

---

## 审查方式

### 在 sub agent 中执行

```markdown
**代码审查** (在 sub agent 中执行):
- [ ] 使用 `/code-review` skill 审查相关文件
- [ ] 审查通过后才能提交
```

**为什么在 sub agent 中**:
- 保持主 agent 的上下文清晰
- 专注于审查任务
- 可以并行审查多个单元

---

## 审查内容

`/code-review` skill 会检查以下 4 个维度:

### 1. 语法检查
- 编译通过 (无错误、无警告)
- Linter 检查通过 (`.golangci.yml` 规则)
- 格式化正确

### 2. 规范一致性
- 错误处理符合规范
- 日志记录使用 `pkg/logger`
- 注释完整清晰

### 3. 逻辑一致性
- 函数签名兼容
- 错误消息一致
- 数据结构兼容

### 4. 代码质量
- 实现模式与已迁移代码一致
- 代码简洁无重复

---

## 渐进式迁移的特殊检查

在阶段 3 (逐个迁移) 中，需要额外确保:

### 模式一致性检查

对比已迁移的代码，确保新迁移的代码使用相同模式:

```go
// 已迁移的单元 A 使用的模式
func (w *WorkflowA) Execute(ctx context.Context) error {
    relays, err := dao.GetRelayListByFilter(ctx, filter)
    if err != nil {
        return fmt.Errorf("failed to get relays: %w", err)
    }
    // ...
}

// 新迁移的单元 B 应该使用相同模式 ✅
func (w *WorkflowB) Execute(ctx context.Context) error {
    relays, err := dao.GetRelayListByFilter(ctx, filter)  // 相同函数
    if err != nil {
        return fmt.Errorf("failed to get relays: %w", err)  // 相同错误处理
    }
    // ...
}
```

### 使用 serena 工具验证

```bash
# 查找已迁移的代码模式
mcp__serena__find_symbol(
    name_path_pattern="WorkflowA/Execute",
    include_body=true
)

# 对比当前迁移的代码
mcp__serena__find_symbol(
    name_path_pattern="WorkflowB/Execute",
    include_body=true
)
```

---

## 审查通过标准

**必须满足所有条件才能通过**:

- ✅ `/code-review` skill 报告无 ❌ 严重问题
- ✅ 无阻塞性的 ⚠️ 重要问题
- ✅ 实现模式与已迁移代码一致 (阶段 3)

---

## 审查失败处理

1. 停止提交
2. 记录问题详情
3. 修复问题
4. 重新运行 `/code-review` skill
5. 通过后提交

---

## 在 tasks.md 中的模板

每个任务都应包含以下审查检查点:

```markdown
### 任务 X.X: [任务名称]

**代码审查** (在 sub agent 中执行):
- [ ] 使用 `/code-review` skill 审查以下文件:
  - `pkg/workflow/node_install.go`
  - `pkg/dao/relay.go`
- [ ] 确保实现模式与已迁移代码一致 (阶段 3)
- [ ] 审查通过后才能提交
```

---

## 参考资源

- **code-review skill**: 完整的审查流程和工具
- **项目规范**: `AGENTS.md`
- **Linter 配置**: `.golangci.yml`
