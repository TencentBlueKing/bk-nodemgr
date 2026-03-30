# 代码审查详细工作流程

本文档提供详细的代码审查步骤和检查清单。

## 审查状态管理

审查过程中使用 `.review/` 文件夹持久化审查状态，支持跨会话续审和门禁终止后恢复。

### 文件结构

```
.review/
├── checklist.md        # 审查进度：各步骤 [x]/[ ] 状态 + 已发现问题摘要
└── meta.json           # 元数据：审查目标（commit/branch）、开始时间、当前步骤
```

### `checklist.md` 职责

- 实时记录审查进度和已发现的问题（每个步骤完成后更新）
- 最终报告（步骤 7）从 checklist 的 Issues Found 部分汇总生成，而非独立重新收集
- 中途终止（门禁触发或会话中断）时，已发现的问题不丢失

**checklist.md 示例：**
```markdown
# Review Checklist

Target: feature/add-iam-auth (commit abc1234)
Started: 2026-03-02T10:30:00Z

## Progress

- [x] 1. 确定范围 — 3 files, internal/backend/
- [x] 2. Design 审查 — ⛔ 发现设计问题，用户选择终止
- [ ] 3. 语法检查
- [ ] 4. 规范检查
- [ ] 5. 一致性检查
- [ ] 6. 质量检查
- [ ] 7. 生成报告

## Issues Found

### ❌ 严重（Design）
- `pkg/service/auth.go` — service 逻辑不应放在 `pkg`，应移至 `internal/backend/service/`

### ⚠️ 重要
（续审时继续补充）
```

### `meta.json` 结构

```json
{
  "target": "feature/add-iam-auth",
  "commit": "abc1234",
  "startedAt": "2026-03-02T10:30:00Z",
  "currentStep": 4,
  "totalSteps": 7
}
```

### 续审检测

审查开始时检查 `.review/` 是否存在：

| 场景 | 行为 |
|------|------|
| `.review/` 不存在 | 创建 `checklist.md` 和 `meta.json`，从步骤 1 开始 |
| branch 名不匹配 | 清空 `.review/` 重新开始 |
| branch + commit 均匹配 | 提示"检测到上次审查进度，是否从步骤 N 继续？" |
| branch 匹配但 commit 变化 | 提示"commit 已变更，是否继续还是重新开始？" |

### 门禁终止时的记录

- 门禁步骤标记 `[x]` 并附注"⛔ 发现设计问题，用户选择终止"
- 后续步骤保持 `[ ]`
- Issues Found 中记录发现的问题
- 续审时 agent 先检查门禁步骤记录的问题是否已修复，再决定继续

### 每步骤更新

每个步骤完成后：
1. 更新 `checklist.md` 中对应步骤的状态（`[ ]` → `[x]`）
2. 将该步骤发现的问题追加到 Issues Found 部分
3. 更新 `meta.json` 的 `currentStep`

---

## 审查步骤详解

### 步骤 1: 确定审查范围

首先确定需要审查的文件：
```bash
git status -uno  # 查看已暂存的提交文件
```

**识别文件类型和模块：**

根据文件路径判断：
- `proto/*.proto` → Proto 文件规范
- `internal/backend/dpmgr/` → dpmgr-executor 模式
- `pkg/dao/` → 数据访问层规范
- `pkg/rest/` → API 框架规范
- API 相关代码 → API 开发流程

### 步骤 2: Design 审查 + AGENTS.md 门禁（阻塞性）

在理解审查范围之后、进入语法检查之前，评估代码的设计合理性 **并检查 AGENTS.md 合规性**。Design 问题和 AGENTS.md Anti-pattern 违反比语法错误更根本——如果代码放错了位置或引入了不合理的依赖，后续的修复都无意义。

**使用工具：**
- `🧠 sequential-thinking` — 结构化评估设计合理性，逐项分析代码放置、依赖方向、抽象层级
- `🔍 serena.get_symbols_overview` — 查看变更文件的符号结构，判断代码是否放在了正确的模块
- `🔍 serena.find_referencing_symbols` — 检查新增代码的依赖方向，确认无反向依赖（`pkg` 不应依赖 `internal`）
- `Read AGENTS.md` — 必须读取项目根目录 AGENTS.md，获取最新的 Anti-patterns 和 Where to look 约定

**检查内容：**

**A. AGENTS.md Anti-patterns（违反即阻塞 ❌）：**
- 是否手动修改了 `*.pb.go` 文件？
- proto struct 是否直接用于业务逻辑而未通过 `pkg/proto/*` 转换？
- 是否引入了 duplicate helper / parallel conversion logic？（用 `serena.find_symbol` 搜索同名或功能相似的函数）
- service-specific 逻辑是否错误地放在了 `pkg/` 而非 `internal/<service>/`？

