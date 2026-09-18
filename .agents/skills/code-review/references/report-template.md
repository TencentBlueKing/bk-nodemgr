# 代码审查报告模板

本文档提供标准的代码审查报告格式和示例。

## 标准报告结构

审查报告应按优先级分类问题，每个问题提供具体位置和修复建议。

### 模板

```markdown
# 代码审查报告

## 审查范围

- **文件数**: X 个文件
- **变更行数**: +X -Y
- **涉及模块**: [列出主要模块]

## 严重问题（必须修复）❌

[如果没有严重问题，写"无"]

### 问题 1: [简短描述]

- **位置**: `文件路径:行号`
- **问题**: [详细描述问题]
- **影响**: [说明可能的影响]
- **修复建议**:
  ```go
  [建议的代码]
  ```
- **参考**: [相关规范文档]

## 重要问题（强烈建议修复）⚠️

[如果没有重要问题，写"无"]

### 问题 1: [简短描述]

- **位置**: `文件路径:行号`
- **问题**: [详细描述问题]
- **规范**: [引用具体规范]
- **当前**:
  ```go
  [当前代码]
  ```
- **应改为**:
  ```go
  [建议代码]
  ```

## 建议改进（可选）💡

[如果没有建议，写"无"]

### 建议 1: [简短描述]

- **位置**: `文件路径:行号`
- **建议**: [改进建议]
- **参考**: [相关文档]
- **好处**: [说明改进的好处]

## 审查总结

- **审查结果**: [通过 / 需要修复 / 不建议合并]
- **必须修复的问题数**: X 个严重问题
- **建议修复的问题数**: Y 个重要问题
- **改进建议数**: Z 个建议
- **文档同步**: [无需更新 / 建议更新 X 文档]
- **整体评价**: [1-2句话总结代码质量和主要问题]
```

## 示例：完整审查报告

```markdown
# 代码审查报告

## 审查范围

- **文件数**: 3 个文件
- **变更行数**: +120 -45
- **涉及模块**: Deploy Policy Manager Executor, DAO Layer

## 严重问题（必须修复）❌

### 问题 1: 未检查错误返回值

- **位置**: `internal/backend/dpmgr/executor.go:123`
- **问题**: 调用 `db.Create()` 后未检查错误
- **影响**: 可能导致数据未成功保存但程序继续执行，造成数据不一致
- **修复建议**:
  ```go
  if err := db.Create(&record); err != nil {
      return fmt.Errorf("failed to create record: %w", err)
  }
  ```
- **参考**: [go-standards.md](go-standards.md#错误处理规范)

### 问题 2: 潜在的 goroutine 泄露

- **位置**: `internal/backend/workflow/runner.go:89`
- **问题**: 启动 goroutine 没有提供取消机制
- **影响**: 长时间运行的任务无法被取消，可能导致资源泄露
- **修复建议**:
  ```go
  ctx, cancel := context.WithCancel(ctx)
  defer cancel()
  go func(ctx context.Context) {
      // 在循环中检查 ctx.Done()
  }(ctx)
  ```

## 重要问题（强烈建议修复）⚠️

### 问题 1: 日志格式不符合项目规范

- **位置**: `internal/backend/dpmgr/executor.go:156`
- **问题**: 日志使用单行格式，不符合项目规范
- **规范**: 应使用两行格式（参考 [go-standards.md](go-standards.md#日志记录规范)）
- **当前**:
  ```go
  logger.G.Sys().With("workflow-id", workflowID).Info("workflow completed")
  ```
- **应改为**:
  ```go
  logger.G.Sys().With("workflow-id", workflowID).
      Info("workflow completed")
  ```

### 问题 2: 未使用 conv 包进行数据转换

- **位置**: `pkg/dao/plugin/dao.go:78-85`
- **问题**: 手动循环将 map 的 keys 转换为 slice
- **规范**: 应使用 `conv.MapKeyToSlice()` （参考 [go-standards.md](go-standards.md#数据转换规范)）
- **当前**:
  ```go
  ids := make([]int64, 0, len(idMap))
  for id := range idMap {
      ids = append(ids, id)
  }
  sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
  ```
- **应改为**:
  ```go
  ids := conv.MapKeyToSlice(idMap)
  ```

### 问题 3: 与相似函数不一致

- **位置**: `internal/backend/dpmgr/executor.go:234`
- **问题**: `CreatePluginDeployment` 函数的错误消息格式与同模块其他函数不一致
- **对比**: 同模块的 `CreateAgentDeployment` 使用 "failed to create agent deployment record"
- **当前**: `fmt.Errorf("create plugin deployment failed: %w", err)`
- **应改为**: `fmt.Errorf("failed to create plugin deployment record: %w", err)`
- **参考**: [../patterns/dpmgr-executor.md](../patterns/dpmgr-executor.md)

## 建议改进（可选）💡

### 建议 1: 函数可读性改进

- **位置**: `pkg/dao/plugin/dao.go:120`
- **建议**: 将复杂的查询条件提取为独立函数
- **好处**: 提高代码可读性和可测试性
- **示例**:
  ```go
  func (d *Dao) buildPluginQueryConditions(req *QueryRequest) map[string]interface{} {
      // ...
  }
  ```

### 建议 2: 添加单元测试

- **位置**: `internal/backend/dpmgr/executor.go`
- **建议**: 为新增的 `CreatePluginDeployment` 函数添加单元测试
- **参考**: 同模块的 `executor_test.go`
- **好处**: 确保功能正确性，防止回归

## 审查总结

- **审查结果**: 需要修复
- **必须修复的问题数**: 2 个严重问题
- **建议修复的问题数**: 3 个重要问题
- **改进建议数**: 2 个建议
- **文档同步**: 无需更新
- **整体评价**: 代码功能实现基本正确，但存在错误处理缺失和规范不一致的问题。修复严重问题和重要问题后可以合并。
```

