# Review 案例库

本目录逐步积累代码审查中发现的典型案例，供 reviewer 参考和学习。

## 目的

- **沉淀经验**：将日常 CR 中的发现系统化保存，避免反复口头传授
- **辅助审查**：review 时引用具体案例说明问题，比抽象原则更有说服力
- **新人培训**：新 reviewer 通过阅读案例快速理解项目的代码质量标准

## 案例格式

每个案例一个 Markdown 文件，文件名格式：`{序号}-{简短描述}.md`

例如：`001-duplicate-proto-converter.md`

使用 `_TEMPLATE.md` 作为模板创建新案例。

## 标签体系

每个案例标注 **违反的原则**（对应 `references/architecture-principles.md` 的章节编号）：

| 标签 | 原则 |
|------|------|
| `DRY` | 1.1 不要重复 |
| `orthogonality` | 1.2 正交性 |
| `DIP` | 1.3 依赖方向 |
| `composition` | 1.4 组合优于继承 |
| `single-level` | 2.1 单层级函数 |
| `tell-dont-ask` | 2.2 只管命令不要询问 |
| `minimal-params` | 2.3 传入最小参数集 |
| `contract` | 2.4 契约式设计 |
| `crash-early` | 3.1 尽早崩溃 |
| `error-wrapping` | 3.2 错误传递 |
| `silence` | 3.3 缄默原则 |
| `explicit` | 4.1 显式处理 |
| `terminology` | 4.2 统一术语 |
| `no-hardcode` | 4.3 不写死可变值 |
| `reversibility` | 5.1 可逆性 |
| `entropy` | 5.2 控制熵 |
| `AGENTS` | 违反 AGENTS.md 强约束 |

## 案例索引

> 随着案例积累，在此处维护索引表。

| 编号 | 标题 | 标签 | 严重程度 |
|------|------|------|----------|
| 001 | API 路由放置在错误的领域 | `orthogonality` `AGENTS` | ⚠️ 重要 |

## 如何添加新案例

1. 复制 `_TEMPLATE.md` 为 `{下一个序号}-{描述}.md`
2. 填写案例内容（问题代码、修复代码、违反原则、点评）
3. 在上方索引表中添加一行
4. 提交 PR

## 审查流程中如何使用

在代码审查步骤 5（一致性检查）和步骤 6（质量检查）中：
1. 如果发现的问题与某个已有案例类似，在报告中引用案例编号
2. 格式：`参见案例 cases/001-duplicate-proto-converter.md`
3. 如果发现了新的典型问题，审查完成后考虑补充为新案例
