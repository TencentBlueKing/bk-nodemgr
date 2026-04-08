# node

`node` 路由组统一挂载在 `/node` 前缀下，用于承载节点安装、升级、重试、离线处理、部署常量查询等节点执行相关 API。

入口在 `node.go`，本身只负责聚合子路由，不承载具体业务逻辑。当前 `node` 下的 router 子包如下。

| 子包 | 路由前缀 | 作用概述 |
| --- | --- | --- |
| `workflow` | `/node/workflow` | 节点任务流查询与控制。负责查询 workflow、operation、operation instance，并提供 retry、terminate、manual install、offline install result 回传、日志查询、状态分布统计等接口。 |
| `agent` | `/node/agent` | 普通 Agent 生命周期操作。负责 Agent 的 install、install_check、upgrade、upgrade_check、reconfig、restart、uninstall，以及 assign unit 等接口。 |
| `proxy` | `/node/proxy` | Proxy 生命周期操作。职责与 `agent` 类似，但面向 Proxy 节点，提供 install、install_check、upgrade、upgrade_check、restart、reconfig、update、uninstall 等接口。 |
| `constant` | `/node/constant` | 部署常量查询。根据 `generation` 和 `os_type` 返回节点与插件安装所需的默认 deploy constant。 |

## 包职责边界

### workflow

`workflow` 关注“任务过程”，而不是“直接下发动作”。

适合放在这里的接口：

- 查询某次节点安装/升级任务的 workflow 和 operation
- 查询 operation instance、执行日志、状态分布
- 对已有任务执行 retry、terminate
- 处理 manual install / offline install 相关的信息获取与结果回传

如果接口的核心语义是“查看任务执行过程”或“控制已经创建的任务流”，应优先放到 `workflow`。

### agent

`agent` 关注普通 Agent 的生命周期变更，直接对应节点 Agent 本身的安装、升级、重配、重启、卸载等动作。

适合放在这里的接口：

- 新建 Agent 安装或升级动作
- 安装前校验、升级前校验
- Agent 重配、重启、卸载
- Agent 与网络单元等归属关系调整

如果接口的核心对象是“普通 Agent”，应放到 `agent`。

### proxy

`proxy` 关注 Proxy 节点的生命周期管理。它与 `agent` 的操作类型接近，但对象不同，通常用于 Proxy 节点专属的安装、升级和维护流程。

适合放在这里的接口：

- Proxy 安装、升级、更新
- Proxy 安装前或升级前检查
- Proxy 重启、重配、卸载

如果接口的核心对象是“Proxy 节点/Proxy 进程”，应放到 `proxy`，不要混入 `agent`。

### constant

`constant` 不负责创建任务，也不负责执行节点动作，只负责返回部署时需要的默认常量配置。

适合放在这里的接口：

- 按不同 `generation` 查询默认部署参数
- 按不同 `os_type` 查询节点和插件安装常量

如果接口的输出是“给前端或调用方提供默认部署配置”，应放到 `constant`。

## 放置建议

- 新增“节点任务查询/任务控制”接口：放 `workflow`
- 新增“普通 Agent 生命周期操作”接口：放 `agent`
- 新增“Proxy 生命周期操作”接口：放 `proxy`
- 新增“部署默认值/常量查询”接口：放 `constant`

这样划分后，`node` 目录下的 router 结构不是按实现细节拆分，而是按节点领域里的“任务过程 / Agent 对象 / Proxy 对象 / 部署常量”四类职责拆分。
