# Guard Clause 模式：减少嵌套循环复杂度

**标签：** `single-level` `entropy`
**严重程度：** 💡 建议
**发现日期：** 2026-04-07
**涉及文件：** `internal/backend/storage/topo/topo.go`

## 问题描述

在嵌套循环中使用临时标志变量（如 `hasMatch`）来控制外层逻辑，导致代码嵌套层级过深（3层），主干逻辑被包裹在条件判断中，可读性下降。

## 问题代码

```go
// ❌ 深嵌套 + 临时标志变量
func GetNetworkUnitIDsByAccessPoints(accessPointIDs []int64) ([]int64, error) {
    units, err := storage.ListNetworkUnit(ctx, nil, nil)
    if err != nil {
        return nil, err
    }
    
    requestedSet := make(map[int64]struct{})
    for _, id := range accessPointIDs {
        requestedSet[id] = struct{}{}
    }
    
    unitIDSet := make(map[int64]struct{})
    for _, unit := range units {
        hasMatch := false
        for _, apID := range unit.AccessPoints {
            if _, found := requestedSet[apID]; found {
                hasMatch = true
                break
            }
        }
        if hasMatch {
            unitIDSet[unit.ID] = struct{}{}
        }
    }
    
    return mapKeysToSlice(unitIDSet), nil
}
```

**问题点：**
- 嵌套层级：3 层（外层 for + 内层 for + if found）
- 临时变量 `hasMatch` 增加认知负担
- 主干逻辑（`unitIDSet[unit.ID] = struct{}{}`）被嵌套包裹

## 修复后

```go
// ✅ Guard Clause + 早退出
func GetNetworkUnitIDsByAccessPoints(accessPointIDs []int64) ([]int64, error) {
    units, err := storage.ListNetworkUnit(ctx, nil, nil)
    if err != nil {
        return nil, err
    }
    
    requestedSet := make(map[int64]struct{})
    for _, id := range accessPointIDs {
        requestedSet[id] = struct{}{}
    }
    
    unitIDSet := make(map[int64]struct{})
    for _, unit := range units {
        for _, apID := range unit.AccessPoints {
            if _, found := requestedSet[apID]; !found {
                continue  // Guard: 跳过不匹配的项
            }
            unitIDSet[unit.ID] = struct{}{}  // 主干逻辑
            break  // 找到匹配即退出
        }
    }
    
    return mapKeysToSlice(unitIDSet), nil
}
```

**改进效果：**
- 嵌套层级：3 层 → 2 层
- 消除临时变量 `hasMatch`
- 主干逻辑清晰可见
- 快速失败路径（fast path）优先

## 违反的原则

**2.1 单层级函数（single-level）**：函数内部应保持单一抽象层级，避免深层嵌套。临时标志变量和多层条件判断增加了理解成本。

**5.2 控制熵（entropy）**：代码复杂度应该被主动控制。Guard Clause 通过提前退出减少分支路径，降低圈复杂度。

参考 `references/architecture-principles.md` 第 2.1 节和第 5.2 节。

## 通用模式

### 模式 1：循环中的条件跳过

```go
// ❌ 避免
for _, item := range items {
    if condition(item) {
        process(item)
    }
}

// ✅ 推荐
for _, item := range items {
    if !condition(item) {
        continue
    }
    process(item)
}
```

### 模式 2：嵌套循环中的匹配查找

```go
// ❌ 避免
for _, item := range items {
    found := false
    for _, target := range targets {
        if item == target {
            found = true
            break
        }
    }
    if found {
        process(item)
    }
}

// ✅ 推荐
for _, item := range items {
    for _, target := range targets {
        if item != target {
            continue
        }
        process(item)
        break
    }
}
```

### 模式 3：函数入口的参数校验

```go
// ❌ 避免
func Process(data *Data) error {
    if data != nil {
        if data.IsValid() {
            return doProcess(data)
        } else {
            return errors.New("invalid data")
        }
    } else {
        return errors.New("nil data")
    }
}

// ✅ 推荐
func Process(data *Data) error {
    if data == nil {
        return errors.New("nil data")
    }
    if !data.IsValid() {
        return errors.New("invalid data")
    }
    return doProcess(data)
}
```

## 项目中的候选位置

通过代码扫描，发现以下文件中存在可应用 Guard Clause 的模式：

**高优先级（21+ 处）：**
1. `internal/backend/router/api-v3/node/agent/install_check.go` (11 处)
2. `internal/backend/router/api-v3/node/proxy/install_check.go` (10 处)

**已应用 Guard Clause 的正面示例：**
- `internal/backend/storage/plugin/dao_plugin.go` (L76-79)
- `internal/backend/storage/topo/host.go` (L301-308, L311-324)

## 不适用场景

以下情况**不应**使用 Guard Clause：

1. **需要处理所有匹配项**（不能提前 break）
2. **条件逻辑本身很复杂**（取反后更难读）
3. **主干逻辑非常短**（已经足够简洁）

## 审查点评

嵌套循环中使用临时标志变量是常见的代码异味。Guard Clause 模式通过提前 `continue` 或 `break`，将"不符合条件"的快速失败路径前置，让主干逻辑保持在较浅的嵌套层级，显著提升可读性。建议在代码审查时，对 3 层以上嵌套的循环结构进行重点检查。

## 参考资料

- [Refactoring: Replace Nested Conditional with Guard Clauses](https://refactoring.com/catalog/replaceNestedConditionalWithGuardClauses.html)
- [Clean Code: Chapter 3 - Functions](https://www.oreilly.com/library/view/clean-code-a/9780136083238/)
