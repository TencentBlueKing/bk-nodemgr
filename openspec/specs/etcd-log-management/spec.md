# etcd-log-management Specification

## Purpose

将 `go.etcd.io/etcd/client/v3` 的客户端日志统一接入 `bk-nodemgr` 的 `pkg/logger` 体系，避免依赖默认标准输出，保持关键诊断语义与等级映射，并复用现有服务日志配置模型。

## Requirements

### Requirement: etcd client logs SHALL use the unified nodemgr logging pipeline
系统 SHALL 在初始化 etcd client 时显式注入受控 logger，使 etcd SDK 内部产生的连接、认证、重试与请求失败日志进入 `bk-nodemgr` 的统一日志链路，而不是依赖默认标准输出行为。

#### Scenario: etcd client is initialized in a service
- **WHEN** `application`、`backend` 或 `file` 服务创建 etcd client
- **THEN** 系统 SHALL 为该 client 注入统一 logger
- **THEN** etcd SDK 内部日志 SHALL 进入 nodemgr 统一日志输出

#### Scenario: etcd SDK emits retry warning
- **WHEN** etcd client 在重试拦截器中产生 `warn` 或 `error` 级别日志
- **THEN** 系统 SHALL 通过统一日志链路输出该日志
- **THEN** 系统 SHALL 保留原始错误语义与 message

### Requirement: etcd logs SHALL preserve original severity and diagnostic content
系统 SHALL 在将 etcd SDK 日志接入 nodemgr logger 时，保留原始关键诊断信息，并保证日志等级与 SDK 输出一一对应。

#### Scenario: emitted etcd log is forwarded to nodemgr logger
- **WHEN** 系统输出一条 etcd 相关日志
- **THEN** 日志 SHALL 保留原始 `message`
- **THEN** 日志 SHALL 保留原始 `caller` 与 `error` 等关键诊断信息

#### Scenario: emitted etcd log level is monitored
- **WHEN** etcd SDK 输出任意级别日志
- **THEN** 系统 SHALL 将该日志等级一一映射到 nodemgr logger 等级
- **THEN** 系统 SHALL 不改变该日志用于后续统一监控的等级语义

### Requirement: etcd logs SHALL reuse existing service logging configuration
系统 SHALL 复用现有服务日志配置模型管理 etcd 日志接入行为，而不单独引入 etcd 专项日志级别或开关配置。

#### Scenario: service runs with default logging policy
- **WHEN** 服务未为 etcd 日志设置额外覆盖策略
- **THEN** 系统 SHALL 复用服务现有日志配置模型
- **THEN** 系统 SHALL 不因日志纳管改变现有 etcd 访问行为

#### Scenario: service uses shared etcd initialization path
- **WHEN** `application`、`backend` 或 `file` 服务通过共享路径创建 etcd client
- **THEN** 系统 SHALL 自动应用统一 logger 注入
- **THEN** 服务无需为 etcd 日志管理增加独立服务级改造逻辑