**B. 代码放置位置（基于 AGENTS.md Where to look）：**
- service startup 代码是否在 `cmd/*`？
- router/handler 是否在 `internal/*/router/api-v3`？
- DAO 是否在 `pkg/dao/mongo` 或 `internal/*/storage`？
- proto 转换是否在 `pkg/proto/**`？

**C. 模块依赖合理性：**
- `pkg/` 是否 import 了 `internal/` 下的包？（严重 ❌）
- 是否引入了不合理的跨模块依赖？
- 新增的接口/结构是否在合理的抽象层级上？

**D. 架构原则（参见 [architecture-principles.md](architecture-principles.md)）：**
- DRY：是否与已有代码功能重叠？
- 正交性：是否引入了不必要的耦合？
- 可逆性：设计决策是否允许后续变更？

**门禁行为：**

- **无设计问题**：在 `.review/checklist.md` 中标记步骤完成，继续执行后续步骤
- **发现设计问题或 AGENTS.md 违反**：
  1. 生成**早期报告**（包含设计问题和 AGENTS.md 违反项）
  2. 提示用户："发现设计层面问题 / AGENTS.md 违反，建议先修正后重新提交审查。是否继续后续检查？"
  3. **用户选择终止**：输出早期报告，在 checklist 中记录"⛔"标记和问题，结束审查
  4. **用户选择继续**：问题记录到 Issues Found，照常执行后续步骤 3-7

### 步骤 3: 语法错误检查（阻塞性门禁）

对于 Go 项目：
```bash
go build ./path/to/package/...
go vet ./path/to/package/...
```

使用 ReadLints 检查文件的 linter 错误。

**为什么优先检查语法错误？**
- 阻止编译的问题必须首先解决
- 避免在语法错误的代码上浪费时间审查逻辑

### 步骤 4: 项目规范检查 + AGENTS.md Conventions

**首先检查 AGENTS.md Conventions（违反 → ⚠️ 重要）：**
- 导出的 Go 函数/类型是否有英文文档注释？
- 是否使用 `pkg/logger` 记录日志？
- 是否优先扩展已有 code path 而非新建平行实现？
- 前端代码是否遵循 `pnpm@9.8.0` + `@blueking/eslint-config-bk/tsvue3` 规范？

**然后根据文件类型和模块，参考相应的规范文档：**

#### API 接口开发
- 文件路径包含 `proto/` 或 API 相关代码
- **参考文档**: `docs/api/API接口开发流程.md`
- **检查项**:
  - API 接口命名和设计规范
  - 请求/响应结构定义
  - 错误码和错误处理

#### Proto 文件
- 文件后缀为 `.proto`
- **参考文档**: `proto/README.md`
- **检查项**:
  - 命名规范（service、message、field）
  - 字段编号管理
  - 注释完整性

#### 开发规范
- **参考文档**: `docs/developer/README.md`
- **检查项**:
  - 代码风格和命名规范
  - 错误处理模式
  - 日志记录规范

### 步骤 5: 逻辑一致性检查

**模块特定规范：**
- `internal/backend/dpmgr/executor.go` → [../patterns/dpmgr-executor.md](../patterns/dpmgr-executor.md)
- 其他模块规范参见 [../patterns/README.md](../patterns/README.md)

**一般检查项：**
- 函数签名是否与相似函数一致
- 错误处理模式是否统一
- 日志格式是否一致
- 数据结构构建是否符合模式

**架构原则检查（参见 [architecture-principles.md](architecture-principles.md)）：**
- DRY：搜索是否存在功能重复的代码
- 控制熵：新代码是否遵循同模块的既定范式
- 统一术语：新增的命名是否与已有约定一致
- 显式处理：忽略返回值是否显式标注

**案例库参考：**
- 检查 `cases/` 目录中是否有与当前发现类似的案例
- 如有匹配案例，在报告中引用：`参见案例 cases/XXX.md`

**如何进行一致性检查：**
1. 使用 `serena.find_symbol`（substring_matching=true）查找相似函数，降级时用 Grep
2. 使用 `serena.search_for_pattern` 查找相关模式，降级时用 Grep
3. 比对函数签名、错误处理、日志格式
4. 检查数据结构构建模式

### 步骤 6: 代码质量检查

**核心规范：**
- 遵循 `.golangci.yml` 规则
- 遵循 `.cursor/rules/conv.mdc` 数据转换规则
- **遵循 `AGENTS.md` 约定（Conventions 部分，步骤 4 已检查 Anti-patterns）**

