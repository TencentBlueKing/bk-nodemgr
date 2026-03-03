## ADDED Requirements

### Requirement: 代码必须符合空行规范
代码文件中不得包含不必要的前导空行。所有 storage 层文件必须通过 golangci-lint 的 whitespace 检查。

#### Scenario: 检测到不必要的空行
- **WHEN** 运行 `make lint` 命令
- **THEN** 不应出现 "unnecessary leading newline" 错误

#### Scenario: 修复后通过检查
- **WHEN** 移除所有不必要的前导空行后运行 lint
- **THEN** whitespace linter 不报告任何错误

### Requirement: 函数认知复杂度必须不超过 20
所有函数的认知复杂度必须控制在 20 以内，以保证代码可维护性和可读性。

#### Scenario: 高复杂度函数被拆分
- **WHEN** 函数认知复杂度超过 20
- **THEN** 必须将函数拆分为多个子函数或简化逻辑

#### Scenario: 拆分后通过检查
- **WHEN** 重构后运行 `make lint`
- **THEN** gocognit linter 不报告任何复杂度超标错误

#### Scenario: GetProxyUpstreamAccessEndpoints 复杂度降低
- **WHEN** 重构 `GetProxyUpstreamAccessEndpoints` 函数
- **THEN** 认知复杂度从 21 降低到 20 或以下

#### Scenario: DistinctHost 复杂度降低
- **WHEN** 重构 `DistinctHost` 函数
- **THEN** 认知复杂度从 21 降低到 20 或以下

#### Scenario: UpdateNetworkUnit 复杂度降低
- **WHEN** 重构 `UpdateNetworkUnit` 函数
- **THEN** 认知复杂度从 31 降低到 20 或以下

### Requirement: 函数返回值必须被使用
函数声明的返回值必须在实际使用中被调用方处理，不得存在始终返回固定值的返回值。

#### Scenario: 移除未使用的 error 返回值
- **WHEN** `convertTopoEventConditionsToOptions` 函数的 error 返回值始终为 nil
- **THEN** 必须移除该 error 返回值并更新所有调用方

#### Scenario: 修复后通过检查
- **WHEN** 移除未使用的返回值后运行 lint
- **THEN** unparam linter 不报告任何错误

### Requirement: WrapFn 闭包必须使用局部 err 变量
所有 `basestorage.WrapFn` 的闭包参数中，当闭包需要处理 error 时，必须在闭包内部声明 `var err error` 遮蔽外部同名变量。禁止在闭包内直接赋值外部 `err` 变量。

#### Scenario: 闭包捕获多返回值时使用局部 err
- **WHEN** WrapFn 闭包内调用返回 `(result, error)` 的函数
- **THEN** 闭包内必须声明 `var err error`，通过 `result, err = fn()` 赋值，最终 `return err`

#### Scenario: 仅返回 error 的函数直接返回 WrapFn
- **WHEN** 外层函数签名仅返回 `error` 且闭包不需要捕获额外返回值
- **THEN** 必须使用 `return s.WrapFn(...)` 形式，禁止使用 `var err error; err = s.WrapFn(...); if err != nil { return err }; return nil` 冗余模式

#### Scenario: 检测到违规的外部 err 引用
- **WHEN** 闭包内通过赋值（`err =` 或 `, err =`）修改外部函数声明的 `err` 变量
- **THEN** 必须在闭包开头添加 `var err error` 声明

### Requirement: 错误包装必须使用 %w 而非 %v
所有 `fmt.Errorf` 中包装 error 类型的占位符必须使用 `%w`，禁止使用 `%v` 或 `%s`。

#### Scenario: storage 层 fmt.Errorf 使用 %w
- **WHEN** `internal/backend/storage/` 下任意文件使用 `fmt.Errorf` 包装 error
- **THEN** 必须使用 `%w` 占位符以保持错误链完整

#### Scenario: 错误链可通过 errors.Is 追溯
- **WHEN** 调用者对 storage 层返回的 error 使用 `errors.Is()` 或 `errors.As()`
- **THEN** 能够正确匹配底层错误类型

### Requirement: WrapFn metric defer 必须正确捕获 error
`basestorage.WrapFn` 中的 `defer metric.End(err)` 必须使用闭包形式以正确捕获 named return value。

#### Scenario: metric 记录真实错误状态
- **WHEN** WrapFn 包装的函数返回 error
- **THEN** `metric.End()` 接收到非 nil 的 error 参数

#### Scenario: metric 记录成功状态
- **WHEN** WrapFn 包装的函数返回 nil
- **THEN** `metric.End()` 接收到 nil 的 error 参数

#### Scenario: defer 使用闭包形式
- **WHEN** 检查 `pkg/basestorage/base.go` 中 `WrapFn` 的 metric defer
- **THEN** 代码形式为 `defer func() { metric.End(err) }()` 而非 `defer metric.End(err)`

### Requirement: 修复不得影响外部行为
所有 lint 修复必须保持代码的外部行为不变，仅改变内部实现。

#### Scenario: API 接口保持不变
- **WHEN** 完成所有 lint 修复
- **THEN** 所有 public API 的签名和行为保持不变

#### Scenario: 测试用例全部通过
- **WHEN** 完成所有 lint 修复后运行测试
- **THEN** 所有现有测试用例必须通过

#### Scenario: Metric 行为变更已知且受控
- **WHEN** 修复 WrapFn metric defer 后
- **THEN** Prometheus `request_total` 指标中 error 标签将反映真实错误状态，此行为变更是预期的且已记录
