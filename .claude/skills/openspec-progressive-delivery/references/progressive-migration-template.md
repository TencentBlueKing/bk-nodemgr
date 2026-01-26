# 渐进式迁移模板

本文档提供标准 6 阶段渐进式迁移的详细模板和指导。

## 标准 6 阶段模式

### 阶段 1: 准备新能力（无破坏性）

**目标**：添加新的实现，不影响现有代码

**典型工作**：
- 实现新的 API/函数/模块
- 添加新的数据查询方法
- 准备新的工具函数
- 编写单元测试

**验证清单**：
- [ ] 编译通过（< 30s）
- [ ] 新增代码的单元测试通过（100%）
- [ ] 现有测试全部通过（无回归）
- [ ] 代码审查通过（使用 `/code-review` skill）
- [ ] 性能基准测试（如有需要）

**提交格式**：
```
feat: add new capability for XXX

- Implement YYY method
- Add ZZZ utility function
- Write unit tests for new code

Related to openspec/changes/<change-id>
```

**回滚策略**：
- 直接回滚 commit：`git revert <commit-hash>`
- 影响范围：0（无影响）

---

### 阶段 2: 双轨运行（新旧共存）

**目标**：确保新旧实现可以同时存在，互不干扰

**典型工作**：
- 验证新旧 API 可以共存
- 添加特性开关（如果需要）
- 更新文档说明新旧两种方式

**验证清单**：
- [ ] 旧功能正常运行
- [ ] 新功能可以独立调用
- [ ] 两者互不影响
- [ ] 集成测试通过
- [ ] 代码审查通过

**提交格式**：
```
feat: enable dual-track mode for XXX

- New implementation coexists with old one
- Both paths are tested and working
- Add feature flag (if needed)

Related to openspec/changes/<change-id>
```

**回滚策略**：
- 回滚 commit
- 影响范围：低（新功能未启用）

---

### 阶段 3: 逐个迁移（独立切换）

**目标**：将使用方逐个从旧实现迁移到新实现

**关键原则**：
- **每次只迁移一个单元**（1 个文件或 1 个模块）
- **每个单元一个独立 commit**
- **立即验证，失败立即回滚**

**典型工作**：
- 更新单个使用方的代码
- 运行该使用方的测试
- 提交该使用方的变更

**迁移单元示例**：
```markdown
## 迁移清单

| 单元 | 文件 | 状态 | Commit | 备注 |
|------|------|------|--------|------|
| 模块 A | module_a.go | ✅ | abc123 | 已完成 |
| 模块 B | module_b.go | ⚠️ | - | 进行中 |
| 模块 C | module_c.go | ❌ | - | 待开始 |
| 模块 D | module_d.go | ❌ | - | 待开始 |

**进度**: 1/4 (25%)
```

**每个单元的验证清单**：
- [ ] 代码审查通过（在 sub agent 中）
- [ ] 编译通过
- [ ] 该单元的单元测试通过
- [ ] 该单元的集成测试通过
- [ ] 功能手动验证通过
- [ ] 性能符合要求

**提交格式**：
```
refactor(module-a): migrate to new implementation

- Replace old API with new API
- Update unit tests
- Verify functionality

Progress: 1/4 modules migrated (25%)
Related to openspec/changes/<change-id>
```

**回滚策略**：
- 回滚单个 commit：`git revert <commit-hash>`
- 影响范围：单个模块（1-5% 代码库）
- 回滚时间：< 5 分钟

---

### 阶段 4: 标记废弃（保留代码）

**目标**：明确告知开发者旧 API 已废弃，但不删除代码

**典型工作**：
- 添加 `// Deprecated:` 注释
- 更新文档说明废弃状态
- 添加迁移指南

**Go 语言废弃标记标准**：
```go
// Deprecated: OldFunction is deprecated, use NewFunction instead.
// This function will be removed in v2.0.0.
func OldFunction() {
    // ... 保留原有实现
}
```

**验证清单**：
- [ ] 废弃注释格式正确
- [ ] 废弃注释包含替代方案
- [ ] 废弃注释包含移除版本
- [ ] 文档已更新
- [ ] 代码仍可编译运行

**提交格式**：
```
deprecate: mark OldAPI as deprecated

- Add Deprecated comment to OldFunction
- Update documentation
- Provide migration guide to NewFunction
- Will be removed in v2.0.0

Related to openspec/changes/<change-id>
```

**回滚策略**：
- 移除废弃标记
- 影响范围：0（代码未删除）

---

### 阶段 5: 观察期（监控稳定）

**目标**：在生产环境运行 1-2 周，确保新实现稳定

**典型工作**：
- 监控性能指标
- 监控错误日志
- 收集用户反馈
- 确认无遗漏的使用方

**监控清单**：
- [ ] 性能指标正常（< 目标值）
- [ ] 错误率正常（无回归）
- [ ] 无废弃 API 的调用（或已知的计划迁移）
- [ ] 用户反馈正面
- [ ] 无未发现的边界情况

**监控工具示例**：
```bash
# 监控废弃 API 调用
grep -r "OldFunction" logs/ | wc -l

# 监控性能
# (使用项目的性能监控工具)

# 搜索代码中是否还有使用
mcp__serena__find_referencing_symbols(
    name_path="OldFunction",
    relative_path="pkg/",
)
```

**观察期结束标准**：
- 运行时间 >= 1-2 周
- 所有监控指标正常
- 确认无遗漏的使用方
- 获得团队/用户批准

---

### 阶段 6: 清理删除（安全移除）

**目标**：移除废弃的代码和数据

**前置条件**：
- ✅ 观察期结束
- ✅ 所有监控指标正常
- ✅ 确认无使用方
- ✅ 获得批准

