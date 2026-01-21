# 代码审查检查清单

本文档提供可操作、可验证的代码审查检查清单。所有检查项都应该在代码审查过程中逐一验证。

---

## 一、语法检查（阻塞性问题）

这些是必须通过的检查项，任何失败都会阻止代码合并。

### 1.1 编译检查
- [ ] **编译通过**：运行 `go build ./path/...` 确保代码能够成功编译
  - 验证方式：在项目根目录或相关模块目录执行编译命令
  - 预期结果：无编译错误输出
  - 常见问题：导入路径错误、类型不匹配、未定义的函数/变量

### 1.2 静态分析
- [ ] **静态分析无错误**：运行 `go vet ./path/...` 检查潜在问题
  - 验证方式：对变更的包执行 vet 命令
  - 预期结果：无 vet 警告或错误
  - 常见问题：可疑的函数调用、不正确的 Printf 格式、未使用的赋值

### 1.3 Linter 检查
- [ ] **Linter 无错误**：使用 ReadLints 工具检查代码规范
  - 验证方式：运行 ReadLints 查看 linter 输出
  - 预期结果：无 linter 错误（警告需要评估）
  - 配置文件：`.golangci.yml`
  - 常见问题：命名不规范、未导出的注释缺失、复杂度过高

---

## 二、规范一致性检查

这些检查确保代码符合项目规范和最佳实践。

### 2.1 错误处理规范
- [ ] **错误包装使用 `%w` 格式**：所有错误链使用 `fmt.Errorf("...: %w", err)`
  - 验证方式：Grep 搜索 `fmt.Errorf` 确认使用 `%w` 而非 `%v` 或 `%s`
  - 错误示例：`fmt.Errorf("failed: %v", err)` ❌
  - 正确示例：`fmt.Errorf("failed to create: %w", err)` ✅
  - 参考规范：`references/go-standards.md#错误处理规范`

### 2.2 日志记录规范
- [ ] **日志使用两行结构化格式**：
  - 第一行：使用 `.With()` 添加结构化字段
  - 第二行：调用 `.Info()` / `.Error()` 等方法记录消息
  - 验证方式：检查所有 `logger.G.Sys()` 调用
  - 错误示例：`logger.G.Sys().With("id", id).Info("done")` ❌（单行格式）
  - 正确示例：
    ```go
    logger.G.Sys().With("workflow-id", id).
        Info("workflow completed")
    ```
  - 参考规范：`references/go-standards.md#日志记录规范`

### 2.3 数据转换规范
- [ ] **优先使用 `pkg/runtime/conv` 包**：常见数据转换应使用项目提供的工具函数
  - 验证方式：检查手动循环或类型转换代码，评估是否可用 conv 包替代
  - 常见场景：
    - Map keys 转 Slice：使用 `conv.MapKeyToSlice()`
    - 类型转换：使用 `conv.To*()` 系列函数
    - 字符串处理：使用 `conv.String*()` 系列函数
  - 参考规范：`.cursor/rules/conv.mdc`

### 2.4 注释规范
- [ ] **公共函数有完整注释**：所有导出的函数、类型、常量都应有文档注释
  - 验证方式：检查所有首字母大写的标识符
  - 注释格式：
    - 以标识符名称开头
    - 说明功能、参数含义、返回值
    - 特殊情况（如线程安全性、性能注意事项）需说明
  - 错误示例：无注释或 `// Create function` ❌
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

---

## 使用指南

### 快速审查流程
1. ✅ 先完成"一、语法检查"所有项（阻塞性）
2. ✅ 再完成"二、规范一致性检查"所有项（重要性）
3. ✅ 最后完成"三、逻辑一致性检查"所有项（质量保证）

### 工具使用
- **编译/静态分析**：使用 Shell 工具执行 `go build` 和 `go vet`
- **Linter 检查**：使用 ReadLints 工具
- **代码搜索**：使用 Grep 工具查找相似代码模式
- **文件读取**：使用 Read 工具查看具体实现

### 问题优先级
- **阻塞性问题**（语法检查未通过）：❌ 必须修复
- **重要问题**（规范一致性问题）：⚠️ 强烈建议修复
- **质量问题**（逻辑一致性问题）：⚠️ 建议修复

---

## 相关文档
- 详细工作流：[workflow-guide.md](workflow-guide.md)
- Go 规范速查：[go-standards.md](go-standards.md)
- 报告模板：[report-template.md](report-template.md)
- 工具使用指南：[tools-usage.md](tools-usage.md)
