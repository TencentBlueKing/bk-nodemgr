## MODIFIED Requirements

### Requirement: WrapFn 闭包必须使用局部 err 变量（统一风格）
所有 `internal/backend/storage/` 下**全部子包**的 `basestorage.WrapFn` 闭包参数中，当闭包需要处理 error 时，MUST 在闭包内部声明 `var err error` 遮蔽外部同名变量。MUST NOT 在闭包内直接赋值外部 `err` 变量。此规则不再仅限于 `topo` 和 `workflow`，而是覆盖所有 storage 子包。**包括单调用闭包也必须使用 `var err error; err = fn(...); return err` 模式，不允许简化为 `return fn(...)`。**

#### Scenario: cipher 闭包使用局部 err
- **WHEN** `cipher/` 下 3 个 WrapFn 闭包需要处理 error
- **THEN** 每个闭包内必须声明 `var err error`，禁止直接赋值外部 `err`

#### Scenario: deploypolicy 闭包使用局部 err
- **WHEN** `deploypolicy/` 下 8 个 WrapFn 闭包需要处理 error
- **THEN** 每个闭包内必须声明 `var err error`，禁止直接赋值外部 `err`

#### Scenario: plugin 闭包使用局部 err
- **WHEN** `plugin/` 下 42 个 WrapFn 闭包需要处理 error
- **THEN** 每个闭包内必须声明 `var err error`，禁止直接赋值外部 `err`

#### Scenario: workflow 残留闭包使用局部 err
- **WHEN** `workflow/` 中残留的 2 处 WrapFn 闭包使用外层 `err` 变量
- **THEN** 必须在闭包内添加 `var err error` 声明

#### Scenario: topo 残留闭包风格统一
- **WHEN** `topo/host.go` 和 `topo/topo.go` 中存在 `if err :=` 风格的闭包
- **THEN** 统一为 `var err error` 风格以保持全项目一致性

### Requirement: 仅返回 error 的函数必须直接返回 WrapFn（闭包内统一 var err error）
当外层函数签名仅返回 `error` 且闭包不需要捕获额外返回值时，MUST 使用 `return s.WrapFn(...)` 形式。**闭包内部仍须使用 `var err error; err = fn(...); return err` 模式，不允许简化为 `return fn(...)`。**此规则从 `topo`/`workflow` 扩展到所有 storage 子包。

#### Scenario: cipher 消除冗余返回
- **WHEN** `cipher/` 中 `CreateCipher` 等 error-only 方法使用冗余返回模式
- **THEN** 改为 `return s.WrapFn(...)` 直接返回

#### Scenario: deploypolicy 消除冗余返回
- **WHEN** `deploypolicy/` 中 8 个 error-only 方法使用冗余返回模式
- **THEN** 改为 `return s.WrapFn(...)` 直接返回

#### Scenario: plugin 消除冗余返回
- **WHEN** `plugin/` 中 42 个 error-only 方法使用冗余返回模式
- **THEN** 改为 `return s.WrapFn(...)` 直接返回

#### Scenario: release 消除冗余返回
- **WHEN** `release/` 中 27 个 error-only 方法使用冗余返回模式
- **THEN** 改为 `return s.WrapFn(...)` 直接返回

#### Scenario: tenant 消除冗余返回
- **WHEN** `tenant/` 中 5 个 error-only 方法使用冗余返回模式
- **THEN** 改为 `return s.WrapFn(...)` 直接返回

#### Scenario: configpolicy 消除冗余返回
- **WHEN** `configpolicy/` 中 8 个 error-only 方法使用冗余返回模式
- **THEN** 改为 `return s.WrapFn(...)` 直接返回

#### Scenario: 多返回值函数保留现有模式
- **WHEN** 函数返回 `(T, error)` 等多值
- **THEN** 保留 `var err; err = WrapFn; return result, err` 模式，不视为冗余

### Requirement: 错误包装必须使用 %w 而非 %v
所有 `internal/backend/storage/` 下**全部子包**的 `fmt.Errorf` 中包装 error 类型的占位符 MUST 使用 `%w`。此规则从 `topo`/`workflow` 扩展到所有 storage 子包。

#### Scenario: plugin 中 %v 替换为 %w
- **WHEN** `plugin/dao_plugin_workflow.go` 中 7 处和 `plugin/dao_plugin_deployment.go` 中 4 处使用 `%v` 包装 error
- **THEN** 全部替换为 `%w`

#### Scenario: workflow 残留 %v 替换为 %w
- **WHEN** `workflow/domain_stop_oper_inst.go` 和 `workflow/dao_operation.go` 中各 1 处使用 `%v` 包装 error
- **THEN** 替换为 `%w`

#### Scenario: node 中 %v 替换为 %w
- **WHEN** `node/dao_node_deloyment.go` 中 2 处使用 `%v` 包装 error
- **THEN** 替换为 `%w`

#### Scenario: 非错误类型的 %v 不受影响
- **WHEN** `fmt.Errorf` 中 `%v` 用于格式化非 error 类型变量（如 string、int）
- **THEN** 保持 `%v` 不变，仅 error 类型参数需使用 `%w`
