# 安装部署

## 依赖

### Mongodb

主要用于存储大部分管理数据，必须使用副本集模式`Replica Set`，最低版本`>=3.6`，建议使用`>=6.0`。

### Redis

主要用于消息队列的管理，最低版本`>=3.2`，建议使用`>=7.0`。

### Etcd

主要用于服务发现，最低版本`>=3.0`，建议使用`>=3.5.6`

## 权限申请

安装前需要为 `bk-nodemgr` 申请第三方 APIGateway 权限，详见 [第三方 APIGateway 权限申请](installation/thirdparty_apigateway_permission.md)。

## Helm部署

通过 Helm 初始化证书和工具包，或在部署期间手工导入 V2/V3 插件包时，参考 [初始化包导入](installation/init_packages.md)。

### 公共 tracing 配置

`Application`、`Backend`、`File` 三个服务均支持 `config.tracing` 配置。默认使用 `stdout` exporter，仅输出本地 trace；如需上报到 OpenTelemetry Collector、Tempo 或其它兼容 OTLP 的后端，将 `exporterType` 改为 `otlp`，并配置对应的 `otlpEndpoint`。

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `tracing.exporterType` | `stdout` | trace exporter 类型。可使用 `stdout` 或 `otlp`；使用 `otlp` 时需要配置 OTLP 连接参数。 |
| `tracing.otlpEndpoint` | 空 | OTLP collector/exporter endpoint。`otlpProtocol: grpc` 时填写 gRPC endpoint，例如 `otel-collector:4317`；`otlpProtocol: http` 时填写 HTTP endpoint 的 host 和 port，例如 `otel-collector:4318`，默认上报 path 为 `/v1/traces`。 |
| `tracing.otlpProtocol` | `grpc` | OTLP 传输协议，合法值为 `grpc` 或 `http`。配置为其它非空值时服务启动会失败，并返回 `otlp protocol is invalid`。 |
| `tracing.otlpInsecure` | `false` | 是否跳过 OTLP TLS 校验或使用非 TLS 连接。集群内明文 Collector 可设置为 `true`；公网或跨网络访问建议保持 `false` 并使用 HTTPS/TLS。 |
| `tracing.otlpHeaders` | `{}` | 发送 OTLP 请求时附加的 headers，常用于上游 collector 的认证或租户标识。 |

`otlpProtocol: http` 适用于网络环境不便开放 gRPC 出口、只允许 HTTP/HTTPS 出口，或上游只暴露 OTLP HTTP 接收端的场景。示例：

```yaml
config:
  tracing:
    exporterType: "otlp"
    otlpEndpoint: "otel-collector.example.com:4318"
    otlpProtocol: "http"
    otlpInsecure: true
    otlpHeaders:
      Authorization: "Bearer <token>"
```

### Backend配置

