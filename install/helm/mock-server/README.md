## Mock Server

此 Chart 用于在 Kubernetes 集群中通过 Helm 部署 mock-server, 为节点管理(bk-nodemgr)提供 CMDB 和 BKRepo 的 Mock API 服务, 用于测试场景

### 集群准备

开始部署前, 请准备好一套Kubernetes集群(版本1.12或更高), 并安装Helm命令行工具(3.0或更高版本)

### 安装 Chart

执行以下命令, 在集群内安装名为 `mock-server` 的 Helm release:

```shell
$ helm install mock-server ./install/helm/mock-server -n <namespace>
```

上述命令将使用默认配置在 Kubernetes 集群中部署 mock-server, 并输出相关运行信息

### 配置说明

**Chart 全局配置**

| 参数             | 类型   | 默认值               | 描述         |
| ---------------- | ------ | -------------------- | ------------ |
| image.registry   | string | hub.bktencent.com    | 镜像源地址   |
| image.repository | string | blueking/mock-server | 服务镜像     |
| image.tag        | string | Chart对应的既定版本号 | 服务镜像标签 |
| image.pullPolicy | string | IfNotPresent         | 镜像拉取策略 |

**Service 配置**

| 参数         | 类型   | 默认值    | 描述         |
| ------------ | ------ | --------- | ------------ |
| service.type | string | ClusterIP | Service 类型 |
| service.port | int    | 28400     | Service 端口 |

**Mock Server 配置**

| 参数                            | 类型   | 默认值                    | 描述                |
| ------------------------------- | ------ | ------------------------- | ------------------- |
| config.basicServer.bindIP       | string | 0.0.0.0                   | 服务监听地址        |
| config.basicServer.port         | int    | 28400                     | 服务监听端口        |
| config.log.level                | string | INFO                      | 日志级别            |
| config.bkrepoConfig.baseDir     | string | /mock-server/data/bk-repo | BKRepo 文件存储目录 |
| config.mockData                 | object | 空                        | 预置 Mock 数据      |

完整参数列表请参考 values.yaml

### 升级

通过以下命令升级 `mock-server`：

```shell
$ helm upgrade --install mock-server ./install/helm/mock-server -n <namespace>
```

### 卸载 Chart

通过以下命令卸载 `mock-server`：

```shell
$ helm uninstall mock-server -n <namespace>
```

上述命令将移除所有和 mock-server 相关的 Kubernetes 组件, 并删除 release
