# sshx

## 设计意图

sshx 包提供简单易用的 SSH 连接抽象，用于执行远程命令、文件读取和文件传输，同时提供重试机制以增强在不稳定网络环境下的可靠性。

## 功能边界

1. 此包负责：SSH 连接创建与管理、远程命令执行、单次远端文件读取、SFTP 文件传输
2. 此包不负责：文件轮询与内容解析、SSH连接配置的持久化存储、SSH连接的高级功能（如端口转发、代理跳转）

`sshx.Client.ReadFile` reads a fixed Unix remote path once, while `sshx.Client.ReadLatestFile` reads the newest file matching a glob pattern once. Polling, timeout, reconnection, and file-content interpretation belong to callers such as `pkg/installer/poller`.

## 设计考量

采用构造器模式创建Client以确保配置有效性，并集成了自动重试机制以处理网络不稳定情况。支持多种认证方式（密码、私钥、无认证）以满足不同场景需求。

## 使用限制

1. 默认跳过主机密钥验证，在安全敏感环境中需谨慎使用
2. 文件传输依赖SFTP协议，对方主机需支持此协议
3. 普通命令执行、文件读取和文件传输操作没有内置重连

## 演进方向

1. 增加SSH隧道和端口转发功能支持更复杂的网络拓扑
2. 提供基于公钥指纹的主机验证机制增强安全性