**典型工作**：
- 删除废弃的代码
- 更新文档
- 更新 CHANGELOG


**验证清单**：
- [ ] 废弃代码已删除
- [ ] 数据库迁移成功
- [ ] 数据库迁移已验证
- [ ] 所有测试通过
- [ ] 文档已更新
- [ ] CHANGELOG 已更新
- [ ] 代码审查通过

**提交格式**：
```
cleanup: remove deprecated OldAPI

- Delete OldFunction implementation
- Remove database field old_field
- Update documentation
- Update CHANGELOG

Completed openspec/changes/<change-id>
```

**回滚策略**：
- 恢复废弃代码：`git revert <commit-hash>`
- 回滚数据库迁移（使用备份）
- 影响范围：中（需要数据修复）
- 回滚时间：< 30 分钟（代码）+ 数据恢复时间

---

## 快速验证方法模板

### 编译验证（< 30s）
```bash
make build
# 或
go build ./...
# 或
npm run build
```

**成功标准**：
- 无编译错误
- 无编译警告
- 构建产物生成成功

---

### 单元测试验证（< 2min）
```bash
make test
# 或
go test ./...
# 或
npm test
```

**成功标准**：
- 100% 测试通过
- 无跳过的测试
- 覆盖率 >= 目标值（如 80%）

---

### 集成测试验证（< 5min）
```bash
make integration-test
# 或
go test -tags=integration ./...
```

**成功标准**：
- 关键路径测试通过
- 端到端场景测试通过

---

### 功能验证（< 10min）
手动或自动化测试关键功能：

**验证步骤示例**：
1. 启动应用：`make run`
2. 执行操作 A
3. 验证结果 A
4. 执行操作 B
5. 验证结果 B

**成功标准**：
- 所有关键功能正常
- 用户体验无变化
- 无错误或异常

---

### 性能验证（< 5min）
```bash
# 运行性能测试
go test -bench=. ./...

# 或使用性能分析工具
# ab, wrk, k6, etc.
```

**成功标准**：
- 响应时间 < 目标值（如 < 100ms）
- 吞吐量 >= 目标值
- 资源使用正常

---

### 代码审查（< 10min）
在 sub agent 中执行：
```bash
# 使用 /code-review skill
/code-review <file-path>
```

**成功标准**：
- 语法检查通过
- 规范一致性通过
- 逻辑一致性通过
- 代码质量通过

---

## 回滚策略模板

### 回滚决策树

```
验证失败？
├─ 是 → 立即回滚
│   ├─ 阶段 1-2: git revert (影响范围: 0)
│   ├─ 阶段 3: git revert 单个 commit (影响范围: 1 模块)
│   ├─ 阶段 4: 移除废弃标记 (影响范围: 0)
│   └─ 阶段 6: 恢复代码 + 数据 (影响范围: 中)
└─ 否 → 继续下一步
```

### 回滚步骤模板

#### 代码回滚
```bash
# 1. 查看最近的 commit
git log --oneline -n 5

# 2. 回滚特定 commit
git revert <commit-hash>

# 3. 解决冲突（如有）
git mergetool

# 4. 验证回滚成功
make build && make test

# 5. 提交回滚
git commit -m "revert: rollback XXX due to YYY"
```

#### 数据库回滚
```bash
# 1. 使用备份恢复
mongorestore --db mydb --drop backup/

# 2. 或执行反向迁移脚本
mongo mydb < rollback_migration.js

# 3. 验证数据完整性
# (运行验证脚本)
```

---

## 常见问题和解决方案

### Q1: 阶段 3 的迁移单元如何划分？

**A**: 单元划分原则：
- 功能独立性：每个单元可以独立运行和测试
- 影响范围：单个单元影响 < 5% 代码库
- 时间可控：单个单元迁移 < 2 小时

**划分示例**：
- 按文件划分：每个文件一个单元
- 按模块划分：每个模块一个单元
- 按功能划分：每个功能点一个单元

### Q2: 如果一个单元的迁移失败了怎么办？

**A**: 立即回滚流程：
1. 停止继续迁移其他单元
2. 分析失败原因
3. 回滚该单元的 commit
4. 修复问题
5. 重新测试
6. 再次提交

### Q3: 观察期发现问题怎么办？

**A**: 根据问题严重程度决定：
- **严重问题**（功能故障、性能严重下降）：立即回滚到阶段 2（双轨运行）
- **中等问题**（边界情况错误）：修复问题，延长观察期
- **轻微问题**（日志错误、文档问题）：修复问题，继续观察

### Q4: 如何确认所有使用方都已迁移？

**A**: 使用 MCP 工具搜索：
```bash
# 搜索废弃 API 的引用
mcp__serena__find_referencing_symbols(
    name_path="DeprecatedFunction",
    relative_path=".",
)

# 搜索代码中的模式
mcp__serena__search_for_pattern(
    pattern="DeprecatedFunction|old_api",
    relative_path=".",
    output_mode="files_with_matches",
)
```

如果搜索结果 = 0，则确认已全部迁移。

### Q5: 渐进式迁移是否适用于所有变更？

**A**: 不适用，参考判断标准：
- **必须使用**：大规模变更（> 5 文件）、破坏性变更、架构重构
- **可选使用**：中等规模变更（2-5 文件）
- **不需要**：小规模变更（单文件、配置参数）

---

## 总结

渐进式迁移的核心是：
- 🚀 **小步快跑**：每次只改一点
- ✅ **快速验证**：每步 < 30 分钟
- ↩️ **可回滚**：失败立即回滚
- 📊 **可追踪**：清晰的 Git 历史
- 🎯 **风险可控**：单次影响 < 5%

遵循本模板，你可以将高风险的大规模变更转变为低风险的系列小变更。
