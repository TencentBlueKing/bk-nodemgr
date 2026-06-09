## Host（主机）

Host 是节点管理中的主机领域模型，用于承载 CMDB 主机属性、节点运行状态，以及安装、升级、重载等节点操作需要的网络参数。

## 模型分层

| 层级             | 含义        | 典型来源                         | 使用边界                          |
|----------------|-----------|------------------------------|-------------------------------|
| `Host.Static`  | 主机静态属性    | CMDB 同步；安装时也可能先写入节点管理以消除同步延迟 | 描述主机身份、归属和基础 IP 信息            |
| `Host.Dynamic` | 节点管理运行态属性 | 节点安装、升级、重载、Proxy 更新等操作写入     | 描述节点角色、状态、版本、登录信息和 Proxy 服务参数 |

静态属性不等于不可变属性，而是指其语义来源于主机资产信息；动态属性也不等于临时字段，而是节点管理维护的运行态配置。

## 主机身份与归属

| 字段                                                 | 含义                                         |
|----------------------------------------------------|--------------------------------------------|
| `host_id`                                          | CMDB 主机 ID，也是节点管理识别主机的主键。                  |
| `biz_id` / `set_id` / `module_id`                  | 主机在 CMDB 业务拓扑中的归属。                         |
| `bk_networkarea_id`                                | 管控区域 ID，用于区分 GSE 网络区域。                     |
| `bk_networkunit_id`                                | Network Unit（管控单元）ID，由节点管理维护，表示主机所属代理管控单元。 |
| `node_role`                                        | 节点角色，主要包括 Agent、Proxy。                     |
| `node_status` / `node_version` / `node_generation` | 节点运行状态、版本和 GSE 迭代。                         |

## IP 字段定义

| 字段                                       | 常见界面名称 | 所属模型                                                      | 含义                                                  |
|------------------------------------------|--------|-----------------------------------------------------------|-----------------------------------------------------|
| `bk_host_innerip` / `bk_host_innerip_v6` | 内网 IP  | `Host.Static.InnerIPList` / `Host.Static.InnerIPV6List`   | 主机的内网地址列表。                                          |
| `login_ip`                               | 登录 IP  | `Host.Dynamic.LoginIP`                                    | Backend 通过 SSH/WMI 登录目标主机时使用的地址，不要求与 CMDB 内网 IP 相同。 |
| `export_ip` / `export_ip_v6`             | 出口 IP  | `Host.Dynamic.ExportIP` / `Host.Dynamic.ExportIPV6`       | Proxy 与上游 GSE 服务通信时的出口地址，常见于公网或 NAT 场景。             |
| `advertise_ip` / `advertise_ip_v6`       | 服务 IP  | `Host.Dynamic.AdvertiseIP` / `Host.Dynamic.AdvertiseIPV6` | Proxy 暴露给下游 GSE Agent 连接的服务地址；未填写时通常按内网 IP 语义处理。    |

这些字段都描述“主机可被如何访问”，但访问方向不同：
- `login_ip` 面向 Backend 登录目标主机，
- `export_ip` 面向 Proxy 访问上游，
- `advertise_ip` 面向下游 Agent 连接 Proxy，
- `inner_ip` 是主机资产与单元内通信的基础地址。

> 特别注意：`advertise_ip` 不一定是单元内可访问的 IP，它可能是经过多级 NAT 后暴露给下游 Agent 的服务地址，因此不能替代 Proxy inner IP 用于 Relay bind/callback。

## Proxy 服务字段

| 字段                             | 含义                                       |
|--------------------------------|------------------------------------------|
| `proxy_tags`                   | Proxy 承载的能力标签，例如安装跳板、Agent 控制、文件通道、数据通道。 |
| `relay_download_port`          | Relay 插件提供 download 服务的端口。               |
| `relay_callback_port`          | Relay 插件提供 callback 服务的端口。               |
| `proxy_install_origin_unit_id` | Proxy 安装来源单元，用于跨单元安装场景。                  |
| `proxy_access_disabled`        | 是否禁止新的 Proxy 接入连接；不影响已有连接。               |

Relay 端口只描述服务监听端口，实际访问地址还需要结合 Host 的 IP 语义判断。

## Proxy Inner IP 特殊语义

Proxy 的内网 IP 不只是 CMDB 展示字段。对于 Proxy 节点，`inner_ip` 还承担单元内 Relay 服务地址的语义：

- Relay 插件 bind address 使用 Proxy 的首个 inner IP。
- Proxy 自身升级、重载时，installer 回调本机 Relay `callback` 服务，地址使用首个 inner IP + `relay_callback_port`。
- 当填写多个 inner IP 时，只有列表第一位参与当前 P-Agent 通信、Relay bind 和本机 Relay callback 选址；后续 IP 仅作为主机属性保留。

IPv4 和 IPv6 分别按各自列表处理：`bk_host_innerip[0]` 是实际 IPv4 内网通信地址，`bk_host_innerip_v6[0]` 是实际 IPv6
内网通信地址。

因此，安装或更新 Proxy 时，首个 inner IP 必须是同 Network Unit 内 Agent / Proxy 能访问的真实内网地址。若填写顺序错误，可能导致
Relay 监听或状态回写地址不可达。

## 相关文档

- [节点操作：传输模式与状态回写](control_mode.md)
- [Network Unit（管控单元）](../topo/networkunit.md)
