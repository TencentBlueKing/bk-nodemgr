# Code Review: Worktree 工具使用策略

## 核心原则

**Serena 完全支持 `.worktrees/` 目录**，通过 `relative_path` 参数即可访问。
审查 PR 时，**所有搜索和符号操作必须优先使用 serena**，不得降级到 grep/rg。

## 已验证的 Serena 工具 worktree 能力

| 工具                     | worktree 支持 | 用法                                       |
|------------------------|-------------|------------------------------------------|
| `search_for_pattern`   | ✅           | `relative_path=".worktrees/pr-XXX/path"` |
| `get_symbols_overview` | ✅           | 同上                                       |
| `find_symbol`          | ✅           | 同上（已删除的符号正确返回空）                          |
| `list_dir`             | ✅           | 可列出 `.worktrees/` 子目录                    |

## PR 审查中的正确工具选择

| 审查任务         | 正确工具                                        | 示例                                                                                      |
|--------------|---------------------------------------------|-----------------------------------------------------------------------------------------|
| 查找被删除字段的剩余引用 | `serena.search_for_pattern`                 | `substring_pattern="FieldName", relative_path=".worktrees/pr-XXX"`                      |
| 查找函数所有引用     | `serena.search_for_pattern` 或 `find_symbol` | 在 worktree 中搜索                                                                          |
| 查找特定代码模式     | `serena.search_for_pattern`                 | `substring_pattern="defer metric\\.End"`                                                |
| 统计外部 err 声明  | `serena.search_for_pattern`                 | `substring_pattern="var err error", relative_path=".worktrees/pr-XXX/specific/file.go"` |
| 获取文件符号概览     | `serena.get_symbols_overview`               | 不需要读全文即可了解结构                                                                            |
| 对比新旧代码差异     | `git diff` (Shell)                          | serena 不提供 diff 功能，此场景仍需 Shell                                                          |

## 仅在以下场景使用 Shell/grep

1. **git 操作**：`git diff`、`git log`、`git fetch` 等
2. **编译检查**：`go build`、`go vet`
3. **lint 检查**：`make lint`
4. serena 工具调用**返回错误**时的降级方案

## 反模式

- ❌ 假设 serena 不能访问 worktree 而直接用 grep
- ❌ 在 worktree 中用 `grep -rn` 搜索代码模式
- ❌ 用 `grep` 统计代码中的特定 pattern 出现次数
- ❌ 不先尝试 serena 就降级到传统工具