```yaml
config:
  # 基础信息
  runMode: release # 生产环境强制使用release, 其他runMode参数会导致性能下滑, 严禁在生产环境使用其他参数.
  tenantMode: single # 租户模式, multiple或single, 单租户环境使用single即可.

  # 环境信息
  system:
    env: gse2 # 当前环境的GSE环境标识, 这个会作为agent安装目录的唯一性前缀, 如/usr/local/gse2
    edition: ce # 当前环境使用的GSE版本类型, inner/ee/ce, 不同的版本对应的证书处理方式不同
  gseDeployConfs: # 支持的节点操作系统类型和对应的配置, 一般情况下支持linux/windows/darwin
    - generation: 2 # GSE V2
      osType: linux # 操作系统名称
      baseWorkDir: "/tmp/bknm/" # 节点操作目录
      baseDeployDir: "/usr/local/" # 节点安装目录
      manualScriptPath: "/bk-nodemgr/script/manual/linux/install.sh" # 节点手动安装脚本, 默认集成在镜像里, 无需改动
    - generation: 2
      osType: windows
      baseWorkDir: "c:\\tmp\\bknm\\"
      baseDeployDir: "c:\\"
      manualScriptPath: "/bk-nodemgr/script/manual/windows/install.bat"
    - generation: 2
      osType: darwin
      baseWorkDir: "/tmp/bknm/"
      baseDeployDir: "/usr/local/"
      manualScriptPath: "/bk-nodemgr/script/manual/darwin/install.sh"

  # 服务端口信息
  basicServer: # 基础服务, 需要让APIGW直接访问
    bindIP: "0.0.0.0" # 绑定IP
    port: 28102 # 绑定端口
    authIdentity: api-gateway # 鉴权模式, api-gateway指的是校验apigw过来的JWT
  callbackServer: # 节点回调服务, 需要使用HostNetwork模式, 让节点直接访问
    bindIP: "0.0.0.0"
    port: 28103
    authIdentity: none
  proxyServer: # 代理服务, 需要让GSE-Cluster直接访问
    bindIP: "0.0.0.0"
    port: 28104
    authIdentity: none

  # 服务实例配置
  workflow:
    workerNum: 4096 # 单台Pod的工作流上限, 直接影响Pod的服务效率
  encryptKey: "1234567890abcdef" # 内部信息对称加密密钥

  # 第三方依赖配置
  cmdb: # cmdb连接配置
    supplierAccount: "0" # 固定参数
    user: admin # 调用用户
    endpoints: # APIGW调用地址
      - "https://example.com/api/bk-cmdb/prod"
    appCode: bk-nodemgr # app-code
    appSecret: xxxxxx # app-secret
  gse: # gse连接配置
    endpoints: # APIGW调用地址
      - "https://example.com/api/gse/prod"
    appCode: bk-nodemgr # app-code
    appSecret: xxxxxx # app-secret
    pluginSlotID: 0 # 插件slot-id
    pluginSlotToken: "" # 插件slot-token
  userManager: # 用户管理连接配置
    endpoints: # APIGW调用地址
      - "https://example.com/api/bk-user/prod"
    appCode: bk-nodemgr # app-code
    appSecret: xxxxxx # app-secret
  creditVault: # 第三方密码管理服务连接配置（如铁将军）
    hostCreditVault:
      enable: true
      type: "iegtjj" # 铁将军密码库
      iegtjj:
        endpoints: # APIWG调用地址
          - "https://example.com/api/iegtjj/prod"
        appCode: bk-nodemgr # app-code
        appSecret: xxxxxx # app-secret

  # 默认直连网络单元自动创建
  networkUnit:
    defaultDirectUnit:
      enabled: false # 是否在同步完管控区域后, 于默认管控区域(id=0)自动创建直连网络单元
      name: "default" # 自动创建的直连网络单元名称
      clusterEndpoints: [] # 上游GSE cluster通道endpoint列表
      fileEndpoints: [] # 上游GSE file通道endpoint列表
      dataEndpoints: [] # 上游GSE data通道endpoint列表
```

> `tenantMode`的合法取值是`multiple`或`single`，不能写作`multi`

> 虚拟用户缓存说明：`tenantMode=multiple` 时，通过 `userManager` 查询的 `bk_username` 成功结果按租户和登录名在进程内缓存 1 分钟，时长不可通过配置调整。多副本之间不共享缓存，用户管理侧变更后可能短暂返回不同结果，排查方法详见 [虚拟用户变更后解析结果未立即更新](troubleshooting/virtual_user_cache.md)。

> 默认直连网络单元说明：`networkUnit.defaultDirectUnit` 仅影响**默认管控区域（id=0）**中直连网络单元的自动创建，默认关闭。同步数据工作流会在默认管控区域已从 CMDB 同步、且该区域下尚不存在直连网络单元时，按此配置自动创建一个 `is_direct=true` 的网络单元（幂等，已存在则跳过）。
>
> - **启用前提**：需确认 `clusterEndpoints` / `fileEndpoints` / `dataEndpoints` 已填写环境中真实可用的 GSE 接入地址（对应 GSE 的 cluster / file / data 三类通道）。
> - **为空风险**：若开启开关但 endpoints 为空，将创建出没有上游通道地址的直连单元，默认管控区域内的 Agent 安装与管控通道建立会失败。
>
> 网络单元概念详见 [Network Unit（管控单元）](../concepts/topo/networkunit.md)，新环境如何配置 `defaultDirectUnit` 详见 [默认直连网络单元配置](installation/default_direct_unit.md)。

