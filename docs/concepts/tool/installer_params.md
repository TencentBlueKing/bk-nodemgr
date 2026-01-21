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

## 相关文档

- [Relay 服务概念](../relay/README.md) - 了解跨 VPC 场景下的服务地址提供
- [架构设计](../../operation/architecture.md) - 了解整体架构和服务定位
- [安装部署](../../operation/installation.md) - 了解系统安装配置
