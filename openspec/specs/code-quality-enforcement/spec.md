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

### Requirement: 修复不得影响外部行为
所有 lint 修复必须保持代码的外部行为不变，仅改变内部实现。

#### Scenario: API 接口保持不变
- **WHEN** 完成所有 lint 修复
- **THEN** 所有 public API 的签名和行为保持不变

#### Scenario: 测试用例全部通过
- **WHEN** 完成所有 lint 修复后运行测试
- **THEN** 所有现有测试用例必须通过
