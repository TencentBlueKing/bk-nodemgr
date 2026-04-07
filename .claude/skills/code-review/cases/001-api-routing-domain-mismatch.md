# API 路由放置在错误的领域

**标签：** `orthogonality` `AGENTS`
**严重程度：** ⚠️ 重要
**发现日期：** 2026-04-07
**涉及文件：** `proto/backend/api/v3/topo.proto`, `internal/backend/router/api-v3/topo/constant.go`

## 问题描述

API 路由应该按**领域能力和专业性**分组，而不是按**数据存储位置**分组。当一个 API 的数据虽然存储在某个领域（如 topo），但实际的消费者和领域专家是另一个领域（如 node）时，应该将 API 放在实际使用该数据的领域中。

## 问题代码

```protobuf
// ❌ 放在 topo 服务中，但 topo 自己不消费这些数据
service TopoService {
  // DefaultDeployConstantGet provides getting default deploy constant.
  rpc DefaultDeployConstantGet(TopoDefaultDeployConstantGetReq)
      returns (TopoDefaultDeployConstantGetResp) {
    option (google.api.http) = {
      post : "/api/v3/topo/constant/default_deploy/get"
      body : "*"
    };
  }
}

// TopoDefaultDeployConstantGetResp 返回的是 node 部署运行时配置
message TopoDefaultDeployConstantGetResp {
  // ...
  message Data { 
    CustomDeployConfig default_deploy_config = 1;  // InstallerRuntime, NodeRuntime, PluginRuntime
  }
  Data data = 5;
}
```

**问题分析：**
- `CustomDeployConfig` 包含 `InstallerRuntime`、`NodeRuntime`、`PluginRuntime`，都是 node 部署运行时配置
- 虽然数据存储在 `NetworkUnit`（topo 域），但 topo 自己不消费这些数据
- 实际消费者是 `internal/backend/manager/workflowdef/node/action_inject_node_custom_deploy_config.go`
- Node 域才是理解和使用这些配置的领域专家

## 修复后

```protobuf
// ✅ 放在 node 服务中，与实际使用者对齐
service NodeService {
  // GetDefaultDeployConfig provides getting default deploy configuration.
  rpc GetDefaultDeployConfig(NodeDefaultDeployConfigGetReq)
      returns (NodeDefaultDeployConfigGetResp) {
    option (google.api.http) = {
      post : "/api/v3/node/constant/default_deploy/get"
      body : "*"
    };
  }
}
```

或者更明确的路由：

```
/api/v3/node/deploy_config/default/get
```

## 违反的原则

**1.2 正交性（Orthogonality）**：
- API 路由应该按领域职责正交分解，每个领域只负责自己理解和消费的数据
- Topo 域负责拓扑结构管理，Node 域负责节点部署配置
- 将 node 部署配置的 API 放在 topo 域违反了正交性原则

**AGENTS.md - SRP (Single Responsibility Principle)**：
> one unit=one reason to change|anchor:{internal/<service>,internal/*/router/api-v3}|pattern:{domain-split,service-private-internal}

- Topo 路由组的变更原因应该是拓扑结构变化，不应该因为 node 部署配置变化而修改
- Node 路由组才应该承担 node 部署配置相关的 API

**AGENTS.md - Cohesion**：
> each package carves and owns its domain; keep complexity local to its owning layer

- Node 域拥有部署配置的领域知识和使用场景
- 将 API 放在 node 域能保持领域内聚性

## 审查点评

虽然 `CustomDeployConfig` 存储在 `NetworkUnit` 中，但这不意味着 API 必须放在 topo 路由组。REST API 应该按**领域能力**组织，而不是按**存储位置**组织。Node 域是这些部署配置的实际消费者和领域专家，API 应该放在 `/api/v3/node/` 下，以反映真实的领域边界和使用关系。

## 如何识别此类问题

在 Design 审查阶段（步骤 2）检查：

1. **识别数据流向**：
   - 数据存储在哪里？（NetworkUnit in topo）
   - 数据被谁消费？（node workflows）
   - 谁理解这些数据的语义？（node domain）

2. **检查领域职责**：
   - 查找相关的 workflow action 或 manager 代码
   - 确认哪个领域实际使用这些数据
   - 评估 API 所在路由组是否与使用者对齐

3. **验证正交性**：
   - 如果 API 所在领域自己不消费这些数据 → 可能放错位置
   - 如果另一个领域是主要消费者 → 应该移到消费者领域

4. **参考 AGENTS.md**：
   - 检查 `Where to look` 约定
   - 确认路由放置符合领域边界定义
