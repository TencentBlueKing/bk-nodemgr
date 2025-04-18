# Protocol Buffers 编写规范

## 基本原则

1. **命名规范**
    - Message命名：使用PascalCase（首字母大写），如`NodeWorkflowInfo`
    - Field命名：使用snake_case（小写+下划线），如`workflow_id`
    - Service命名：使用PascalCase，如`NodeWorkflow`
    - RPC方法命名：使用PascalCase，如`NodeWorkflowList`

2. **字段命名一致性**
    - 对于集合类型的字段，统一使用原始形式，如`workflow`、`bk_biz_id`等, 具体的单复数根据结构体的语义来决定
    - 不必过度纠结单复数形式（如是用s还是es），但在同一项目中保持一致性

## 类型使用

- 整数类型使用 `int64`
- 无符号整数类型使用 `uint64`
