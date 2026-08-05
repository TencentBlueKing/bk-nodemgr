# 包版本状态 FAQ

## 1. `enabled` 和 `is_visible` 分别表示什么？

包版本包含两个相互独立的布尔字段：

| 字段         | 作用范围 | 含义 |
|------------|------|----|
| `enabled`  | backend 业务流程 | 控制包是否可被安装、重装、升级等流程使用。 |
| `is_visible` | application 和前端 | 控制包是否在前端作为可见版本展示。 |

backend 执行涉及包版本的业务流程时只判断 `enabled`。`is_visible` 不参与包是否可用的判断，只服务于前端展示。

## 2. 两个字段的优先级是什么？

`enabled` 的优先级高于 `is_visible`。

当 `enabled=false` 时，无论 `is_visible` 是什么值，该包都不能参与安装、重装、升级等 backend 业务流程。
`is_visible` 只在展示层表达是否显示，不会把一个已停用的包重新变成可用状态。

| `enabled` | `is_visible` | backend 是否可用 | 前端展示状态 | 包管理主要操作 |
|-----------|--------------|----------------|------------|--------------|
| `false`   | `false`      | 否 | 隐藏 | 启用并显示 |
| `false`   | `true`       | 否 | 显示，但仍不可用 | 启用并显示 |
| `true`    | `false`      | 是 | 隐藏 | 停用、显示 |
| `true`    | `true`       | 是 | 显示 | 停用并隐藏、隐藏 |

前端组合按钮首先判断 `enabled`，只有 `enabled=true` 时才继续根据 `is_visible` 决定显示“显示”还是“隐藏”。

## 3. 修改一个字段会自动修改另一个字段吗？

不会。两个字段的单一接口保持独立语义：

- `enable` 只把 `enabled` 设置为 `true`
- `disable` 只把 `enabled` 设置为 `false`
- `visible` 只把 `is_visible` 设置为 `true`
- `unvisible` 只把 `is_visible` 设置为 `false`

因此，隐藏包不会导致包停用，停用包也不会自动隐藏。

application 额外提供两个组合接口，用于包管理页面的常用操作：

- `enable_and_visible`：依次执行 `enable` 和 `visible`
- `disable_and_unvisible`：依次执行 `disable` 和 `unvisible`

组合能力是新增接口，不改变原有 `enable`、`disable`、`visible`、`unvisible` 的单一职责。

## 4. 包管理页面如何组合操作按钮？

Agent、Proxy 和插件包使用相同的判断规则：

1. `enabled=false`：提供“启用并显示”。
2. `enabled=true` 且 `is_visible=true`：提供“停用并隐藏”和“隐藏”。
3. `enabled=true` 且 `is_visible=false`：提供“停用”和“显示”。
4. 默认版本操作以 `enabled` 为前置条件，不由 `is_visible` 决定。

这个顺序保证包的业务可用性始终由 `enabled` 决定，同时允许前端独立控制展示状态。

