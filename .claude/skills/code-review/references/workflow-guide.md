# 代码审查详细工作流程

本文档提供详细的代码审查步骤和检查清单。

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

### 步骤 2: 语法错误检查（最高优先级）

对于 Go 项目：
```bash
go build ./path/to/package/...
go vet ./path/to/package/...
```

使用 ReadLints 检查文件的 linter 错误。

**为什么优先检查语法错误？**
- 阻止编译的问题必须首先解决
- 避免在语法错误的代码上浪费时间审查逻辑

### 步骤 3: 项目规范检查

**根据文件类型和模块，参考相应的规范文档：**

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

### 步骤 4: 逻辑一致性检查

**模块特定规范：**
- `internal/backend/dpmgr/executor.go` → [../patterns/dpmgr-executor.md](../patterns/dpmgr-executor.md)
- 其他模块规范参见 [../patterns/README.md](../patterns/README.md)

**一般检查项：**
- 函数签名是否与相似函数一致
- 错误处理模式是否统一
- 日志格式是否一致
- 数据结构构建是否符合模式

**如何进行一致性检查：**
1. 使用 `serena.find_symbol`（substring_matching=true）查找相似函数，降级时用 Grep
2. 使用 `serena.search_for_pattern` 查找相关模式，降级时用 Grep
3. 比对函数签名、错误处理、日志格式
4. 检查数据结构构建模式

### 步骤 5: 代码质量检查

**核心规范：**
- 遵循 `.golangci.yml` 规则
- 遵循 `.cursor/rules/conv.mdc` 数据转换规则
- 参考 `AGENTS.md` 和 `CLAUDE.md` 项目规范

**检查清单：**
- [ ] 错误处理是否完整（使用 `fmt.Errorf` 和 `%w`）
- [ ] 是否遵循项目代码风格（导入别名、命名、注释）
- [ ] 是否使用了 `pkg/runtime/conv` 进行数据转换
- [ ] 日志记录是否使用结构化日志
- [ ] 公共函数是否有完整注释

**参考 [go-standards.md](go-standards.md) 了解详细的 Go 项目规范。**

### 步骤 6: 潜在问题检查

关注以下问题：
- **性能问题**: 不必要的循环、重复计算、资源浪费
- **安全问题**: SQL 注入、XSS、敏感信息泄露
- **并发问题**: 竞态条件、死锁、goroutine 泄露
- **资源泄露**: 文件句柄、数据库连接、内存泄露

## 审查 Commit 的特殊流程

审查 commit 时：
1. 使用 `git show <commit-hash>` 查看变更
2. 识别变更的文件类型和模块
3. 读取修改的文件（完整内容，了解上下文）
4. 执行编译和静态检查
5. 对比相似代码的模式
6. 参考相关规范文档
7. 生成审查报告（参考 [report-template.md](report-template.md)）

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

1. **语法错误** - 阻止编译的问题，最高优先级
2. **安全问题** - 可能导致安全漏洞的问题
3. **逻辑错误** - 影响功能正确性的问题
4. **规范违反** - 不符合项目规范的问题
5. **代码质量** - 可读性、可维护性改进

## 注意事项

- **保持客观**: 基于规范和最佳实践提供反馈，避免主观判断
- **提供上下文**: 引用具体的规范文档和行号
- **建设性反馈**: 不仅指出问题，还提供解决方案
- **优先级明确**: 区分必须修复和建议改进
- **简洁明了**: 避免冗长的解释，直接指出问题和修复方法
