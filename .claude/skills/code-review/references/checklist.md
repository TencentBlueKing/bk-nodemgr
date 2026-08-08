# 代码审查检查清单

本文档提供可操作、可验证的代码审查检查清单。所有检查项都应该在代码审查过程中逐一验证。

---

## 〇、AGENTS.md 合规性（强制）

AGENTS.md 是项目铁律。以下检查项在步骤 2（Design 审查）和步骤 4（规范检查）中执行，违反即为阻塞性或重要问题。

### 0.1 Anti-patterns（违反 → ❌ 严重）
- [ ] **不手编 generated 代码**：未手动修改 `*.pb.go` 文件
- [ ] **Proto 仅作 boundary type**：proto struct 在 `pkg/proto/*` 转换后才进入业务逻辑，业务层不直接操作 proto struct
- [ ] **复用优先 / 不引入重复**：新增代码优先扩展已有 code path/helper/proto 转换；新增的函数在 `pkg/proto/`、`pkg/runtime/conv` 或同层已有代码中无功能重复（同时覆盖 3B.1 DRY 原则）
- [ ] **service 逻辑不放 pkg**：service-specific 逻辑在 `internal/<service>` 而非 `pkg`

### 0.2 Conventions（违反 → ⚠️ 重要）
- [ ] **先读后写**：编码前已阅读相关模块 + 至少一个同服务/同层的类似实现（通过代码模式一致性间接验证）
- [ ] **公共类型英文注释**：导出的 Go 函数/类型有英文文档注释（详细格式要求参见 2.4）
- [ ] **依赖方向**：代码放置符合 `Where to look` 约定——service startup 在 `cmd/*`，router 在 `internal/*/router/api-v3`，DAO 在 `pkg/dao`，proto 转换在 `pkg/proto/**`
- [ ] **使用项目标准构建命令**：通过 `make pre` / `make lint` / `make all` 验证（而非自行拼凑的 go build/vet 命令）

---

## 一、语法检查（阻塞性问题）

这些是必须通过的检查项，任何失败都会阻止代码合并。

### 1.1 编译与静态分析
- [ ] **`make pre` 或 `make lint` 通过**：使用项目标准构建入口验证（与 0.2 一致）
  - 预期结果：无编译错误、无 vet 警告
  - 常见问题：导入路径错误、类型不匹配、未定义的函数/变量、可疑的函数调用、不正确的 Printf 格式

### 1.2 Linter 检查
- [ ] **Linter 无错误**：使用 ReadLints 工具检查代码规范
  - 验证方式：运行 ReadLints 查看 linter 输出
  - 预期结果：无 linter 错误（警告需要评估）
  - 配置文件：`.golangci.yml`
  - 常见问题：命名不规范、未导出的注释缺失、复杂度过高

---

## 二、规范一致性检查

这些检查确保代码符合项目规范和最佳实践。

### 2.1 错误处理规范
- [ ] **加载项目 owner skill**：错误创建、传播、包装、REST 映射、log-or-return、`err == nil` 迁移以 `bk-nodemgr-error-handling` 为准
  - 验证方式：审查涉及 `err`、`fmt.Errorf`、`resterrf.ErrWrap`、`WithErr`、`errors.Is/As`、panic/recover 时，先对照该 skill
  - 参考规范：`references/go-standards.md#错误处理规范`、`.claude/skills/bk-nodemgr-error-handling/SKILL.md`

- [ ] **禁止 `err == nil`**：新增或修改的 Go 代码不得出现 `err == nil`
  - 验证方式：`rg -n "err\s*==\s*nil" --glob '*.go'`
  - diff 新增/修改命中：阻塞性问题，必须改
  - 历史存量命中：记录为 migration debt，不因普通 PR 自动扩大范围
  - 禁止绕过：`isNilErr(err)`、`ok := err == nil`、重命名变量、`nolint`、helper 马甲
  - 审查要求：先按 migration catalog 判断语义，再迁移；禁止机械替换

