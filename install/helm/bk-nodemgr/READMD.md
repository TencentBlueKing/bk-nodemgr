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
