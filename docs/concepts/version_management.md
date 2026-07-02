## version_management

版本命名用于描述 bk-nodemgr 发布包的版本阶段、迭代序号和可选标记。

### 格式

```text
v{major}.{minor}.{patch}-{phase}.{serial}[-{mark}]
```

### 正则

```text
^v(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)-(?P<phase>[a-zA-Z]+)\.(?P<serial>\d+)(?:-(?P<mark>[a-zA-Z0-9][a-zA-Z0-9.-]*))?$
```

### 字段

| 字段 | 说明 | 示例 |
| --- | --- | --- |
| `major` | 主版本号 | `3` |
| `minor` | 次版本号 | `0` |
| `patch` | 修订版本号 | `1` |
| `phase` | 发布阶段 | `alpha`、`beta`、`release` |
| `serial` | 当前阶段内的迭代序号 | `1` |
| `mark` | 可选标记，用于补充测试、渠道或临时标识 | `test` |

### 迭代规则

- `major`、`minor`、`patch` 共同描述基础版本。
- `phase` 表示当前发布阶段，不同阶段的语义由发布流程约定。
- `serial` 在同一基础版本和阶段内递增。
- `mark` 为可选字段，仅用于补充说明，不参与基础版本定义。

### Tag 规则

Git Tag 应直接使用完整版本名，保持与发布包版本一致。

### 示例

| 版本 | major | minor | patch | phase | serial | mark |
| --- | --- | --- | --- | --- | --- | --- |
| `v3.0.1-beta.1` | `3` | `0` | `1` | `beta` | `1` | - |
| `v3.0.1-beta.1-test` | `3` | `0` | `1` | `beta` | `1` | `test` |

### 说明范围

本文仅描述版本命名方法，不定义版本比较语义、发布自动化流程或兼容性保证。
