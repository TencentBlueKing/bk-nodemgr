# 工具使用指南

本文档详细说明在渐进式提案设计、实施和验证过程中如何使用 MCP 工具。

## 工具概览

渐进式提案需要以下 MCP 工具支持:

1. **serena** - 语义级代码操作 (核心工具)
2. **context7** - 文档查询和上下文理解
3. **sequential-thinking** - 复杂决策分析
4. **Bash** - 命令执行和验证

---

## serena - 语义级代码操作

serena 是渐进式迁移的**核心工具**,提供语义级别的代码分析和修改能力。

### 核心工具函数

#### 1. `get_symbols_overview` - 获取文件符号概览

**用途**: 快速了解文件结构,无需读取完整文件内容

**参数**:
- `relative_path`: 目标文件相对路径
- `depth`: 符号深度 (0=仅顶层, 1=包含子符号)
- `max_answer_chars`: 最大返回字符数

**使用场景**:
- 分析文件包含哪些类/函数/方法
- 确定迁移单元的范围
- 了解文件的整体结构

**示例**:
```python
mcp__serena__get_symbols_overview(
    relative_path="pkg/workflow/node_install.go",
    depth=1
)
```

**输出**: 文件中的类、函数、方法列表及其签名

---

#### 2. `search_for_pattern` - 模式搜索

**用途**: 在代码库中搜索特定模式,支持正则表达式

**参数**:
- `substring_pattern`: 搜索模式 (支持正则)
- `relative_path`: 限制搜索路径 (可选)
- `restrict_search_to_code_files`: 是否仅搜索代码文件
- `context_lines_before/after`: 上下文行数

**使用场景**:
- 查找所有使用某个函数的位置
- 查找特定模式的代码 (如错误处理)
- 评估迁移影响范围

**示例**:
```python
# 查找所有使用 GetRelayInfo 的位置
mcp__serena__search_for_pattern(
    substring_pattern="GetRelayInfo",
    relative_path="pkg/",
    restrict_search_to_code_files=true,
    context_lines_before=2,
    context_lines_after=2
)
```

**输出**: 匹配的代码片段及其上下文

---

#### 3. `find_symbol` - 查找符号定义

**用途**: 精确查找符号 (类/函数/方法) 的定义

**参数**:
- `name_path_pattern`: 符号名称路径 (如 "ClassName/MethodName")
- `relative_path`: 限制搜索路径 (可选)
- `include_body`: 是否包含符号体 (代码实现)
- `depth`: 包含子符号的深度

**使用场景**:
- 查找函数/方法的完整定义
- 获取类的所有方法
- 分析符号的实现细节

**示例**:
```python
# 查找 NodeInstallWorkflow 类的所有方法
mcp__serena__find_symbol(
    name_path_pattern="NodeInstallWorkflow",
    relative_path="pkg/workflow/node_install.go",
    include_body=false,
    depth=1
)

# 查找特定方法的实现
mcp__serena__find_symbol(
    name_path_pattern="NodeInstallWorkflow/Execute",
    relative_path="pkg/workflow/node_install.go",
    include_body=true
)
```

**输出**: 符号定义及其元数据 (签名、位置、子符号等)

---

#### 4. `find_referencing_symbols` - 查找符号引用

**用途**: 查找所有引用某个符号的位置

**参数**:
- `name_path`: 符号名称路径
- `relative_path`: 符号所在文件
- `include_info`: 是否包含引用符号的详细信息

**使用场景**:
- 评估迁移影响范围 (哪些地方使用了这个函数)
- 确保所有引用都已更新
- 识别潜在的破坏性变更

**示例**:
```python
# 查找所有调用 GetRelayInfo 的位置
mcp__serena__find_referencing_symbols(
    name_path="GetRelayInfo",
    relative_path="pkg/dao/relay.go",
    include_info=true
)
```

**输出**: 所有引用该符号的位置及其代码片段

---

#### 5. `replace_symbol_body` - 替换符号体

**用途**: 精确替换函数/方法的实现

**参数**:
- `name_path`: 符号名称路径
- `relative_path`: 符号所在文件
- `body`: 新的符号体 (包含签名)

**使用场景**:
- 迁移单个函数的实现
- 更新方法逻辑
- 精确的符号级替换

