# API 文档编写详细流程

完整的 API 文档编写 8 阶段流程。

## 阶段 1: 理解需求

**目标**: 明确文档编写的目的和范围

**步骤**:

1. 确认需要编写的 API 范围
2. 阅读模板规范 (`docs/developer/api_doc_template.md`)
3. 查看示例参考 (`docs/developer/api_doc_example*.md`)
4. 确定使用哪种模板变体

## 阶段 2: 收集技术资料

**目标**: 深入理解 API 的技术实现和业务逻辑

**步骤**:

1. **阅读 Proto 定义** (必读)
    - 位置: `proto/backend/api/v3/*.proto`
    - 关注点: Service/RPC 方法、Request/Response 消息、字段类型和注释、URL 路径和 HTTP 方法

2. **阅读类型定义** (必读)
    - 位置: `pkg/types/*.go`
    - 关注点: 业务模型结构、枚举类型和常量、验证逻辑和约束规则

3. **阅读概念文档** (推荐)
    - 位置: `docs/concepts/`
    - 关注点: 业务概念、术语、使用场景

4. **查看实现代码** (可选)
    - 位置: `internal/backend/router/api-v3/`
    - 关注点: 参数校验逻辑、业务处理流程

## 阶段 3: 设计文档结构

**目标**: 规划文档的整体结构

**步骤**:

1. 选择合适的模板（基础/复杂查询/文件操作）
2. 规划文档章节
3. 设计参数说明方式（特别是嵌套和多态字段）

## 阶段 4: 编写 API 文档

**目标**: 按照模板编写完整准确的文档

**步骤**:

1. **编写描述部分**
   ```markdown
   ### 描述

   - 该接口提供版本: v3.0.0+
   - 该接口所需权限: action_id（中文名）、action_id2（中文名2）。
   - 该接口功能描述: [一句话说明]
   ```
   
   **权限格式说明**:
   - 格式: `action_id（中文名）`
   - 多个权限用顿号分隔: `action1（名称1）、action2（名称2）`
   - 示例: `agent_operate（操作Agent）、networkunit_use_for_agent（使用网络单元部署Agent）`
   - 无权限要求时写: `无`
   - 查找方法:
     1. 在 handler 代码中搜索 `h.authorizer.Check` 调用
     2. 找到 `auth.ActionXxx` 常量
     3. 在 `internal/backend/auth/action.go` 中查找对应的 action_id 和中文名
     4. ActionDisplayName 函数返回中文名

2. **编写 URL**
   ```markdown
   ### URL

   POST /api/v3/[service]/[resource]/[method]
   ```

3. **编写输入参数表格**
    - 顶层参数表格
    - 嵌套对象的子表格（使用 #### 标题）
    - 多态参数详细说明每种类型的结构

4. **编写调用示例**
    - 提供真实可用的 JSON 示例
    - 对于复杂 API，提供多个场景的示例

5. **编写响应示例**
    - 提供成功响应的完整 JSON
    - 包含所有重要字段

6. **编写响应参数说明**
    - 顶层响应字段表格
    - data 对象的字段表格
    - 嵌套对象的详细说明

## 阶段 5: 补充实用信息（可选）

**目标**: 提供帮助用户使用的补充信息

**步骤**:

1. 添加概述章节（功能简介、核心概念）
2. 添加典型使用场景示例
3. 添加常见问题解答
4. 添加相关文档链接

## 阶段 6: 交叉验证（关键）

**目标**: 确保字段语义准确，避免仅凭 Proto 定义产生误解

**核心原则**: Proto 只是起点，不是终点。Proto 定义通常只有字段名和基本类型，缺乏业务语义。

### 验证流程

```
Proto 定义 → Go 类型定义 → 使用场景分析 → 验证文档描述
    ↓              ↓              ↓              ↓
 字段名/类型    枚举值/约束    实际业务含义    确保准确性
```

### 必须验证的字段类型