- [ ] **错误包装使用 `%w` 格式**：所有错误链使用 `fmt.Errorf("...: %w", err)`
  - 验证方式：Grep 搜索 `fmt.Errorf` 确认使用 `%w` 而非 `%v` 或 `%s`
  - 错误示例：`fmt.Errorf("failed: %v", err)` ❌
  - 正确示例：`fmt.Errorf("failed to create: %w", err)` ✅

- [ ] **链式错误检查安全**：wrapped/joined error 使用 `errors.Is` / `errors.As`
  - 错误示例：`if err == ErrNotFound` ❌
  - 正确示例：`if errors.Is(err, ErrNotFound)` ✅

- [ ] **REST 边界映射稳定**：API handler 使用 `resterrf.ErrWrap(code, err)`，内部层只加 `%w` 上下文
  - 验证方式：对比同 router group 的 handler，确认错误码语义一致
  - 错误示例：内部 service 重复 `ErrWrap` 或 handler 直接返回 raw error ❌

- [ ] **单一处理责任**：同一条 error chain 不应在多个层级重复 `ERROR` 日志
  - 验证方式：检查 lower layer 是否只 wrap+return，边界层是否才 `WithErr(err).Error(...)`
  - 错误示例：DAO、service、handler 连续 log 同一错误 ❌

- [ ] **panic/recover 边界合理**：业务请求路径不使用 panic/recover 表达预期失败
  - 正确方式：返回 error，由 REST 或系统边界统一映射/记录

### 2.2 日志规范（框架 + 格式 + 数量）

本节统一覆盖日志相关的所有检查，不再散落于多个章节。

**(a) 框架选择** [AGENTS.md Conventions]
- [ ] **使用 `pkg/logger`**：所有日志通过 `pkg/logger` 记录，不使用 `fmt.Println`、`log.Printf` 等

**(b) 格式规范**
- [ ] **两行结构化格式**：
  - 第一行：使用 `.With()` 添加结构化字段
  - 第二行：调用 `.Info()` / `.Error()` 等方法记录消息
  - 验证方式：检查所有 `logger.G.Sys()` 调用
  - 错误示例：`logger.G.Sys().With("id", id).Info("done")` ❌（单行格式）
  - 正确示例：
    ```go
    logger.G.Sys().With("workflow-id", id).
        Info("workflow completed")
    ```

**(c) 数量控制（缄默原则）** [architecture-principles.md 3.3]
- [ ] **正常路径无冗余日志**：非错误分支不应打 Info/Debug 级别日志（除关键业务节点）
- [ ] **循环体内无高频日志**：避免在循环中每次迭代都打日志
- [ ] **调试日志不残留**：调试阶段的临时日志应在提交前移除或降为 Debug 级别

  - 参考规范：`references/go-standards.md#日志记录规范`

### 2.3 数据转换规范
- [ ] **优先使用 `pkg/runtime/conv` 包**：常见数据转换应使用项目提供的工具函数
  - 验证方式：检查手动循环或类型转换代码，评估是否可用 conv 包替代
  - 常见场景：
    - Map keys 转 Slice：使用 `conv.MapKeyToSlice()`
    - 类型转换：使用 `conv.To*()` 系列函数
    - 字符串处理：使用 `conv.String*()` 系列函数
  - 参考规范：`.cursor/rules/conv.mdc`

### 2.4 注释规范 [含 AGENTS.md 要求]
- [ ] **公共函数有完整英文注释**：所有导出的函数、类型、常量都应有**英文**文档注释（AGENTS.md Conventions 要求）
  - 验证方式：检查所有首字母大写的标识符
  - 注释格式：
    - 以标识符名称开头
    - 使用英文撰写
    - 说明功能、参数含义、返回值
    - 特殊情况（如线程安全性、性能注意事项）需说明
  - 错误示例：无注释或 `// Create function` ❌ 或中文注释 `// 创建部署` ❌
  - 正确示例：
    ```go
    // CreateDeployment creates a new deployment with the given configuration.
    // It returns the deployment ID and an error if the creation fails.
    func CreateDeployment(config *Config) (string, error)
    ```

