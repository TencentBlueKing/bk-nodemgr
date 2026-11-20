## integration

面向第三方平台接入开发者的接入指南，按接入能力组织：每个子目录提供一条 API-first 的端到端使用流程，并链接各接口的契约页。

### [deploy_policy](deploy_policy/README.md)

部署策略（deploy_policy）接入指南：声明目标 desired state 并执行收敛。

### [plugin_migration](plugin_migration/README.md)

插件 V2 到 V3 进程迁移接入指南：V2 进程数据迁移、状态同步、安装 V3（不启动）、停止 V2、启动 V3 的完整流程。

### [import_package](import_package/README.md)

插件包导入接入指南：通过 package 服务 API 导入插件包并读取导入结果。

### [debug_plugin](debug_plugin/README.md)

插件调试接入指南：在目标主机上以隔离方式运行一次插件调试命令并读取调试输出。

### [ipchooser](ipchooser/README.md)

IP / 主机选择器接入指南：组合 topo 接口按业务、拓扑节点、主机过滤维度查询主机。