| 字段类型                     | 风险              | 验证方法                         |
|--------------------------|-----------------|------------------------------|
| `repeated string`        | 可能是预定义枚举而非自定义值  | 搜索对应的 Go 类型定义                |
| `string` (带注释 "枚举")      | 枚举值可能不完整        | 查找 `const` 定义和 `Validate` 方法 |
| 嵌套 `object`              | 字段间可能有隐藏的业务关联   | 查找 struct 定义和字段注释            |
| `google.protobuf.Struct` | 多态结构，Proto 无法体现 | 必须找到实际的 Go 类型定义              |

### 验证步骤

**步骤 1: 识别高风险字段**

阅读 Proto 后，标记以下字段为"待验证":

- 看起来像枚举但 Proto 中没有明确定义的字段
- `repeated string` 类型的字段（可能是标签/标记）
- 嵌套的 object 字段
- 任何使用 `google.protobuf.Struct` 的字段

**步骤 2: 使用 Serena 深入代码**

```python
# 搜索字段在代码中的所有使用
serena.search_for_pattern("字段名|FieldName", restrict_search_to_code_files=True)

# 查找类型定义
serena.find_symbol("TypeName", relative_path="pkg/types", include_body=True, depth=1)

# 查看验证逻辑
serena.search_for_pattern("Validate.*FieldName|FieldName.*Validate", relative_path="pkg/types")
```

**步骤 3: 验证文档描述**

对比代码中发现的真实语义与文档描述:

| 验证项  | 检查内容                        |
|------|-----------------------------|
| 枚举值  | 文档列出的枚举值是否与代码 `const` 定义一致？ |
| 字段含义 | 文档描述是否与代码注释/业务逻辑一致？         |
| 示例值  | JSON 示例中的值是否是有效的枚举值？        |
| 约束关系 | 是否遗漏了字段间的依赖或互斥关系？           |

**步骤 4: 修正文档**

如发现不一致，立即修正:

- 更新字段描述
- 补充/修正枚举值
- 修正 JSON 示例中的错误值
- 添加遗漏的约束说明

### 典型案例

#### 案例 1: proxy_tags 字段

**Proto 定义** (误导性):

```protobuf
repeated string proxy_tags = 19;  // 看起来像自定义标签
```

**初始文档** (错误):

```markdown
| proxy_tags | array | 否 | Proxy 标签列表，用于对 Proxy 进行分组和筛选 |

示例: `["region-beijing", "env-prod"]`
```

**Serena 验证**:

```python
serena.search_for_pattern("proxy_tags|ProxyTag", restrict_search_to_code_files=True)
```

**发现的 Go 类型定义**:

```go
// pkg/types/host.go
type ProxyTag string

const (
ProxyTagDedicatedInstaller = "dedicated_installer"
ProxyTagClusterTunnel      = "cluster_tunnel"
ProxyTagFileTunnel = "file_tunnel"
ProxyTagDataTunnel = "data_tunnel"
)

func (tag ProxyTag) Validate() error {
switch tag {
case ProxyTagDedicatedInstaller, ProxyTagClusterTunnel, ProxyTagFileTunnel, ProxyTagDataTunnel:
return nil
default:
return errors.New("invalid proxy tag")
}
}
```

**修正后的文档**:

```markdown
| proxy_tags | array | 否 | Proxy 功能标签列表（枚举值：dedicated_installer、cluster_tunnel、file_tunnel、data_tunnel） |

**proxy_tags 枚举说明**:

- `dedicated_installer`: 专用安装器，该 Proxy 可作为跨网络单元安装的中转节点
- `cluster_tunnel`: 集群通道，该 Proxy 提供集群管控通信服务
- `file_tunnel`: 文件通道，该 Proxy 提供文件传输服务
- `data_tunnel`: 数据通道，该 Proxy 提供数据上报服务

示例: `["dedicated_installer", "cluster_tunnel", "file_tunnel", "data_tunnel"]`
```

#### 案例 2: 多态 param 字段

**Proto 定义**:

