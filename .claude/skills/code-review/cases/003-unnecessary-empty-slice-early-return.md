# 无意义的空切片 Early Return

**标签：** `entropy` `simplicity`
**严重程度：** 💡 建议
**发现日期：** 2026-06-29

## 问题描述

在流水线式的函数中，对每个中间步骤的空切片结果都做 early return，是过度防御。这些检查没有实际作用，却增加了函数出口数量，每个出口还需维护相同的清理逻辑。

## 问题代码

```go
// ❌ 凭空制造函数出口
candidates := filterCandidates(...)
if len(candidates) == 0 {
    logStats(stats)
    return nil
}

validated := validateCandidates(candidates, ...)
if len(validated) == 0 {
    logStats(stats)
    return nil
}

rechecked := recheckCandidates(validated, ...)
if len(rechecked) == 0 {
    logStats(stats)
    return nil
}

updateHosts(rechecked)
logStats(stats)
```

## 修复后

```go
// ✅ 数据自然流动，只在有副作用处判断
candidates := filterCandidates(...)
validated := validateCandidates(candidates, ...)
rechecked := recheckCandidates(validated, ...)

logStats(stats)
if len(rechecked) > 0 {
    updateHosts(rechecked)
}
```

## 违反的原则

- **控制熵**：不必要的分支增加代码复杂度，每个出口都是维护负担
- **简单性**：空切片是 Go 的零值，`for range` 对空切片天然安全，不需要特殊处理

## 审查点评

空切片检查的 early return 是过度防御的表现。让数据自然流动，只在真正有副作用（写数据库、发送请求）的地方做条件判断。