### 2.5 命名规范
- [ ] **命名符合 Go 规范**：
  - 错误类型：以 `Err` 开头（如 `ErrNotFound`）或以 `Error` 结尾（如 `ValidationError`）
  - 变量名：使用驼峰命名，避免下划线
  - 常量：使用驼峰命名或全大写（根据上下文）
  - 接口：单方法接口以 `-er` 结尾（如 `Reader`, `Writer`）
  - 验证方式：Review 所有新增的类型、变量、常量命名
  - 参考规范：`references/go-standards.md#命名规范`

---

## 三、逻辑一致性检查

这些检查确保新代码与现有代码保持一致的实现模式。

### 3.1 函数签名一致性
- [ ] **函数签名与相似函数一致**：
  - 验证方式：使用 Grep 查找同模块中的相似函数（如 `CreateAgent`, `CreatePlugin`）
  - 检查项：
    - 参数顺序和类型是否一致
    - 返回值类型和顺序是否一致
    - Context 参数位置是否符合惯例（通常为第一个参数）
    - 错误返回是否在最后一个位置
  - 示例：如果 `CreateAgent(ctx context.Context, name string) error`，
    则 `CreatePlugin` 应该是 `(ctx context.Context, name string) error` 而非 `(name string, ctx context.Context) error`

### 3.2 错误消息一致性
- [ ] **错误消息格式与模块其他函数一致**：
  - 验证方式：Grep 搜索同模块的错误消息模式
  - 检查项：
    - 消息前缀是否一致（如都以 "failed to" 开头）
    - 动词时态是否一致
    - 参数插入格式是否一致
  - 错误示例：模块中其他函数使用 "failed to create agent"，新代码使用 "create plugin error" ❌
  - 正确示例：统一使用 "failed to create plugin" ✅

### 3.3 日志格式一致性
- [ ] **日志格式与相似代码一致**：
  - 验证方式：Grep 搜索同模块的日志记录代码
  - 检查项：
    - 字段名称约定（如使用 `workflow-id` 还是 `workflowId`）
    - 日志级别选择（Info vs Debug）
    - 消息动词时态（completed vs complete）
  - 错误示例：模块中使用 "workflow-id"，新代码使用 "workflow_id" ❌
  - 正确示例：统一使用 "workflow-id" ✅

### 3.4 数据结构构建一致性
- [ ] **数据结构构建模式一致**：
  - 验证方式：对比同模块中相似的数据结构构建代码
  - 检查项：
    - 结构体字段赋值顺序
    - 使用字面量初始化还是逐字段赋值
    - 可选字段处理方式
    - 默认值设置模式
  - 示例：如果模块中统一使用结构体字面量初始化，新代码也应遵循该模式

### 3.5 边界情况
- [ ] **边界情况**：以 API 消费者视角审视——nil 输入、空集合、零值、超大输入是否被合理处理

> 💭 **审查时留意**（复杂度参考提示，非硬性检查项）：
>
> 以下模式可能暗示过度工程化，linter 无法检测，需要设计层面的判断：
> - **YAGNI 倾向**：是否引入了当前不需要的抽象或扩展点？"以后可能用到"不是充分理由
> - **过度泛化**：是否为单一用例创建了过于通用的接口或框架？
> - **冗余抽象（透传 wrapper）**：是否存在仅透传调用的 wrapper 函数/结构，未增加任何逻辑？

---

## 三-B、架构与设计原则检查

这些检查基于团队沉淀的架构原则，在步骤 5（一致性检查）和步骤 6（质量检查）中执行。
详细说明参见 [architecture-principles.md](architecture-principles.md)。

### 3B.1 DRY — 不要重复
> 0.1 已作为 AGENTS.md 强约束前置检查了 helper/proto 转换的重复。本项是更广义的 DRY 审视。
- [ ] **无功能重叠**：新增代码未与已有代码重复实现相同逻辑（用 `serena.find_symbol` 搜索验证）

### 3B.2 正交性
- [ ] **无不必要的全局变量**：未引入包级全局变量（除非有充分理由）
- [ ] **无不合理的跨模块依赖**：修改一个模块不需要同时修改不相关模块

### 3B.3 单层级函数
- [ ] **抽象层级统一**：函数体内不混合高层编排和底层细节

### 3B.4 显式处理
- [ ] **无静默忽略**：忽略返回值使用 `_ = F()` 而非裸调用；空实现有注释说明
- [ ] **无 shadow variable**：`:=` 未意外遮蔽外层同名变量