**架构原则检查（参见 [architecture-principles.md](architecture-principles.md)）：**
- [ ] 单层级函数：函数体内抽象层级是否统一
- [ ] 只管命令不要询问：是否先 Get 再 Set 同一对象
- [ ] 传入最小参数集：函数参数是否过于宽泛
- [ ] 缄默原则：正常路径是否有不必要的日志
- [ ] 尽早崩溃：是否有 `recover()` 后空操作、静默忽略错误

**常规检查清单：**
- [ ] 错误处理是否完整（使用 `fmt.Errorf` 和 `%w`）
- [ ] 是否遵循项目代码风格（导入别名、命名、注释）
- [ ] 是否使用了 `pkg/runtime/conv` 进行数据转换
- [ ] 日志记录是否使用结构化日志
- [ ] 公共函数是否有完整注释

**参考 [go-standards.md](go-standards.md) 了解详细的 Go 项目规范。**

### 步骤 7: 生成报告

**报告前检查——Documentation 同步：**

在生成最终报告前，快速扫描变更是否涉及需要同步更新的关联文档：
- 变更涉及 proto 定义或 API 行为 → 检查 `docs/` 和 proto README 是否已更新
- 变更涉及公共接口或项目结构 → 检查 `AGENTS.md` 是否需要更新
- 若相关文档未随代码变更，在报告审查总结中添加文档同步提醒（提醒式，非阻塞）

**报告生成：**

从 `.review/checklist.md` 的 Issues Found 汇总生成最终报告，按优先级分类：
- ❌ 严重问题（必须修复）
- ⚠️ 重要问题（强烈建议修复）
- 💡 建议改进（可选）

关注以下潜在问题：
- **性能问题**: 不必要的循环、重复计算、资源浪费
- **安全问题**: SQL 注入、XSS、敏感信息泄露
- **并发问题**: 竞态条件、死锁、goroutine 泄露
- **资源泄露**: 文件句柄、数据库连接、内存泄露

参考 [report-template.md](report-template.md) 了解完整报告模板。

## 审查 Commit 的特殊流程

审查 commit 时：
1. 使用 `git show <commit-hash>` 查看变更
2. 识别变更的文件类型和模块
3. 评估代码设计（放置位置、依赖合理性）
4. 读取修改的文件（完整内容，了解上下文）
5. 执行编译和静态检查
6. 对比相似代码的模式
7. 参考相关规范文档
8. 生成审查报告（参考 [report-template.md](report-template.md)）

## 常见审查场景

### 场景 1: 审查新增的 API 接口

1. 读取 `docs/api/API接口开发流程.md`
2. 检查 proto 文件定义（如有）
3. 验证错误码定义和 HTTP 状态码映射
4. 检查认证和授权实现
5. 验证请求/响应结构

### 场景 2: 审查 executor 模式代码

1. 读取 `../patterns/dpmgr-executor.md`
2. 识别函数系列（Agent/Plugin/PluginPkg）
3. 对比同系列的其他函数
4. 检查日志格式、错误消息、数据结构构建
5. 验证 Manager 调用是否正确

### 场景 3: 审查数据转换代码

1. 检查是否使用了 `pkg/runtime/conv` 包
2. 如果手动实现转换，评估是否必要
3. 验证错误处理是否完整
4. 建议使用 conv 包的合适函数（参考 [go-standards.md](go-standards.md)）

## 审查技巧

### 高效使用工具

MCP 工具优先，传统工具作为降级方案：

- **sequential-thinking**: 规划审查策略、分析复杂问题、生成结构化报告（每次审查开始和结束必用）
- **serena.get_symbols_overview**: 快速了解文件结构，替代全文 Read（节省 75% token）
- **serena.find_symbol**: 精确查找函数/类/方法定义，替代 Grep
- **serena.find_referencing_symbols**: 分析代码依赖和影响范围（无可替代）
- **serena.search_for_pattern**: 语义级模式搜索，替代 Grep
- **ReadLints**: 优先检查 linter 错误，无可替代
- **Shell**: 编译检查和 Git 操作，无可替代
- **Grep / Read**: 降级方案，仅在 serena 无法满足时使用

### 审查优先级

1. **设计问题** - 代码放置错误、依赖方向不合理，最高优先级
2. **语法错误** - 阻止编译的问题
3. **安全问题** - 可能导致安全漏洞的问题
4. **逻辑错误** - 影响功能正确性的问题
5. **规范违反** - 不符合项目规范的问题
6. **代码质量** - 可读性、可维护性改进

## 注意事项

- **保持客观**: 基于规范和最佳实践提供反馈，避免主观判断
- **提供上下文**: 引用具体的规范文档和行号
- **建设性反馈**: 不仅指出问题，还提供解决方案
- **优先级明确**: 区分必须修复和建议改进
- **简洁明了**: 避免冗长的解释，直接指出问题和修复方法