### File配置

```yaml
config:
  # 基础信息
  runMode: release # 生产环境强制使用release, 其他runMode参数会导致性能下滑, 严禁在生产环境使用其他参数.
  tenantMode: single # 租户模式, multiple或single, 单租户环境使用single即可.

  # 服务端口信息
  basicServer: # 基础服务, 需要让Backend和Application直接访问
    bindIP: "0.0.0.0" # 绑定IP
    port: 28202 # 绑定端口
    authIdentity: none # 鉴权模式
  downloadServer: # 文件下载服务, 需要让节点直接访问
    bindIP: "0.0.0.0"
    port: 28203
    authIdentity: none
  exportServer: # 导出插件包的HTTP服务及公开下载地址配置
    # 填写完整的 http(s) URL，作为插件包导出下载地址的公开访问前缀，例如：https://file.example.com:28204
    # 留空时使用 advertiseIPV4:port；配置 tls.certFile 和 tls.keyFile 后自动使用 https，否则使用 http
    address: ""
    bindIP: "0.0.0.0"
    port: 28204
    authIdentity: none

  # 服务实例配置
  mountHostDir: "/data/bk-nodemgr-file-mount/" # 母机上的文件缓存目录, 将被挂载到Pod里

  # 第三方依赖配置
  repo: # 制品库配置
    endpoint: "http://bkrepo.example.com/" # 制品库地址
    projectID: "" # 项目ID
    repoName: "" # 仓库名称
    accessKey: "" # access key
    secretKey: "" # secret key
  gse: # gse连接配置
    endpoints: # APIGW调用地址
      - "https://example.com/api/gse/prod"
    appCode: bk-nodemgr # app-code
    appSecret: xxxxxx # app-secret
    pluginSlotID: 0 # 插件slot-id
    pluginSlotToken: "" # 插件slot-token
```

> BKRepo 多租户说明：`tenantMode=multiple` 时，节点管理访问 BKRepo 固定使用 `system` 租户；`repo.projectID` 仍填写原始项目 ID，例如填写 `blueking` 时实际请求项目为 `system.blueking`，不要在配置中预先填写 `system.` 前缀。

### Application配置

```yaml
config:
  # 基础信息
  runMode: release # 生产环境强制使用release, 其他runMode参数会导致性能下滑, 严禁在生产环境使用其他参数.
  tenantMode: single # 租户模式, multiple或single, 单租户环境使用single即可.

  # 前端配置
  front:
    passwordVaultSwitch: true # 是否开启第三方密码库
    passwordVaultName: "password_vault" # 第三方密码库名字

  # 蓝鲸登录配置
  bkSaaS:
    bkLogin:
      loginURL: "https://login.example.com" # 蓝鲸登录URL
      endpoints: # 蓝鲸登录后端接口调用地址
        - "https://example.com/prod"
      appCode: bk-nodemgr # APIGW app-code，多租户模式必填
      appSecret: xxxxxx # APIGW app-secret，多租户模式必填
      user: admin # APIGW 用户认证用户，多租户 un 模式必填
      authMode: "un" # APIGW 认证模式，支持 un 或 at
      accessToken: "" # APIGW access-token，at 模式必填
      authType: bk_ticket # 登陆验证方式

  # 依赖配置
  backend: # backend服务
    endpoints: # APIGW调用地址
      - "https://example.com/api/nodemgr/prod"
    appCode: bk-nodemgr # app-code
    appSecret: xxxxxx # app-secret
```

### Relay配置

relay将在安装proxy的时候自动安装，无需手动配置

