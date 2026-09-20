## Blueking NodeMgr

此Chart用于在Kubernetes集群中通过helm部署蓝鲸智云节点管理服务(bk-nodemgr)

### K8S集群准备

开始部署前，请准备好一套Kubernetes集群（版本1.12或更高），并安装Helm命令行工具（3.0或更高版本）

### 安装Chart

安装bk-nodemgr，你必须先添加一个有效的Helm repo仓库

```shell
## 请将 `<HELM_REPO_URL>` 替换为本 Chart 所在的 Helm 仓库地址
$ helm repo add bk <HELM_REPO_URL>
```

添加仓库成功后，执行以下命令，在集群内安装名为`bk-nodemgr`的Helm release（使用默认项目配置）：

```shell
$ helm install bk-nodemgr bk/bk-nodemgr
```

> 注: ChartName为bk-nodemgr，基于模板设计规则，请保证部署的ReleaseName中包含ChartName

上述命令将使用默认配置在Kubernetes集群中部署bk-nodemgr, 并输出相关运行信息。

### 配置说明

下面展示了可配置的参数列表以及默认值

#### 全局公共配置

**Chart全局配置**

| 参数             | 类型   | 默认值                     | 描述         |
| ---------------- | ------ | -------------------------- | ------------ |
| image.registry   | string | hub.bktencent.com          | 镜像源地址   |
| image.repository | string | blueking/bk-nodemgr-server | 服务镜像     |
| image.tag        | string | Chart对应的既定版本号      | 服务镜像标签 |
| image.pullPolicy | string | IfNotPresent               | 镜像拉取策略 |

#### 公共组件配置

**外置Etcd配置**

| 参数                   | 类型   | 默认值 | 描述                                         |
| ---------------------- | ------ | ------ | -------------------------------------------- |
| externalEtcd.endpoints | list   | 空     | 外置etcd服务endpoint地址, 多个以列表形式声明 |
| externalEtcd.username  | string | 空     | 外置etcd服务username                         |
| externalEtcd.password  | string | 空     | 外置etcd服务password                         |
| externalEtcd.tls.ca    | string | 空     | 外置etcd服务的客户端证书CA内容(base64编码)   |
| externalEtcd.tls.cert  | string | 空     | 外置etcd服务的客户端证书Cert内容(base64编码) |
| externalEtcd.tls.key   | string | 空     | 外置etcd服务的客户端证书Key内容(base64编码)  |

**外置Redis配置**

| 参数                   | 类型   | 默认值  | 描述                                          |
| ---------------------- | ------ | ------- | --------------------------------------------- |
| externalRedis.type     | string | cluster | 外置redis部署类型                             |
| externalRedis.host     | string | 空      | 外置redis服务地址                             |
| externalRedis.port     | int    | 6379    | 外置redis服务端口                             |
| externalRedis.username | string | default | 外置redis服务用户名                           |
| externalRedis.password | string | 空      | 外置redis服务密码                             |
| externalRedis.tls.ca   | string | 空      | 外置redis服务的客户端证书CA内容(base64编码)   |
| externalRedis.tls.cert | string | 空      | 外置redis服务的客户端证书Cert内容(base64编码) |
| externalRedis.tls.key  | string | 空      | 外置redis服务的客户端证书Key内容(base64编码)  |

**外置MongoDB配置**

| 参数                           | 类型   | 默认值          | 描述                                                     |
| ------------------------------ | ------ | --------------- | -------------------------------------------------------- |
| externalMongodb.replicaSetName | string | rs0             | 外置mongodb服务replica set名称                           |
| externalMongodb.hosts          | array  | 空              | 外置mongodb服务地址列表；mongodb.enabled=false时必须配置 |
| externalMongodb.username       | string | nodemgr         | 外置mongodb服务用户名                                    |
| externalMongodb.password       | string | defaultpassword | 外置mongodb服务密码                                      |
| externalMongodb.database       | string | nodemgr         | 外置mongodb服务DB名称                                    |
| externalMongodb.authSource     | string | admin           | 外置mongodb服务认证数据库                                |
| externalMongodb.authMechanism  | string | SCRAM-SHA-256   | 外置mongodb服务认证机制                                  |
| externalMongodb.tls.ca         | string | 空              | 外置mongodb服务的客户端证书CA内容(base64编码)            |
| externalMongodb.tls.cert       | string | 空              | 外置mongodb服务的客户端证书Cert内容(base64编码)          |
| externalMongodb.tls.key        | string | 空              | 外置mongodb服务的客户端证书Key内容(base64编码)           |

#### 自监控配置

自监控组件默认关闭，按需启用。Grafana 只负责展示，需要同时启用至少一个内嵌数据源：`tempo.enabled=true` 或 `prometheus.enabled=true`。

| 参数                         | 类型   | 默认值 | 描述                                                           |
| ---------------------------- | ------ | ------ | -------------------------------------------------------------- |
| tempo.enabled                | bool   | false  | 启用内嵌 Tempo，用于接收和查询链路追踪数据                     |
| opentelemetryGateway.enabled | bool   | false  | 启用 OpenTelemetry Collector Gateway，用于集中接收和转发 trace |
| prometheus.enabled           | bool   | false  | 启用内嵌 Prometheus，用于采集 bk-nodemgr 服务指标              |
| grafana.enabled              | bool   | false  | 启用内嵌 Grafana，用于查看 Tempo 和 Prometheus 数据源          |
| grafana.rootURL              | string | 空     | Grafana 访问地址，启用 Grafana 时必填，需包含结尾 `/`          |
| grafana.adminUser            | string | admin  | Grafana 管理员用户名                                           |

最小启用示例：

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

> 注: `tempo.enabled=true` 的默认 local 存储仅适合开发或验证环境；生产环境建议参考 `values-example.yaml` 配置对象存储。`grafana.rootURL` 需要和实际 Ingress、网关或端口转发访问路径保持一致。

启用 Grafana 后，登录用户默认为 `admin`。管理员密码由 Grafana Chart 在 Secret 中生成并在升级时复用。若 ReleaseName 为 `bk-nodemgr`，默认 Secret 名称为 `bk-nodemgr-grafana`，可通过以下命令获取：

```shell
$ kubectl get secret -n <namespace> bk-nodemgr-grafana -o jsonpath='{.data.admin-password}' | base64 -d; echo
```

如果使用了其它 ReleaseName，请将 Secret 名称中的 `bk-nodemgr` 替换为实际 ReleaseName，例如 `<release-name>-grafana`。如果配置了 `grafana.fullnameOverride` 或 `grafana.nameOverride`，请以实际生成的 Secret 名称为准。

#### Ingress 访问配置

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

### 滚动升级

通过以下命令滚动升级`bk-nodemgr`:

```shell
$ helm upgrade --install bk-nodemgr bk/bk-nodemgr
```

### 卸载Chart

通过以下命令卸载`bk-nodemgr`:

```bash
helm uninstall bk-nodemgr
```

上述命令将移除所有和蓝鲸智云节点管理相关的Kubernetes组件，并删除release。