**示例**:
```python
# 替换方法实现
mcp__serena__replace_symbol_body(
    name_path="NodeInstallWorkflow/Execute",
    relative_path="pkg/workflow/node_install.go",
    body="""func (w *NodeInstallWorkflow) Execute(ctx context.Context) error {
    // 新的实现
    return w.newExecuteLogic(ctx)
}"""
)
```

**输出**: 替换成功确认

---

#### 6. `insert_after_symbol` - 在符号后插入

**用途**: 在符号定义后插入新代码

**参数**:
- `name_path`: 参考符号名称路径
- `relative_path`: 符号所在文件
- `body`: 要插入的代码

**使用场景**:
- 添加新方法到类
- 在现有函数后添加辅助函数

**示例**:
```python
# 在类中添加新方法
mcp__serena__insert_after_symbol(
    name_path="NodeInstallWorkflow/Execute",
    relative_path="pkg/workflow/node_install.go",
    body="""
func (w *NodeInstallWorkflow) NewMethod(ctx context.Context) error {
    // 新方法实现
    return nil
}
"""
)
```

---

### 工具组合使用流程

#### 流程 1: 分析现状

```bash
# 步骤 1: 获取文件概览,了解整体结构
mcp__serena__get_symbols_overview(
    relative_path="pkg/workflow/node_install.go",
    depth=1
)

# 步骤 2: 查找特定符号的详细信息
mcp__serena__find_symbol(
    name_path_pattern="NodeInstallWorkflow",
    relative_path="pkg/workflow/node_install.go",
    include_body=false,
    depth=1
)

# 步骤 3: 搜索使用模式,评估影响范围
mcp__serena__search_for_pattern(
    substring_pattern="GetRelayInfo",
    relative_path="pkg/",
    restrict_search_to_code_files=true
)

# 步骤 4: 查找所有引用
mcp__serena__find_referencing_symbols(
    name_path="GetRelayInfo",
    relative_path="pkg/dao/relay.go",
    include_info=true
)
```

**输出**: 完整的影响范围分析报告

---

#### 流程 2: 迁移单个模块

```bash
# 步骤 1: 读取旧实现
mcp__serena__find_symbol(
    name_path_pattern="NodeInstallWorkflow/Execute",
    relative_path="pkg/workflow/node_install.go",
    include_body=true
)

# 步骤 2: 查找该方法的所有引用 (验证迁移后不会破坏)
mcp__serena__find_referencing_symbols(
    name_path="NodeInstallWorkflow/Execute",
    relative_path="pkg/workflow/node_install.go"
)

# 步骤 3: 替换方法实现
mcp__serena__replace_symbol_body(
    name_path="NodeInstallWorkflow/Execute",
    relative_path="pkg/workflow/node_install.go",
    body="[新实现]"
)

# 步骤 4: 验证其他引用是否需要更新
# (根据步骤 2 的结果逐个检查)
```

**输出**: 单个模块迁移完成

---

#### 流程 3: 验证迁移完整性

```bash
# 步骤 1: 搜索旧模式,确认是否还有残留
mcp__serena__search_for_pattern(
    substring_pattern="GetRelayInfo",  # 旧的函数名
    relative_path="pkg/",
    restrict_search_to_code_files=true
)

# 步骤 2: 搜索新模式,确认迁移是否完整
mcp__serena__search_for_pattern(
    substring_pattern="GetRelayListByFilter",  # 新的函数名
    relative_path="pkg/",
    restrict_search_to_code_files=true
)

# 步骤 3: 对比结果,确保所有旧引用都已迁移
```

**输出**: 迁移完整性验证报告

---

## context7 - 文档查询

**用途**: 查询第三方库文档和最佳实践

**使用场景**:
- 查询 Go 标准库用法
- 查询 MongoDB Go Driver 文档
- 了解框架的最佳实践

**示例**:
```python
# 查询 MongoDB projection 用法
context7("MongoDB Go Driver projection query optimization")
```

---

## sequential-thinking - 决策分析

**用途**: 复杂决策的逐步分析和推理

**使用场景**:
- 设计渐进式迁移路径
- 分析多种方案的优劣
- 评估风险和回滚策略

