# BlueKing Node Manager Helm Chart

本 Chart 用于在 Kubernetes 中部署节点管理的 Application、Backend、File 服务，可按需部署内置 MongoDB、Redis、Etcd 和自监控组件。以下文档面向使用原生 Helm 的部署运维人员。

## 开始之前

- **默认 values 不能直接用于部署**：需要填写目标环境的平台地址、凭证、租户与权限配置，并准备网络和存储。Chart 不负责安装 APIGateway、登录、CMDB、GSE、用户管理、BKRepo 等蓝鲸平台服务。
- 首次安装从 [安装指南](INSTALL.md) 开始；已有部署先读 [升级指南](UPGRADE.md)，不要直接套用新版本默认值。
- 按目标 Chart 版本阅读同包文档；Chart 版本、`appVersion` 与实际镜像版本需分别核对。在线文档及其专题链接指向 `master` 时，应切换到对应发行版本 tag。

## 按任务查阅

| 你要做什么                                         | 阅读文档                                            |
| -------------------------------------------------- | --------------------------------------------------- |
| 首次部署，准备配置并完成安装验收                   | [安装指南](INSTALL.md)                              |
| 调整平台集成、数据库、凭证或访问入口               | [配置参考](CONFIGURATION.md)                        |
| 启用节点管理自身的指标、日志、链路追踪或 Profiling | [自监控指南](MONITORING.md)                         |
| 升级已有部署、迁移配置或评估回滚                   | [升级指南](UPGRADE.md)                              |
| 排查启动、初始化或安装后业务检查失败               | [初始化与安装后验证](INSTALL.md#初始化与安装后验证) |
| 卸载并确认数据和外部资源的保留范围                 | [卸载与数据保留](INSTALL.md#卸载与数据保留)         |

## 配置文件入口

- [values.yaml](values.yaml)：当前 Chart 的默认配置；依赖组件的参数见同包 `charts/*/values.yaml`。
- [values-example.yaml](values-example.yaml)：可选自监控和 Ingress 配置片段，**不是完整安装配置**。
- 环境差异保存到自己的 `values-env.yaml`；必填项、覆盖规则及敏感信息处理见 [准备环境配置](INSTALL.md#准备环境配置)。

使用 Helmfile 部署时，请从 [Helmfile 指南](https://github.com/TencentBlueKing/bk-nodemgr/blob/master/install/helmfile/bk-nodemgr/README.md) 进入；Helmfile environment 与原生 Chart values 不是同一种配置输入。