### 自监控开发套件配置

> **使用范围与维护说明**：Grafana、Tempo、Alloy、Loki、Prometheus、Pyroscope、OpenTelemetry Collector（`opentelemetry-collector`，对应 `opentelemetryGateway` 配置）等自监控组件均为开源组件，仅作为用户自行管理的可选开发套件，不属于节点管理的标准能力，默认部署不提供这些组件（启用开关均为 `false`），用户应结合自身环境评估后决定是否启用。
>
> 该套件仅用于节点管理自身的指标、日志和链路追踪观测，可供未部署蓝鲸监控的环境按需使用。

#### 内嵌组件启用示例

以下配置合并到 Helm values 顶层，而不是各服务的 `config` 中。示例为主动启用开发套件，并非默认部署配置。

Trace 和指标最小启用示例：服务通过 OpenTelemetry Collector Gateway 将 trace 写入 Tempo，Prometheus 采集服务指标，Grafana 展示数据。

```yaml
tempo:
  enabled: true
opentelemetryGateway:
  enabled: true
prometheus:
  enabled: true
grafana:
  enabled: true
  rootURL: https://nodemgr.example.com/grafana/
```

日志采集最小启用示例，可独立使用或与上述配置合并：

```yaml
loki:
  enabled: true
  singleBinary:
    persistence:
      enabled: true
      storageClass: ""
      size: 8Gi
alloy:
  enabled: true
grafana:
  enabled: true
  rootURL: https://nodemgr.example.com/grafana/
```

- Grafana 至少需要启用 Tempo、Prometheus、Loki 中的一个数据源。`grafana.rootURL` 必须替换为实际访问地址，并保留结尾 `/`；该参数本身不会创建外部访问入口。
- Tempo 默认 local 存储仅适合开发或验证环境。Loki 示例使用 PVC，`storageClass` 留空时使用集群默认 StorageClass；请根据环境配置存储类和容量。以上示例不包含生产环境所需的完整安全加固、高可用和数据备份配置。
- Alloy 依赖 Loki，以 DaemonSet 运行；Chart 默认只发现 release 所在 namespace 的 Pod，并只采集容器名匹配 `bk-nodemgr-(application|backend|file)` 的日志。Helmfile 部署会显式指定三个业务 namespace；同 namespace 中同名的非节点管理容器也可能被采集。启用前需自行评估日志隐私、访问权限、存储容量及节点资源开销。
- Grafana 默认登录用户为 `admin`，管理员密码由 Chart 在 Secret 中生成。可在实际 namespace 中读取 `<release-name>-grafana` Secret 的 `admin-password` 字段并进行 Base64 解码；自定义 `fullnameOverride` 或 `nameOverride` 时以实际 Secret 名称为准。

#### 依赖自监控套件的 tracing 配置

使用内嵌 Gateway 和 Tempo 时，以下 Helm values 即可启用 trace 接入，无需手工填写服务 endpoint：

```yaml
tempo:
  enabled: true
opentelemetryGateway:
  enabled: true
```

链路追踪上报的自动配置适用于 `application.config.tracing`、`backend.config.tracing`、`file.config.tracing`，按各服务的配置分别判断：

- 未显式设置 `otlpEndpoint` 且启用 Gateway 时，服务将 trace 发往 Gateway，再由 Gateway 转发到 Tempo。
- 未显式设置 `otlpEndpoint`、未启用 Gateway，但启用 Tempo 时，服务自动直连 Tempo。
- 已显式设置 `otlpEndpoint` 时，不自动配置上报至内嵌组件；需按前述公共 tracing 配置自行设置连接参数。

自动配置链路追踪上报时使用 `exporterType: otlp`、`otlpProtocol: grpc` 和 `otlpInsecure: true`，需自行评估集群内明文传输风险。未启用 Gateway 和 Tempo 时，不会因启用 Grafana、Prometheus、Loki 或 Alloy 而自动开启 OTLP trace 上报。