```protobuf
message DeploySpec {
  string type = 1;
  google.protobuf.Struct param = 2;  // 动态类型，无法从 Proto 看出结构
}
```

**验证步骤**:

```python
# 搜索所有 spec type 的定义
serena.search_for_pattern("SpecType.*=|type.*Spec", relative_path="pkg/types")

# 查找每种 type 对应的 param 结构
serena.find_symbol("SpecifyAgentParam", relative_path="pkg/types", include_body=True)
serena.find_symbol("SpecifyPluginParam", relative_path="pkg/types", include_body=True)
```

### 验证检查清单

完成交叉验证后，确认以下事项:

- [ ] 所有 `repeated string` 字段已验证是否为预定义枚举
- [ ] 所有枚举字段的枚举值来自代码 `const` 定义，而非猜测
- [ ] 所有 JSON 示例中的枚举值是有效值
- [ ] 多态字段的所有变体都已文档化
- [ ] 字段的业务含义与代码注释一致

## 阶段 7: 审查和优化

**目标**: 确保文档的准确性、完整性和可读性

**检查清单**:

### 准确性检查

- [ ] 字段名称与 proto 定义一致
- [ ] 字段类型与 proto 定义一致
- [ ] 枚举值完整且正确
- [ ] URL 路径正确
- [ ] HTTP 方法正确

### 完整性检查

- [ ] 所有必填字段都已说明
- [ ] 所有可选字段都已说明
- [ ] 嵌套对象结构完整
- [ ] 多态参数的所有类型都已说明

### 可用性检查

- [ ] JSON 示例格式正确
- [ ] 示例数据合理可信
- [ ] 描述清晰易懂

### 格式检查

- [ ] Markdown 格式正确
- [ ] 表格对齐
- [ ] 代码块语法高亮

## 阶段 8: 更新任务清单

**目标**: 记录完成的工作

**步骤**:

1. 标记完成的任务
2. 记录遗留问题和后续工作

---

## 常见问题

### Q1: Proto 中使用 google.protobuf.Struct 如何文档化？

1. 查看代码中的类型定义 (`pkg/types/`)
2. 在文档中说明"根据 type 不同而不同"
3. 详细列出每种 type 对应的具体结构
4. 提供完整的 JSON 示例

### Q2: 字段的业务含义不明确？

**必须向开发者告警，不要猜测！**

1. 查阅概念文档 (`docs/concepts/`)
2. 查看代码中的验证逻辑和注释
3. 查看测试代码中的使用示例
4. **如果仍不明确，使用告警格式标记**:

```
⚠️ **待确认**: [具体问题描述]
   - 来源: [proto/类型定义/概念文档]
   - 当前理解: [你的推测]
   - 需要确认: [具体需要确认的内容]
```

### Q3: 如何提供有价值的使用示例？

1. 基于实际业务场景设计
2. 使用真实合理的字段值
3. 覆盖常见和重要的使用场景
4. 提供多个示例展示不同用法

---

## 最佳实践

### DO (应该做)

✅ 深入阅读 proto、类型定义、概念文档
✅ 使用标准模板保持一致性
✅ 提供多个实际场景的调用示例
✅ 详细说明嵌套对象和多态字段的结构
✅ 补充字段的业务含义
✅ 使用合理真实的示例数据
✅ 保持术语和命名的一致性
✅ **对 `repeated string` 字段进行交叉验证** - 检查是否为预定义枚举
✅ **使用 Serena 工具验证字段语义** - 而非仅依赖 Proto 注释

### DON'T (不应该做)

❌ 只读 proto 就开始写文档
❌ 使用过于简单或不合理的示例（如 xxx）
❌ 对复杂嵌套结构说明不清
❌ 忽略字段的业务含义
❌ 遗漏枚举值或可选字段
❌ 跳过审查环节
❌ **猜测或编造业务概念** - 不明确时必须告警
❌ **假设 `repeated string` 是自定义标签** - 必须验证是否为预定义枚举
❌ **跳过交叉验证** - Proto 定义可能具有误导性
