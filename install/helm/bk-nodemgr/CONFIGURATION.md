# Helm Chart 配置参考

本文面向需要调整 bk-nodemgr Chart 配置的运维人员。首次安装所需的配置清单和操作流程见 [安装指南](INSTALL.md)，版本迁移见 [升级指南](UPGRADE.md)，文档入口见 [Chart README](README.md)。

配置示例使用 Chart 原生 values 路径，需合入环境覆盖文件；按目标 Chart 版本阅读，不要直接套用其他版本或 Helmfile environment 字段。下方在线专题链接指向 `master`，使用发行包时应切换至目标版本 tag。

- [全局公共配置](#全局公共配置)
- [用户 Wiki 指引链接](#用户-wiki-指引链接)
- [IAM V3 模型同步](#iam-v3-模型同步)
- [MCP 内置特性与完整覆写](#mcp-内置特性与完整覆写)
- [IAM V4 模型同步](#iam-v4-模型同步)
- [KMS 凭证集成](#kms-凭证集成)
- [公共组件配置](#公共组件配置)
- [自监控指南](MONITORING.md)
- [Application 登录接口](#application-登录接口)
- [用户标识兼容与平台升级](#用户标识兼容与平台升级)
- [Ingress 访问配置](#ingress-访问配置)

## 环境覆盖规则

多份 `-f` 文件按从左到右覆盖，同名字段以后面的文件为准；map 递归合并，列表整体替换。覆盖列表时写出全部所需条目，不会按条目名称自动合并。完整配置文件的准备方式和必填项见 [准备环境配置](INSTALL.md#准备环境配置)。

下列为关键配置，不是全量字段清单；完整参数及依赖组件配置以同包 `values.yaml` 和 `charts/*/values.yaml` 为准。

## 全局公共配置

**Chart全局配置**

| 参数             | 类型   | 默认值                     | 描述         |
| ---------------- | ------ | -------------------------- | ------------ |
| image.registry   | string | hub.bktencent.com          | 镜像源地址   |
| image.repository | string | blueking/bk-nodemgr-server | 服务镜像     |
| image.tag        | string | Chart对应的既定版本号      | 服务镜像标签 |
| image.pullPolicy | string | IfNotPresent               | 镜像拉取策略 |

## 用户 Wiki 指引链接

`backend.config.wiki` 配置 backend 用户可见任务日志中的产品 Wiki 链接，默认为 `github` 模式：

```yaml
backend:
  config:
    wiki:
      mode: github
      github:
        repositoryURL: https://github.com/TencentBlueKing/bk-nodemgr
```

`github` 模式的仓库地址不包含 tag 或文档目录；backend 根据自身版本解析结果构建发布 tag，文档位于该 tag 的 `support-files/wiki/zh/` 和 `support-files/wiki/en/` 下，中文、英文任务提示分别链接到对应语言的文档。构建标记不作为 tag 的一部分；`debug` 等无法解析为发布版本的构建回退到 `master`，此时文档可能与运行版本不同。旧配置未指定 `wiki` 时使用上述默认值。阅读该模式的指引需要用户能够访问 GitHub。

无法访问 GitHub 的环境可切换为 `application` 模式，使用 application 服务随产品提供的文档，无需访问外网：

```yaml
backend:
  config:
    wiki:
      mode: application
      application:
        prefixURL: "https://nodemgr.example.com/#/wiki/"
application:
  config:
    wiki:
      enabled: true
```

application 的 Wiki 默认关闭，需设置 `application.config.wiki.enabled: true` 才会注册 `/wiki/` 路由；设为 `false` 后不提供文档访问。该开关不改变 backend 的链接模式，关闭前需确认 backend 不再生成指向该入口的链接。

`application.prefixURL` 必须配置为用户可访问的绝对 HTTP(S) 文档 URL 前缀，而不是集群内服务地址；不得包含凭据或 query，默认空值仅适用于 `github` 模式。标准 Web 阅读页使用 `/#/wiki/` 前缀，复用现有 Web 登录流程，不增加独立登录配置。通过子路径部署时，应保留 `#` 前的部署路径，例如 `https://nodemgr.example.com/deploy/#/wiki/`。请确保用户访问入口可访问 Web 页面，并将 `/wiki/` 原文请求转发至 application 服务；通过 `/deploy/` 子路径部署时，浏览器请求 `/deploy/wiki/`，入口代理需将其映射到 application 的 `/wiki/`。

backend 在该前缀后拼接 `zh/` 或 `en/` 及相对文件路径。例如，中文 SSH 排障指引为 `https://nodemgr.example.com/#/wiki/zh/troubleshooting/unix-ssh-connection-failed.md`，由 Web 阅读页渲染 Markdown，原文仍通过现有 `/wiki/` 接口读取；无需新增开关或接口。

原有 `https://nodemgr.example.com/wiki/` 前缀仍可使用，直接链接到原始 Markdown。`/wiki/` 保持匿名、只读访问，由上述 `application.config.wiki.enabled` 开关控制；阅读页的登录流程不改变原文可免登录访问的边界，请勿在产品 Wiki 中放置敏感信息。

## IAM V3 模型同步

启用 `backend.config.iamV3.enable` 后，安装和升级会自动执行 IAM V3 模型同步 Job。`iamV3.option.allowActionReplace` 默认为 `false`；设为 `true` 时，Job 向 `do_migrate.py` 传入 `--allow-action-replace`，允许清理策略、删除旧 Action 并按目标资源绑定重建。此开关作用于本次全目录同步中所有发生资源绑定变化的 `upsert_action`，不只限于 `package_history_view`。

```yaml
iamV3:
  option:
    allowActionReplace: true
```

启用前需确认目标系统、租户及全部绑定变更。**策略删除不可逆，已有授权不会自动迁移**；同步后需重新配置并验证受影响的权限。完成改绑后将此开关恢复为 `false`，后续同步对相同绑定只执行普通更新。关闭开关或回退 Chart 无法恢复已删除的策略。

Helmfile 部署时，将上述片段合并到对应 environment 的 `overrides/backend.yaml`。服务运行时 IAM 配置保持不变。使用的 server 镜像需包含支持 `--allow-action-replace` 的迁移脚本；IAM 版本要求、异步删除及失败恢复边界见 [IAM V3 迁移工具](https://github.com/TencentBlueKing/bk-nodemgr/blob/master/support-files/bkiamv3/README.md#显式替换-action-资源绑定)。

## MCP 内置特性与完整覆写

在已有网关凭证、地址和 `config.stages` 配置上合并以下片段，可同步并启用产品内置的 MCP，无需复制资源清单：

```yaml
apigwSync:
  enabled: true
  mcp:
    enabled: true
```

`apigwSync.mcp.enabled` 默认关闭；开启时优先于 `extraEnvVars` 中的 `ENABLE_SYNC_MCP_SERVERS`，自动启用 MCP 同步。它不自动发布网关资源：`config.spec_version` 必须为 `2`，对应 stage 需已发布包含 MCP 所引用资源的版本。内置清单由同步镜像中的产品定义提供；开启特性时将其 `status` 设为 `1`，其他字段沿用产品定义，不额外更改公开性、OAuth 或调用权限。当前内置拓扑查询 MCP 使用 `OpenTopo_*` 资源、`Official` 分类，保持非公开、关闭 public client OAuth，允许 personal client token。

各 stage 的配置优先级如下，不按 MCP 名称或字段合并：

| `config.stages[].mcp_servers` | 行为                                                                     |
| ----------------------------- | ------------------------------------------------------------------------ |
| 未设置                        | 新开关开启时使用产品内置清单；关闭时不注入                               |
| 显式非空列表                  | 完整使用该列表，包括资源、状态、公开性及 OAuth 配置，不追加内置 MCP      |
| 显式 `[]`                     | 不注入内置定义，跳过该 stage 的 MCP 同步；不会删除或停用网关中已有的 MCP |

关闭新开关后，仍可通过原 `ENABLE_SYNC_MCP_SERVERS=true` 同步显式提供的完整定义。同步镜像在运行时读取部署配置和内置清单，生成只包含非空 MCP 定义的 `/data/mcp-definition.yaml`；新开关关闭且缺少定义，或显式提供空列表的 stage 不参与 MCP 同步。需要停用已有 MCP 时，应显式提供完整定义并设置 `status: 0`，不能用空列表代替。

产品内置 MCP 只维护在仓库 `apigw/mcp_servers.yaml` 的 `mcp_servers` 列表中，不包含 stage 或后端配置。新增或调整 MCP 后，正常构建并升级 apigw-sync 镜像即可，无需修改 Helm 模板或复制清单到 values。镜像中的 `/data/mcp_servers.yaml` 不受部署 ConfigMap 覆盖；所有未显式覆写的 stage 共用该清单。

直接运行同步镜像的脚本时，设置 `ENABLE_SYNC_MCP_SERVERS=true` 开启同步，另设 `ENABLE_BUILTIN_MCP_SERVERS=true` 使用内置清单；与 Chart 使用相同的装配和覆写规则。`mcp_servers.yaml` 不是可直接传给网关同步命令的完整 definition，应使用脚本生成的 `/data/mcp-definition.yaml`。

**升级注意：** 默认 `stages` 不再携带 MCP 清单。此前仅设置 `ENABLE_SYNC_MCP_SERVERS=true`、依赖 Chart 默认清单的部署，应迁移为 `apigwSync.mcp.enabled=true`；新开关会启用内置 MCP。若需保持停用状态或其他自定义策略，请在对应 stage 显式保留完整定义。已有显式 MCP 定义保持原样，不会被新开关改写。Chart 与同步镜像需配套升级，旧镜像不支持专用 MCP 同步文件。

## IAM V4 模型同步

`iamV4.option.allowActionReplace` 默认为 `false`。设为 `true` 后，同步 Job 在 Action 资源绑定冲突时自动复用 `replace_action`，以 `cascade: ["role"]` 清理角色中的旧 Action 成员，再删除重建 Action，后续角色同步恢复模板中声明的成员。自定义角色中未在模板声明的成员不会自动恢复；不级联清理 RBAC/ABAC 授权。日志输出旧/新绑定及替换风险；IAM 拒绝删除时立即停止，不自动扩大级联范围或回滚。已有授权不会自动迁移，启用前需确认目标租户和受影响权限。此行为需要使用包含角色级联清理逻辑的工具镜像。

```yaml
iamV4:
  enabled: true
  option:
    allowActionReplace: true
```

运行时 `backend.config.iamV4.enable` 与本节 Job 开关独立；注册的 `iamV4.system.id` 应与运行时 `backend.config.iamV4.systemID` 一致，`provider.host/path` 必须形成 IAM 实际可访问的回调 URL。详细说明见 [IAM V4 迁移工具](https://github.com/TencentBlueKing/bk-nodemgr/blob/master/support-files/bkiamv4/README.md)。

## KMS 凭证集成

`kms.enabled` 默认 `false`。启用后通过 External Secrets Operator（ESO）从 KMS/Vault 拉取全局与已启用模块的凭证，服务启动时读取并解密 envelope，覆盖对应凭证字段；Chart 不安装 ESO 或 KMS 服务。

启用前需完成：

- 在集群安装支持目标 Chart 所用资源版本的 ESO 控制器与 CRD，并确认其能够访问 KMS。
- 配置真实 `kms.server`、`kms.tls` 和 `kms.modules.{global,application,backend,file}` 的 `path/role/credential`；这些值来自目标环境的 KMS 凭证注册结果，不应直接沿用示例。
- 为目标 namespace 的 ServiceAccount 完成 KMS Kubernetes auth 授权。默认 `serviceAccount.create=false` 且名称为空时使用 `default`，需按实际账户绑定 role。
- 准备 TLS/mTLS 所引用的 Secret，确认 ExternalSecret 已同步、目标 Secret 已生成后再检查服务启动。ESO 刷新 Secret 不等于服务已重新加载凭证，轮换时还需确认相应服务的加载方式。

普通 values 中的空凭证不能代替 KMS 侧的注册、授权和同步；排障时只查看状态与必要的脱敏信息，不输出 Secret 或完整 envelope。

## 公共组件配置

`etcd.enabled`、`redis.enabled`、`mongodb.enabled` 默认均为 `true`。使用外部服务时，将对应开关设为 `false` 并填写相应 `external*`；只设置外置地址而不关闭内置组件，不会切换连接。内置依赖的高可用、认证和持久化由各子 Chart 配置，需结合实际渲染结果检查。

外置 Redis 片段（合入完整 `values-env.yaml`）：

```yaml
redis:
  enabled: false
externalRedis:
  type: standalone
  addrs:
    - "redis.example.com:6379"
  password: "<REDIS_PASSWORD>"
  db: 0
```

不使用 `externalRedis.host/port/username` 配置连接，当前模板读取的是 `addrs`。

**外置Etcd配置**

| 参数                   | 类型   | 默认值               | 描述                                                      |
| ---------------------- | ------ | -------------------- | --------------------------------------------------------- |
| externalEtcd.endpoints | list   | `["127.0.0.1:2379"]` | 外置 Etcd endpoint 列表，使用外置服务时必须替换为实际地址 |
| externalEtcd.username  | string | 空                   | 外置etcd服务username                                      |
| externalEtcd.password  | string | 空                   | 外置etcd服务password                                      |
| externalEtcd.tls.ca    | string | 空                   | 外置etcd服务的客户端证书CA内容(base64编码)                |
| externalEtcd.tls.cert  | string | 空                   | 外置etcd服务的客户端证书Cert内容(base64编码)              |
| externalEtcd.tls.key   | string | 空                   | 外置etcd服务的客户端证书Key内容(base64编码)               |

**外置Redis配置**

| 参数                     | 类型   | 默认值     | 描述                                                                                       |
| ------------------------ | ------ | ---------- | ------------------------------------------------------------------------------------------ |
| externalRedis.type       | string | standalone | 外置 Redis 类型：`standalone`、`cluster` 或 `sentinel`                                     |
| externalRedis.addrs      | list   | []         | `host:port` 列表；standalone 填单个服务地址，cluster 填节点地址，sentinel 填 Sentinel 地址 |
| externalRedis.password   | string | 空         | 外置 Redis 服务密码                                                                        |
| externalRedis.db         | int    | 0          | standalone/sentinel 数据库编号，cluster 模式忽略                                           |
| externalRedis.masterName | string | 空         | sentinel 模式的 master 名称                                                                |
| externalRedis.tls.ca     | string | 空         | 外置redis服务的客户端证书CA内容(base64编码)                                                |
| externalRedis.tls.cert   | string | 空         | 外置redis服务的客户端证书Cert内容(base64编码)                                              |
| externalRedis.tls.key    | string | 空         | 外置redis服务的客户端证书Key内容(base64编码)                                               |

**外置MongoDB配置**

| 参数                           | 类型   | 默认值          | 描述                                                              |
| ------------------------------ | ------ | --------------- | ----------------------------------------------------------------- |
| externalMongodb.replicaSetName | string | rs0             | 外置mongodb服务replica set名称                                    |
| externalMongodb.hosts          | array  | `[""]`          | 外置 MongoDB 地址列表；mongodb.enabled=false 时必须替换为实际地址 |
| externalMongodb.username       | string | nodemgr         | 外置mongodb服务用户名                                             |
| externalMongodb.password       | string | defaultpassword | 外置mongodb服务密码                                               |
| externalMongodb.database       | string | nodemgr         | 外置mongodb服务DB名称                                             |
| externalMongodb.authSource     | string | admin           | 外置mongodb服务认证数据库                                         |
| externalMongodb.authMechanism  | string | SCRAM-SHA-256   | 外置mongodb服务认证机制                                           |
| externalMongodb.tls.ca         | string | 空              | 外置mongodb服务的客户端证书CA内容(base64编码)                     |
| externalMongodb.tls.cert       | string | 空              | 外置mongodb服务的客户端证书Cert内容(base64编码)                   |
| externalMongodb.tls.key        | string | 空              | 外置mongodb服务的客户端证书Key内容(base64编码)                    |

## 自监控配置

自监控组件的配置、采集链路、存储限制及验证要求见 [自监控指南](MONITORING.md)。

## Application 登录接口

`application.config.bkSaaS.bkLogin.useLegacyAPI` 默认为 `false`。保持默认值且 `authType=bk_token` 时，所有租户模式均使用 `/open/bk-tokens/userinfo/`，需要配置可访问该接口的 `endpoints` 和网关应用凭证 `appCode`、`appSecret`，以及所选 `authMode` 所需的认证参数。`loginURL` 仍用于浏览器登录跳转，不能代替 API 网关地址。

`useLegacyAPI: true` 仅支持 `application.config.tenantMode=single`，不支持多租户。`authType=bk_ticket` 的接口与认证行为不变。旧版单租户环境的影响范围、迁移示例和验证要求见 [v3.0.1-alpha.114 升级说明](UPGRADE.md#v301-alpha114)。

## 用户标识兼容与平台升级

`backend.config.userManager.useLegacyUsername` 和 `application.config.userManager.useLegacyUsername` 均默认为 `false`，必须在两个服务中同步配置。该开关仅控制是否沿用 `bk_username = login_name` 的旧用户标识规则，与 `tenantMode` 独立；`tenantMode` 仍负责租户模式，不用于判断用户标识体系。

以下 7.2/7.3 指**蓝鲸平台大版本**，不是节点管理版本：

| 蓝鲸平台 | 租户模式             | 两个服务的 `userManager.useLegacyUsername` | 用户标识处理                                                       |
| -------- | -------------------- | ------------------------------------------ | ------------------------------------------------------------------ |
| 7.2      | 非多租户（`single`） | `true`                                     | 沿用 `bk_username = login_name`，不调用 lookup                     |
| 7.3      | `single`             | `false`                                    | 按当前租户通过用户管理 lookup 将 `login_name` 映射为 `bk_username` |
| 7.3      | `multiple`           | `false`                                    | 按当前租户通过用户管理 lookup 将 `login_name` 映射为 `bk_username` |

仍运行蓝鲸 7.2 平台时，即使升级节点管理，也必须保留两个服务的 `useLegacyUsername: true`：

```yaml
backend:
  config:
    userManager:
      useLegacyUsername: true
application:
  config:
    userManager:
      useLegacyUsername: true
```

蓝鲸平台从 7.2 升级到 7.3 时，必须同步将两个服务切换为 `false`，并为节点管理应用开通用户管理 `/open/tenant/virtual-users/-/lookup/` 的调用权限，确保 `userManager.endpoints`、应用凭证及请求租户标识与目标环境一致。7.3 关闭多租户后的 `single` 模式不等同于 7.2 旧版，仍需 lookup。lookup 请求失败、无权限或未找到用户时会返回错误，**不会回退为 `bk_username = login_name`**；应修复权限、租户或用户数据，而不是通过开启兼容开关绕过。

使用 Helmfile 时，在 environment 的 `backend.config.userManager` 和 `application.config.userManager` 中同步设置该字段；Docker Compose 在两个服务的配置模板中提供固定默认值 `false`，部署到蓝鲸 7.2 前须同步改为 `true`。本开关不替代 `application.config.bkSaaS.bkLogin.useLegacyAPI`，登录接口的兼容配置仍按上一节选择。

## Ingress 访问配置

Ingress 默认关闭。按访问对象区分为三类：`application.ingress` 用于节点管理 Web/API 服务入口，`backend.ingress` 用于 backend TCP 服务入口，`grafana.ingress` 用于自监控 Grafana 入口。

**Application Ingress**

`application.ingress` 支持两种模式：`className` 为空时使用 BCS Network Extension Ingress；`className` 非空时使用标准 Kubernetes Ingress。

| 参数                                | 类型   | 默认值              | 描述                                            |
| ----------------------------------- | ------ | ------------------- | ----------------------------------------------- |
| application.ingress.enabled         | bool   | false               | 是否创建 application Ingress                    |
| application.ingress.lbid            | string | 空                  | BCS Network Extension 负载均衡实例 ID           |
| application.ingress.lbPolicy        | string | WRR                 | BCS 负载均衡策略，可选 `WRR`、`LEAST_CONN`      |
| application.ingress.isDirectConnect | bool   | true                | BCS Network Extension 是否使用直连模式          |
| application.ingress.certID          | string | 空                  | BCS HTTPS 证书 ID，配置后使用 HTTPS 监听        |
| application.ingress.domain          | string | nodemgr.example.com | 访问域名                                        |
| application.ingress.path            | string | /                   | 访问路径                                        |
| application.ingress.className       | string | 空                  | 标准 Kubernetes IngressClass 名称，例如 `nginx` |
| application.ingress.tlsSecretName   | string | 空                  | 标准 Kubernetes Ingress TLS Secret 名称         |

标准 Kubernetes Ingress 示例：

```yaml
application:
  ingress:
    enabled: true
    className: nginx
    domain: nodemgr.example.com
    path: /
    tlsSecretName: nodemgr-tls
```

BCS Network Extension Ingress 示例：

```yaml
application:
  ingress:
    enabled: true
    lbid: <bcs-lb-id>
    lbPolicy: WRR
    isDirectConnect: true
    certID: <bcs-cert-id>
    domain: nodemgr.example.com
    path: /
```

**Backend Ingress**

`backend.ingress` 创建的是 BCS Network Extension TCP Ingress，会暴露 `backend.config.basicServer.port` 和 `backend.config.proxyServer.port`。

| 参数                            | 类型   | 默认值 | 描述                                       |
| ------------------------------- | ------ | ------ | ------------------------------------------ |
| backend.ingress.enabled         | bool   | false  | 是否创建 backend TCP Ingress               |
| backend.ingress.lbid            | string | 空     | BCS Network Extension 负载均衡实例 ID      |
| backend.ingress.lbPolicy        | string | WRR    | BCS 负载均衡策略，可选 `WRR`、`LEAST_CONN` |
| backend.ingress.isDirectConnect | bool   | true   | 是否使用直连模式                           |

```yaml
backend:
  ingress:
    enabled: true
    lbid: <bcs-lb-id>
    lbPolicy: WRR
    isDirectConnect: true
```

**Grafana Ingress**

`grafana.enabled=true` 只会创建 Grafana 服务，默认不会创建外部访问入口。需要通过 Ingress 访问 Grafana 时，请同时启用 `grafana.ingress.enabled=true`，并确保 `grafana.rootURL` 与 Ingress 的域名和路径一致。

| 参数                             | 类型   | 默认值              | 描述                            |
| -------------------------------- | ------ | ------------------- | ------------------------------- |
| grafana.ingress.enabled          | bool   | false               | 是否创建 Grafana Ingress        |
| grafana.ingress.ingressClassName | string | 空                  | IngressClass 名称，例如 `nginx` |
| grafana.ingress.hosts            | list   | chart-example.local | Grafana 访问域名                |
| grafana.ingress.path             | string | /                   | Grafana 访问路径                |
| grafana.ingress.pathType         | string | Prefix              | Ingress 路径匹配类型            |
| grafana.ingress.tls              | list   | []                  | HTTPS 证书配置                  |

```yaml
grafana:
  enabled: true
  rootURL: https://nodemgr.example.com/grafana/
  ingress:
    enabled: true
    ingressClassName: nginx
    hosts:
      - nodemgr.example.com
    path: /grafana
    pathType: Prefix
    tls:
      - secretName: nodemgr-grafana-tls
        hosts:
          - nodemgr.example.com
```

> 注: 使用子路径访问 Grafana 时，`grafana.rootURL` 必须包含相同子路径并以 `/` 结尾，例如 `https://nodemgr.example.com/grafana/`。如果只使用根路径访问，可将 `grafana.ingress.path` 设置为 `/`，并将 `grafana.rootURL` 设置为对应域名根路径。