## 问题描述最佳实践

### 严重问题示例

```markdown
❌ 严重：internal/backend/api/handler.go:45
   - 问题：SQL 查询使用字符串拼接，存在 SQL 注入风险
   - 影响：攻击者可能通过构造恶意输入获取或篡改数据
   - 修复：使用参数化查询
     ```go
     db.Where("name = ?", name).Find(&users)
     ```
```

### 重要问题示例

```markdown
⚠️ 重要：pkg/workflow/executor.go:112
   - 问题：错误消息格式与项目规范不一致
   - 规范：应使用 "failed to [action]: %w" 格式
   - 当前：return errors.New("error when executing task")
   - 应改为：return fmt.Errorf("failed to execute task: %w", err)
```

### 建议改进示例

```markdown
💡 建议：internal/backend/service/node.go:89
   - 建议：考虑添加缓存减少数据库查询
   - 参考：类似的 `GetPlugin` 函数已使用缓存
   - 好处：减少数据库压力，提高响应速度
   - 实现：
     ```go
     if cached, ok := cache.Get(key); ok {
         return cached, nil
     }
     ```
```

## 快速检查清单

在生成报告前，使用此清单快速验证：

- [ ] 所有问题都有具体的文件路径和行号
- [ ] 严重问题包含影响说明和修复建议
- [ ] 重要问题引用了具体的规范文档
- [ ] 建议改进说明了好处
- [ ] 问题按优先级正确分类
- [ ] 审查总结包含明确的结论
- [ ] 审查总结包含文档同步字段
- [ ] 使用了正确的图标（❌⚠️💡）
- [ ] 代码示例使用了正确的语法高亮

## 与其他工具集成

### 与 OpenSpec 工作流集成

在审查 OpenSpec 相关变更时，报告应该额外包含：

```markdown
## OpenSpec 合规性

- **相关提案**: `openspec/proposals/XXXX-feature-name.md`
- **实现一致性**: [一致 / 部分一致 / 不一致]
- **未完成的任务**: [列出提案中未完成的任务]
```

### PR 审查特殊要求

PR 审查报告应该在开头添加：

```markdown
## PR 信息

- **PR 标题**: [检查是否符合格式]
- **关联 Issue**: [检查是否正确引用]
- **变更类型**: [feat / fix / refactor / docs / test / chore]
```

## 报告输出格式

默认使用 Markdown 格式，确保：
- 使用正确的标题层级
- 代码块指定语言（```go）
- 使用列表组织多个检查项
- 使用表格展示对比（如果合适）
- 使用引用块强调重要信息

## 注意事项

- **保持简洁**: 每个问题描述控制在 3-5 行
- **提供上下文**: 总是包含文件路径和行号
- **建设性**: 不仅指出问题，还提供解决方案
- **客观**: 基于规范，避免主观判断
- **可操作**: 修复建议应该具体可执行
