# Installer 工具参数说明

## 概述

本文档介绍 installer 工具的参数说明。Installer 是节点管理系统中用于安装和升级节点（pagent）以及插件的命令行工具。

**文档范围**：本文档目前仅覆盖部分关键参数，未来会逐步补充更多参数的说明。

## 参数列表

### 下载服务地址参数 (`--dlsvr_addr`)

**用途**：指定 installer 从何处下载安装包和资源文件。

**格式**：
- **单个地址**：`http://ip:port`
- **多个地址**：`http://ip1:port1,http://ip2:port2,http://ip3:port3`（逗号分隔）

**使用场景**：
- 所有需要下载资源的 installer 操作（如 install、upgrade）
- 自动安装时由系统自动生成并传递给 installer
- 手动安装时由管理员在命令行中指定

**相关概念**：
- [文件服务](../../operation/architecture.md) - 提供文件下载服务
- [Relay 服务](../relay/README.md) - 跨 VPC 场景下提供下载服务端点

**示例**：
```bash
# 单个地址
--dlsvr_addr http://10.0.0.1:28303

# 多个地址（高可用场景）
--dlsvr_addr http://10.0.0.1:28303,http://10.0.0.2:28303,http://10.0.0.3:28303
```

### 回调服务地址参数 (`--cbsvr_addr`)

**用途**：指定 installer 向何处上报安装状态和结果。

**格式**：
- **单个地址**：`http://ip:port`
- **多个地址**：`http://ip1:port1,http://ip2:port2,http://ip3:port3`（逗号分隔）

**使用场景**：
- 所有需要状态上报的 installer 操作
- 自动安装时由系统自动生成并传递给 installer
- 手动安装时由管理员在命令行中指定

**相关概念**：
- [后端服务](../../operation/architecture.md) - 接收状态上报的服务
- [Relay 服务](../relay/README.md) - 跨 VPC 场景下提供回调服务端点

**示例**：
```bash
# 单个地址
--cbsvr_addr http://10.0.0.1:28302

# 多个地址（高可用场景）
--cbsvr_addr http://10.0.0.1:28302,http://10.0.0.2:28302,http://10.0.0.3:28302
```

### CPU 限制参数 (`--cpu-limit`)

**用途**：限制 installer 处理 tgz 包时的 CPU 占用，降低单核机器或 cgroup 限制的容器中解压占满 CPU 的风险。

**格式**：`--cpu-limit <ratio>`，全局参数，`node`、`plugin` 子命令均可使用。

| 取值 | 行为 |
| ---- | ---- |
| 不设置或 `0`（默认） | 不限速 |
| `(0, 1)` | 将 installer 进程的 CPU 使用限制在约 `ratio` 个逻辑核，例如 `0.5` 约为半个核 |
| `>= 1` 或 `< 0` | 不限速 |

**生效范围**：
- Agent 与插件 tgz 包的解压，以及 Agent 包的压缩
- `pluginv2` 子命令不受影响
- 仅限制 tgz 内容的读写拷贝，其余步骤（如下载、配置渲染）不受影响

**注意事项**：
- 限速通过周期性休眠实现，进程 CPU 总时间不变，处理耗时会相应变长；例如 `0.5` 时耗时约为不限速的 2 倍
- 按 installer 进程自身的 CPU 时间（user + sys）计算，无法读取该时间时退回按 wall-clock 比例休眠
- 这是尽力而为的限制，不限制磁盘 IO，也不保证主机总 CPU 或监控告警不被触发
- 单次休眠上限为 1 秒，取值低于 `0.01` 时无法达到预期限速

**示例**：
```bash
# 将 installer 处理安装包时的 CPU 占用限制在约 0.5 个核
installer --cpu-limit 0.5 plugin --deploy_env test --plugin_name bk-nodemgr-relay step install --pkg_file /path/to/package.tgz
```

## 相关文档

- [Relay 服务概念](../relay/README.md) - 了解跨 VPC 场景下的服务地址提供
- [架构设计](../../operation/architecture.md) - 了解整体架构和服务定位
- [安装部署](../../operation/installation.md) - 了解系统安装配置
