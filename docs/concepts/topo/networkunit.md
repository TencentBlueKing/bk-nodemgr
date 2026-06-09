## Network Unit（管控单元）

管控单元是节点管理的**代理管控基本单元**：同一单元内的节点经 GSE Proxy 与上游通信，由该单元内的 Proxy 承载跨网络管理能力。

## 直连与非直连

`is_direct` 标识该单元是否**可直接连通上游 GSE / 管控面服务**：

| `is_direct` | 语义 | 网络假设 |
|-------------|------|----------|
| `true`（直连单元） | 单元内节点可直接访问管控面的回调、下载等服务 | 管控面 HTTP 服务可达 |
| `false`（非直连单元） | 单元与管控面之间存在网络隔离（跨 VPC 等） | 单元内节点默认无法直接访问管控面 HTTP 服务 |

非直连单元的节点不假设能直连管控面；文件下载与状态回写经单元内 Proxy 上的 Relay 插件中转，Relay 再通过 GSE 信令与 Server 通信。详见 [架构设计 — Relay 服务](../../operation/architecture.md)。

## 相关文档

- [节点传输模式](../node/control_mode.md)
- [架构设计](../../operation/architecture.md)