### 3B.5 控制熵
- [ ] **遵循既定范式**：对同类问题使用项目已有的实现模式，未引入"另一种做法"
- [ ] **无 hardcode**：无硬编码的 IP、端口、魔数

### 3B.6 缄默原则
> 详细检查项已合并至 2.2(c)。此处仅作为架构原则提醒——过多正常日志 = 没有日志。

> 💭 **案例辅助**：如果发现的问题与 `cases/` 中的已有案例相似，在报告中引用案例编号以增强说服力。

---

## 四、测试质量检查

> 当 PR 包含 `_test.go` 文件时执行本节检查。如果变更不涉及测试文件，跳过本节。

### 4.1 Assertion 有效性
- [ ] **Assertion 有效性**：测试中的断言是否真正验证了被测行为？
  - 禁止手写 `err == nil` 断言：使用 `require.NoError` / `assert.NoError` / `require.Error` / `assert.Error` 或项目现有断言风格，并继续验证返回值核心契约
  - 警惕过于宽泛的断言：如 `assert.NotNil(result)` 而未检查 result 的具体字段
  - 验证方式：对每个测试函数，检查其断言是否覆盖了被测函数的核心契约

### 4.2 假阳性风险
- [ ] **假阳性风险**：测试是否可能在被测代码有 bug 时仍然通过？
  - 检查是否有被注释掉或被跳过的断言
  - 检查 error path 测试是否真正触发了错误（而非 happy path 碰巧也通过）
  - 检查 mock 是否过于宽松导致任何输入都返回成功

### 4.3 Mock 合理性
- [ ] **Mock 合理性**：mock/stub 是否合理反映了真实依赖的行为？
  - mock 不应掩盖被测代码的实际依赖行为
  - mock 的返回值和错误场景应覆盖真实场景
  - 避免过度 mock：如果 mock 了大部分依赖，测试可能失去意义

### 4.4 测试可维护性
- [ ] **测试可维护性**：测试代码是否易于理解和维护？
  - 测试名称是否清晰描述了被测场景
  - 测试数据是否使用了 table-driven 模式（Go 惯例）
  - 测试辅助函数是否适度抽象（不过度也不过少）

---

## 使用指南

### 快速审查流程
1. ✅ 先完成"〇、AGENTS.md 合规性"所有项（**强制，步骤 2 门禁**）
2. ✅ 完成"一、语法检查"所有项（阻塞性）
3. ✅ 完成"二、规范一致性检查"所有项（重要性）
4. ✅ 完成"三、逻辑一致性检查"所有项（质量保证）
5. ✅ 完成"三-B、架构与设计原则检查"所有项（设计质量）
6. ✅ 如有 `_test.go` 文件，完成"四、测试质量检查"所有项

### 工具使用

MCP 工具优先，传统工具作为降级方案：

- **编译/静态分析**：使用 Shell 工具执行 `go build` 和 `go vet`（无可替代）
- **Linter 检查**：使用 ReadLints 工具（无可替代）
- **代码结构**：优先 `serena.get_symbols_overview`，降级使用 Read
- **符号查找**：优先 `serena.find_symbol`，降级使用 Grep
- **模式搜索**：优先 `serena.search_for_pattern`，降级使用 Grep
- **依赖分析**：`serena.find_referencing_symbols`（无可替代）
- **结构化思考**：`sequential-thinking`（规划策略和生成报告时必用）

### 问题优先级
- **阻塞性问题**（AGENTS.md Anti-patterns 违反、语法检查未通过）：❌ 必须修复
- **重要问题**（AGENTS.md Conventions 违反、规范一致性问题）：⚠️ 强烈建议修复
- **质量问题**（逻辑一致性、架构原则问题）：⚠️ 建议修复
- **改进建议**（可选优化）：💡 建议

---

## 相关文档
- 详细工作流：[workflow-guide.md](workflow-guide.md)
- Go 规范速查：[go-standards.md](go-standards.md)
- 报告模板：[report-template.md](report-template.md)
- 工具使用指南：[tools-usage.md](tools-usage.md)
