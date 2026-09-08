# FAQ

## [Server 配置](server.md)

Server 与部署配置的常见问题，包括监听地址、协议栈选择，以及 `gseDeployConfs` 这类具体配置项的落地说明。

## [APIGateway 同步](apigw-sync.md)

通过 Helm 开启 `bk-nodemgr-apigw-sync` 同步镜像的配置说明，包括 `apigwSync`、发布环境、网关定义和基础验证步骤。

## [包版本状态](pkg-release-status.md)

说明包版本 `enabled` 与 `is_hidden` 的职责边界、优先级、状态组合，以及包管理页面对应的展示和筛选规则。

## [导入包格式](import-package-format.md)

说明包管理上传 Agent / Proxy / Server / 插件 / 证书 / 工具包时的 `.tgz` 目录格式要求。

## [Config Policy 配置](config_policy.md)

通过 Config Policy 修改 Agent / Proxy 日志存储路径的说明，涵盖 `logger_path`、`all_logger_path` 选项的模板位置、路径格式与生效预期。
