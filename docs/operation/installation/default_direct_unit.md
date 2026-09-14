# 新环境配置默认直连网络单元（defaultDirectUnit）

新环境安装时，如果**默认管控区域（id=0）内存在可直接访问管控面的节点**（即节点与 GSE 接入层之间没有网络隔离），需要在默认管控区域下提供一个**直连网络单元**（`is_direct=true`），否则默认管控区域内的 Agent 安装与管控通道建立会失败。

`bk-nodemgr` 支持通过 Backend 配置 `networkUnit.defaultDirectUnit`，在管控区域从 CMDB 同步完成后**自动创建**这个直连网络单元。该开关默认关闭。

直连 / 非直连网络单元的概念见 [Network Unit（管控单元）](../../concepts/topo/networkunit.md)。

## 什么时候需要开启

| 场景 | 是否建议开启 |
| ---- | ---- |
| 默认管控区域内节点可直连 GSE 接入层（同机房 / 同 VPC） | 开启，避免手工创建直连单元 |
| 默认管控区域与管控面之间存在网络隔离（跨 VPC 等） | 不开启，该区域应使用非直连单元 + Proxy |
| 已通过页面或 API 在默认管控区域手工创建过直连单元 | 无需开启，自动创建会因已存在直连单元而跳过（幂等） |

## 配置方法

在 Helm values 的 Backend `config` 中配置：

```yaml
backend:
  config:
    # 默认直连网络单元自动创建
    networkUnit:
      defaultDirectUnit:
        enabled: false # 是否在同步完管控区域后, 于默认管控区域(id=0)自动创建直连网络单元
        name: "default" # 自动创建的直连网络单元名称
        clusterEndpoints: [] # 上游GSE cluster通道endpoint列表
        fileEndpoints: [] # 上游GSE file通道endpoint列表
        dataEndpoints: [] # 上游GSE data通道endpoint列表
```

启用示例（endpoint 为 `host:port` 格式，下方端口仅为示例，请以环境中 GSE 接入层的实际地址为准）：

```yaml
backend:
  config:
    networkUnit:
      defaultDirectUnit:
        enabled: true
        name: "default"
        clusterEndpoints:
          - "10.0.1.1:20001"
          - "10.0.1.2:20001"
        fileEndpoints:
          - "10.0.1.1:20002"
        dataEndpoints:
          - "10.0.1.1:20003"
```

### 字段说明

| 字段 | 类型 | 说明 |
| ---- | ---- | ---- |
| `enabled` | bool | 是否启用自动创建，默认 `false` |
| `name` | string | 自动创建的直连网络单元名称，默认 `default` |
| `clusterEndpoints` | []string | 上游 GSE cluster 通道接入地址，`host:port` 格式 |
| `fileEndpoints` | []string | 上游 GSE file 通道接入地址，`host:port` 格式 |
| `dataEndpoints` | []string | 上游 GSE data 通道接入地址，`host:port` 格式 |

> 三类 endpoint 对应 GSE 的 cluster / file / data 通道，取值与环境中 GSE 接入层（GSE Cluster）对节点暴露的地址一致，可向平台管理员或 GSE 部署配置获取。

## 生效机制

配置修改并重启 Backend 后，由**管控区域定时同步工作流**（`scheduled_sync_networkarea`，每 10 分钟执行一次）驱动，流程为：

1. `sync_networkarea`：从 CMDB 同步管控区域列表。
2. `ensure_direct_networkunit`：按以下条件判断是否创建直连单元：
   - `enabled=false`：跳过，不做任何事；
   - 默认管控区域（id=0）尚未从 CMDB 同步：本轮跳过，等待下一轮；
   - 默认管控区域下**已存在任意直连单元**：跳过（幂等，不会重复创建，也不会更新已有单元）；
   - 否则：创建一个 `is_direct=true`、名称为 `name`、GSE 版本为 V2（generation=2）的网络单元，并写入三类 endpoint。

因此从 Backend 就绪到直连单元创建完成，最长需要等待约 10 分钟（默认管控区域完成同步之后的那一轮调度）。

## 验证

安装完成后，可通过以下方式确认直连单元已创建：

1. 等待一个调度周期（约 10 分钟）后，在页面「拓扑 / 管控区域」中查看默认管控区域下是否存在名称为 `name` 的直连单元；
2. 或查看定时同步工作流日志，出现「已创建默认直连网络单元, name: default」即表示创建成功；
3. 也可调用网络单元查询 API，确认默认管控区域下存在 `is_direct=true` 的单元且 endpoint 与配置一致。

## 注意事项

- **endpoint 不能为空**：若开启 `enabled` 但三类 endpoint 为空，会创建出一个没有上游通道地址的直连单元，默认管控区域内的 Agent 安装与管控通道建立会失败。
- **只影响默认管控区域**：该配置仅作用于默认管控区域（id=0），其他管控区域下的网络单元仍需手工或按需创建。
- **不做更新**：自动创建仅在默认管控区域下不存在直连单元时执行一次；后续修改 `name` / endpoint 配置**不会**同步更新已创建的单元，如需调整请在页面上直接编辑该网络单元。
- **手工触发不包含创建动作**：页面 / API 手动触发的「同步管控区域」只执行 `sync_networkarea`，直连单元的创建只发生在定时同步工作流中。
