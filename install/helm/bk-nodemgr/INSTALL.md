# Helm Chart 安装指南

本文面向首次使用原生 Helm 部署 bk-nodemgr 的运维人员，按准备环境、配置、安装和验收的顺序操作。已有部署升级请使用 [升级指南](UPGRADE.md)；文档入口见 [Chart README](README.md)。

**默认 values 不是可直接使用的环境配置**：其中包含示例地址、空凭证和示例密码。`values-example.yaml` 仅提供可选自监控和 Ingress 配置，不是完整安装配置。

本文参数以同一 Chart 包内的 `values.yaml` 和依赖 Chart 为准。在线专题链接指向源码 `master`；使用已发布版本时，请在 GitHub 切换到目标版本对应的 tag 后阅读，避免将新配置套用到旧版本。Chart 版本、`appVersion` 与实际镜像版本需分别核对，`image.tag` 和各模块 `image.tag` 可以覆盖镜像版本。

- [部署前准备](#部署前准备)
- [准备环境配置](#准备环境配置)
- [安装 Chart](#安装-chart)
- [初始化与安装后验证](#初始化与安装后验证)
- [卸载与数据保留](#卸载与数据保留)

## 部署前准备

### Kubernetes 与 Helm

- 使用 Helm 3，并确认当前 context、namespace 和部署权限。部署账户需要创建工作负载、Service、ConfigMap、Secret、PVC 及已启用组件所需的其他资源。
- 目标集群必须支持所启用模板和依赖 Chart 的 API。标准 Application Ingress 使用 `networking.k8s.io/v1`；BCS Ingress 需要预先安装对应的 Network Extension CRD 和控制器。不能将本 Chart 视为适用于所有 Kubernetes 1.12+ 集群。
- 启用自监控、KMS 或其他可选组件前，检查其依赖 Chart 的 Kubernetes 要求和 CRD 前提。先做本地渲染，再在目标环境验证 API 与控制器支持情况。
- 确保所有节点能够拉取主服务和依赖镜像；离线或私有仓库环境需准备镜像和同 namespace 的拉取凭证。

### 平台、网络与存储

| 项目          | 部署前需要确认                                                                                                                                                                                                                                                                                                                                                               |
| ------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| BlueKing 平台 | 已有可用的 APIGateway、登录、CMDB、GSE、用户管理及 BKRepo；按所选 IAM 模式准备权限中心和回调地址。Chart 不负责安装这些平台服务                                                                                                                                                                                                                                               |
| 第三方权限    | 为实际 appCode 申请 [第三方 APIGateway 资源权限](https://github.com/TencentBlueKing/bk-nodemgr/blob/master/docs/operation/installation/thirdparty_apigateway_permission.md)，并准备 [BKRepo 资源与凭证](https://github.com/TencentBlueKing/bk-nodemgr/blob/master/docs/operation/installation/thirdparty_bkrepo_resource.md)；同步节点管理自己的网关不等于获得第三方调用权限 |
| 数据库        | 使用内置依赖或明确提供外部连接配置；MongoDB 必须为副本集。版本要求见 [安装部署](https://github.com/TencentBlueKing/bk-nodemgr/blob/master/docs/operation/installation.md)                                                                                                                                                                                                    |
| 网络          | Application/Backend/File 能访问数据库和平台依赖；APIGateway 能访问 Backend basic 服务，目标节点能访问 Backend callback 和 File download 服务，GSE 能访问 Backend proxy 服务。仅配置 Web Ingress 不足以建立这些通道                                                                                                                                                           |
| 宿主机网络    | Backend 和 File 默认 `hostNetwork: true`，需允许宿主机网络并规划监听端口、调度和副本，避免同节点端口冲突。修改为 Pod 网络前需重新确认节点及 GSE 的访问路由和通告地址                                                                                                                                                                                                         |
| 本地文件      | File 将 `file.config.mountHostDir`（默认 `/data/bk-nodemgr-file-mount/`）作为 hostPath 挂载到工作目录；确认目录权限、容量、备份及调度策略。该目录不是跨节点共享存储，Pod 换节点后不会自动迁移原目录数据                                                                                                                                                                      |
| 持久化        | 内置 MongoDB 默认申请 PVC；Redis 等依赖以子 Chart 实际渲染结果为准，不能仅凭顶层 `redis.persistence` 判断。为所有已启用的 PVC 准备 StorageClass 或 PV，并检查容量、回收策略与备份方案                                                                                                                                                                                        |

端口以 `<service>.config.*Server.port` 为准。Backend 默认 basic/callback/proxy 为 `28102/28103/28104`，File 默认 download 为 `28203`；管理端口不应作为公共入口开放。详细服务配置见 [安装部署](https://github.com/TencentBlueKing/bk-nodemgr/blob/master/docs/operation/installation.md)。

## 准备环境配置

将环境差异保存为 `values-env.yaml`，不要直接修改发行包内的默认 values。以下字段使用 **Chart 原生路径**；Helmfile 环境中的 `global.appSecret`、`global.serviceJwt` 等不会自动映射到这些路径。

### 必须确认的配置

| 配置位置                                                                                                                                                        | 何时需要及如何填写                                                                                                                                                                                          |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `application.config.tenantMode`、`backend.config.tenantMode`、`file.config.tenantMode`                                                                          | 三个服务保持一致，合法值为 `single` 或 `multiple`；按目标平台的租户模式选择                                                                                                                                 |
| `application.config.backend`                                                                                                                                    | 节点管理网关的完整阶段 URL，以及具备调用权限的 `appCode`、`appSecret`、调用用户或 token                                                                                                                     |
| `backend.config.cmdb`、`backend.config.gse`、`backend.config.userManager`、`file.config.gse`                                                                    | 各平台网关的实际 `endpoints`、应用凭证和所选 `authMode` 所需认证信息；GSE 插件能力还需确认两处 `pluginSlotID/pluginSlotToken`                                                                               |
| `application.config.userManager`                                                                                                                                | 用户管理网关地址、应用凭证与调用身份；多租户用户解析依赖该配置                                                                                                                                              |
| `application.config.bkSaaS.bkLogin`、`application.config.front`                                                                                                 | 登录 API 的 `endpoints`、应用凭证、浏览器可访问的 `loginURL`、匹配平台的 `authType` 及页面使用的 BlueKing URL                                                                                               |
| `application.config.access.virtualUser`、`backend.config.access.virtualUser`                                                                                    | 如覆盖默认虚拟用户，填写环境中实际可用的用户；未指定时单租户默认 `admin`，多租户默认 `bk_admin`                                                                                                             |
| `file.config.repo`                                                                                                                                              | BKRepo 直连 `endpoint/projectID/repoName/accessKey/secretKey`，不是 APIGateway 地址。多租户模式仍填写原始 `projectID`，请求时会转换为 `system.<projectID>`                                                  |
| `backend.config.basicServer.jwtServerConfig.publicKeyPem`                                                                                                       | Backend 校验 APIGateway JWT 的 PEM 公钥，由网关平台提供，不是服务间共享密钥                                                                                                                                 |
| `*.config.adminServer.jwtServerConfig.symmetricKey`、`application.config.file.jwtClientConfig.symmetricKey`、`backend.config.file.jwtClientConfig.symmetricKey` | 配置服务间 JWT，并配置 `file.config.basicServer.jwtServerConfig.symmetricKey`；下方片段采用同一份环境专用密钥，File 客户端密钥必须与 File basic 服务端匹配                                                  |
| `backend.config.encryptKey`                                                                                                                                     | 安装前替换示例加密密钥，并安全保存；不要在已有数据的环境中随意轮换密钥或加密套件                                                                                                                            |
| `backend.config.iamV3`、`backend.config.iamV4`、`application.config.iamV3`                                                                                      | 明确选择运行时权限方案，Backend 的 V3/V4 不能同时启用；两者都关闭会跳过 Backend 授权，不应以关闭鉴权作为部署排障手段。网关地址、应用凭证、system ID 与回调地址需匹配已注册模型                              |
| `backend.config.system`、`backend.config.gseDeployConfs`                                                                                                        | 确认 GSE 环境标识、版本类型、节点操作系统及安装路径；列表覆盖时必须保留所需的完整操作系统配置                                                                                                               |
| `backend.config.networkUnit.defaultDirectUnit`                                                                                                                  | 新环境按 [默认直连网络单元配置](https://github.com/TencentBlueKing/bk-nodemgr/blob/master/docs/operation/installation/default_direct_unit.md) 准备真实 GSE 通道；默认关闭，不能仅开启开关而保留空 endpoints |
| `etcd/redis/mongodb` 或 `externalEtcd/externalRedis/externalMongodb`                                                                                            | 确定依赖部署模式、认证、TLS、存储与备份；替换所有示例密码。切换外置依赖时须关闭对应内置组件，参数见 [公共组件配置](CONFIGURATION.md#公共组件配置)                                                                                         |

本地 values 可能包含凭证，应限制文件访问权限，避免提交到仓库或粘贴完整渲染结果。普通模式的部分凭证会进入 ConfigMap 和 Helm release 记录，应同时限制 namespace 与 release 数据的读取权限。需要 KMS 时见 [KMS 凭证集成](CONFIGURATION.md#kms-凭证集成)，不要混用两种模式的凭证来源。

### 原生 Helm values 片段

以下片段展示单租户模式、服务间 JWT 和 BKRepo 的配置位置，**不是完整安装文件**。所有 `<...>` 必须替换，并按上表补齐平台 endpoints、应用凭证、登录、IAM、数据库密码及访问入口；多租户环境需同时调整三个 `tenantMode`。

```yaml
application:
  config:
    tenantMode: single
    adminServer:
      jwtServerConfig:
        symmetricKey: "<SERVICE_JWT_KEY>"
    file:
      jwtClientConfig:
        symmetricKey: "<SERVICE_JWT_KEY>"
backend:
  config:
    tenantMode: single
    encryptKey: "<ENV_ENCRYPT_KEY>"
    basicServer:
      jwtServerConfig:
        publicKeyPem: |
          <APIGATEWAY_PUBLIC_KEY_PEM>
    adminServer:
      jwtServerConfig:
        symmetricKey: "<SERVICE_JWT_KEY>"
    file:
      jwtClientConfig:
        symmetricKey: "<SERVICE_JWT_KEY>"
file:
  config:
    tenantMode: single
    adminServer:
      jwtServerConfig:
        symmetricKey: "<SERVICE_JWT_KEY>"
    basicServer:
      jwtServerConfig:
        cryptoType: symmetric
        symmetricKey: "<SERVICE_JWT_KEY>"
    repo:
      endpoint: "https://<BKREPO_HOST>"
      projectID: "<BKREPO_PROJECT>"
      repoName: "<BKREPO_REPOSITORY>"
      accessKey: "<BKREPO_ACCESS_KEY>"
      secretKey: "<BKREPO_SECRET_KEY>"
```

多份 `-f` 文件按从左到右覆盖，同名字段以后面的文件为准；map 递归合并，列表整体替换。例如覆盖 `apigwSync.extraEnvVars`、`apigwSync.config.stages` 或 `backend.config.gseDeployConfs` 时，需写出该列表所需的完整内容，不能只添加单个条目。可选配置的更多示例见同包 `values-example.yaml`，选取需要的片段合入环境文件，不要直接将整份示例用于生产。

## 安装 Chart

以下命令以 release 和 namespace 均为 `bk-nodemgr` 为例。替换仓库地址和目标 Chart 版本；release 名称应包含 `bk-nodemgr`，如更改名称，后续检查命令也需同步调整。

```shell
helm repo add bk <HELM_REPO_URL>
helm repo update bk
helm show chart bk/bk-nodemgr --version <CHART_VERSION>
helm show values bk/bk-nodemgr --version <CHART_VERSION> > values-default.yaml
```

结合目标版本的默认值完成 `values-env.yaml` 后，先在本地检查。在保存该环境文件的目录执行以下命令，下载并解压对应 Chart；提前确认 `bk-nodemgr` 子目录不存在，避免覆盖已有文件：

```shell
helm pull bk/bk-nodemgr --version <CHART_VERSION> --untar
helm lint ./bk-nodemgr -f values-env.yaml
helm template bk-nodemgr ./bk-nodemgr --namespace bk-nodemgr -f values-env.yaml > rendered.yaml
```

`rendered.yaml` 可能包含凭证和私钥，需按敏感文件保存并在检查后清理。本地 lint/template 只检查 Chart 和模板输出，不会执行初始化 Job，也不证明平台权限、镜像、存储或网络可用。

确认目标 context、配置和 Hook 副作用后安装：

```shell
helm install bk-nodemgr bk/bk-nodemgr \
  --version <CHART_VERSION> \
  --namespace bk-nodemgr --create-namespace \
  -f values-env.yaml --wait --timeout 15m
```

根据镜像拉取、数据库启动和初始化包大小调整超时。`--wait` 等待资源就绪，启用的 Hook Job 还需执行成功；安装命令返回成功不代表业务验收完成。失败时先保留错误和相关日志，再按下一节排查，不要通过反复升级或关闭鉴权盲目重试。

## 初始化与安装后验证

### 初始化和同步 Job

这些操作会写入外部平台或制品数据，安装前必须确认目标环境和执行权限。未启用的操作需由平台管理员或已有流程完成，不能因为开关关闭就省略业务前提。

| 操作                     | 实际启用条件                                       | 执行时机与准备事项                                                                                                                                                                                                                                                              |
| ------------------------ | -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 包初始化                 | `file.enabled=true` 且 `initPackages.enabled=true` | 默认仅 `post-install`；`initPackages.installOnly=false` 时也在 `post-upgrade` 执行。准备调用用户、File JWT 和镜像内包；证书包需要完整 GSE 证书材料，详见 [初始化包导入](https://github.com/TencentBlueKing/bk-nodemgr/blob/master/docs/operation/installation/init_packages.md) |
| 节点管理 APIGateway 同步 | `apigwSync.enabled=true`                           | `post-install,post-upgrade`；配置 `extraEnvVars` 中的网关、应用凭证和地址模板，以及 `config.stages` 的阶段和 Backend 地址；按 [网关同步权限](https://github.com/TencentBlueKing/bk-nodemgr/blob/master/docs/operation/installation/apigw_sync_permission.md) 准备权限           |
| IAM V3 模型同步          | `backend.config.iamV3.enable=true`                 | `post-install,post-upgrade`；配置 Backend IAM 连接与 `iamV3.system/provider`。当前 Job 不由顶层 `iamV3.enabled` 控制，不能靠该字段阻止同步                                                                                                                                      |
| IAM V4 模型同步          | `iamV4.enabled=true`                               | `post-install,post-upgrade`；凭证和网关取自 `backend.config.iamV4`，另配 `iamV4.system/provider`；与运行时 `backend.config.iamV4.enable` 独立，详见 [IAM V4 模型同步](CONFIGURATION.md#iam-v4-模型同步)                                                                                                         |

上述业务 Hook 成功后会按 `hook-succeeded` 删除，下次执行前也会删除旧同名 Hook。应保留安装输出及必要日志；**Job 不存在不代表未执行，Job 成功也不代表包或权限已满足业务要求**。包目录为空时初始化可能没有导入任何包，需继续检查包清单和默认版本。

APIGateway 的 `NO_PUB=true` 不是 dry-run，也不保证不发布：同步仍可能写入网关、阶段、资源和授权；`RELEASE_STAGES` 非空时仍会执行对应阶段的发布。开启前检查两个参数与同步目标；启用 MCP 同步时，按 [MCP 内置特性与完整覆写](CONFIGURATION.md#mcp-内置特性与完整覆写) 选择产品内置清单或各 stage 的显式定义，排查见 [APIGateway 同步 FAQ](https://github.com/TencentBlueKing/bk-nodemgr/blob/master/docs/faq/apigw-sync.md)。

### 检查资源与业务结果

```shell
helm status bk-nodemgr --namespace bk-nodemgr
kubectl get pods,deploy,statefulset,svc,pvc,jobs --namespace bk-nodemgr
kubectl get events --namespace bk-nodemgr --sort-by=.lastTimestamp
```

逐项确认：

- **资源启动**：已启用服务的 Deployment/StatefulSet 就绪，Pod 无异常重启，PVC 已绑定。若 initContainer 卡住，检查其数据库地址、认证和网络；不要只看主容器日志。
- **失败定位**：对异常 Pod 执行 `kubectl describe pod <POD_NAME> -n bk-nodemgr`，再用 `kubectl logs <POD_NAME> -c <CONTAINER_NAME> -n bk-nodemgr` 查看对应主容器或 initContainer；重启过的容器可增加 `--previous`。不要外传未脱敏日志。
- **同步结果**：检查已启用 Hook 的结果及失败日志；确认网关阶段、路由后端和授权、IAM 模型与回调实际可用。Hook 可能已清理，也需检查外部平台结果。
- **入口与登录**：配置 Application Ingress 或已有访问入口，确认 DNS/TLS、登录跳转、租户及用户身份正确，页面能够加载业务数据。Application Service 为 NodePort，按环境限制其访问范围，不要暴露 admin 端口。
- **安装能力**：确认所需证书包、工具包和 Agent/插件包已导入并设置正确的默认版本，网络区域/网络单元具有实际 GSE 通道；在授权的测试主机上验证一次安装及状态回传。

Ingress 默认关闭；配置方式见 [Ingress 访问配置](CONFIGURATION.md#ingress-访问配置)。仅用于本机检查 Application 连通性时，也可在确认 Service 名称后转发其 `basic-http` 端口：

```shell
kubectl port-forward --namespace bk-nodemgr svc/bk-nodemgr-application 8080:basic-http
```

转发后访问 `http://127.0.0.1:8080`；登录回跳仍受平台域名、Cookie 和回调配置约束，端口转发不替代正式访问入口。设置过 `fullnameOverride/nameOverride` 时以实际 Service 名称为准。

## 卸载与数据保留

确认数据备份、业务停用范围和保留策略后卸载：

```shell
helm uninstall bk-nodemgr --namespace bk-nodemgr
```

该命令删除 Helm 管理的 release 资源，不代表删除全部业务数据或完成外部平台注销：

- StatefulSet 创建的 PVC、标记保留的资源或 Hook 资源可能继续存在，按实际资源和删除策略检查；手工删除 PVC/PV 前确认回收策略，避免永久丢失数据。
- File 的 hostPath 目录不随 release 删除；需要人工按备份与清理策略处理，不能仅因 Pod 已删除就清理目录。
- 外置数据库、BKRepo 制品、APIGateway 配置和 IAM 模型/授权不会自动撤销；需由对应平台管理员确认清理范围。
- 不要将删除 namespace 作为通用清理步骤，其中可能还有其他 release、共享 Secret 或数据资源。