**示例**:
```python
mcp__sequential-thinking__sequentialthinking(
    thought="分析迁移方案: 一次性迁移 vs 渐进式迁移",
    thoughtNumber=1,
    totalThoughts=5,
    nextThoughtNeeded=true
)
```

---

## Bash - 命令执行

**用途**: 执行系统命令

**使用场景**:
- 创建目录: `mkdir -p openspec/changes/<change-id>`
- 验证 OpenSpec: `openspec validate <change-id> --strict`
- 提交代码: `git commit`

**注意**: 编译验证和测试请使用 `/code-review` skill，它包含完整的语法检查流程。

---

## 完整工作流示例

### 阶段 1: 准备新能力

```bash
# 1. 分析现有实现
get_symbols_overview(relative_path="pkg/dao/relay.go", depth=1)

# 2. 查看旧函数实现
find_symbol(
    name_path_pattern="GetRelayInfo",
    relative_path="pkg/dao/relay.go",
    include_body=true
)

# 3. 实现新函数 (使用 insert_after_symbol)
insert_after_symbol(
    name_path="GetRelayInfo",
    relative_path="pkg/dao/relay.go",
    body="[新函数实现]"
)

# 4. 使用 /code-review skill 进行验证
```

### 阶段 2: 双轨运行

```bash
# 1. 查找所有使用旧函数的位置
search_for_pattern(
    substring_pattern="GetRelayInfo",
    relative_path="pkg/"
)

# 2. 使用 /code-review skill 验证新旧函数可以共存
```

### 阶段 3: 逐个迁移

**迁移单元 1**:
```bash
# 1. 查找符号详情
find_symbol(
    name_path_pattern="NodeInstallWorkflow/Execute",
    include_body=true
)

# 2. 查找引用,评估影响
find_referencing_symbols(
    name_path="NodeInstallWorkflow/Execute",
    relative_path="pkg/workflow/node_install.go"
)

# 3. 替换符号体,切换到新函数
replace_symbol_body(
    name_path="NodeInstallWorkflow/Execute",
    body="[使用新函数的实现]"
)

# 4. 使用 /code-review skill 进行验证

# 5. 提交
Bash("git add . && git commit -m 'refactor: migrate NodeInstallWorkflow to new relay query'")
```

**迁移单元 2-N**: 重复上述流程

---

## 最佳实践

### 1. 渐进式使用工具

- ✅ 先用 `get_symbols_overview` 了解整体结构
- ✅ 再用 `search_for_pattern` 评估影响范围
- ✅ 最后用 `find_symbol` 获取详细实现

### 2. 验证优先

- ✅ 每次修改后立即使用 `/code-review` skill 验证
- ✅ 使用 `find_referencing_symbols` 确保所有引用都更新

### 3. 限制搜索范围

- ✅ 使用 `relative_path` 参数限制搜索范围
- ✅ 避免全局搜索,提升性能

### 4. 保持原子性

- ✅ 每个迁移单元使用独立的工具调用序列
- ✅ 失败时可以精确回滚到上一步

---

## 常见问题

### Q1: 如何评估迁移影响范围?

**答案**: 使用组合流程
```bash
# 1. 搜索模式,找到所有使用位置
search_for_pattern(substring_pattern="OldFunction")

# 2. 查找符号引用,确认调用关系
find_referencing_symbols(name_path="OldFunction")

# 3. 统计影响范围 (文件数、调用次数)
```

### Q2: 如何确保迁移完整性?

**答案**: 使用验证流程
```bash
# 1. 搜索旧模式,应该没有结果
search_for_pattern(substring_pattern="OldFunction")

# 2. 搜索新模式,应该有预期的结果
search_for_pattern(substring_pattern="NewFunction")

# 3. 使用 /code-review skill 确保功能正常
```

### Q3: 工具调用失败怎么办?

**答案**: 检查参数和路径
- 确认 `relative_path` 相对于项目根目录
- 确认 `name_path` 符号路径正确 (使用 `/` 分隔)
- 查看错误消息,调整参数后重试

---

## 参考资源

- **serena 工具完整文档**: 查看 MCP 服务器的工具定义
- **实际案例**: `openspec/changes/remove-relayinfo-cache/` 展示了完整的工具使用流程
