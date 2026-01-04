# 安装部署

## 依赖

### Mongodb
主要用于存储大部分管理数据，必须使用副本集模式`Replica Set`，最低版本`>=3.6`，建议使用`>=6.0`。

### Redis
主要用于消息队列的管理，最低版本`>=3.2`，建议使用`>=7.0`。

### Etcd
主要用于服务发现，最低版本`>=3.0`，建议使用`>=3.5.6`

## Helm部署

### Backend配置

```yaml
config:
  # 基础信息
  runMode: release   # 生产环境强制使用release, 其他runMode参数会导致性能下滑, 严禁在生产环境使用其他参数.
  tenantMode: single # 租户模式, multi或single, 单租户环境使用single即可.

  # 环境信息
  system:
    env: gse2                                                         # 当前环境的GSE环境标识, 这个会作为agent安装目录的唯一性前缀, 如/usr/local/gse2
    edition: ce                                                       # 当前环境使用的GSE版本类型, inner/ee/ce, 不同的版本对应的证书处理方式不同
  gseDeployConfs:                                                       # 支持的节点操作系统类型和对应的配置, 一般情况下支持linux/windows/darwin
    - generation: 2                                                   # GSE V2
      osType: linux                                                   # 操作系统名称
      baseWorkDir: "/tmp/bknm/"                                       # 节点操作目录
      baseDeployDir: "/usr/local/"                                    # 节点安装目录
      manualScriptPath: "/bk-nodemgr/script/manual/linux/install.sh"  # 节点手动安装脚本, 默认集成在镜像里, 无需改动
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
  basicServer:                  # 基础服务, 需要让APIGW直接访问
    bindIP: "0.0.0.0"           # 绑定IP
    port: 28102                 # 绑定端口
    authIdentity: api-gateway   # 鉴权模式, api-gateway指的是校验apigw过来的JWT
  callbackServer:               # 节点回调服务, 需要使用HostNetwork模式, 让节点直接访问
    bindIP: "0.0.0.0"  
    port: 28103
    authIdentity: none
  proxyServer:                  # 代理服务, 需要让GSE-Cluster直接访问
    bindIP: "0.0.0.0"
    port: 28104
    authIdentity: none

  # 服务实例配置
  workflow:
    workerNum: 4096              # 单台Pod的工作流上限, 直接影响Pod的服务效率
  encryptKey: "1234567890abcdef" # 内部信息对称加密密钥

  # 第三方依赖配置
  cmdb:                                            # cmdb连接配置
    supplierAccount: "0"                           # 固定参数
    user: admin                                    # 调用用户
    endpoints:                                     # APIGW调用地址
      - "https://example.com/api/bk-cmdb/prod" 
    appCode: bk-nodemgr                            # app-code
    appSecret: xxxxxx                              # app-secret
  gse:                                             # gse连接配置
    endpoints:                                     # APIGW调用地址
      - "https://example.com/api/gse/prod"     
    appCode: bk-nodemgr                            # app-code
    appSecret: xxxxxx                              # app-secret
    pluginSlotID: 0                                # 插件slot-id
    pluginSlotToken: ""                            # 插件slot-token
  userManager:                                     # 用户管理连接配置
    endpoints:                                     # APIGW调用地址
      - "https://example.com/api/bk-user/prod"
    appCode: bk-nodemgr                            # app-code
    appSecret: xxxxxx                              # app-secret
  creditVault:                                     # 第三方密码管理服务连接配置（如铁将军）
    hostCreditVault:
      enable: true
      type: "iegtjj"                               # 铁将军密码库
      iegtjj:
        endpoints:                                 # APIWG调用地址
          - "https://example.com/api/iegtjj/prod"
        appCode: bk-nodemgr                        # app-code
        appSecret: xxxxxx                          # app-secret
```

### File配置

```yaml
config:
  # 基础信息
  runMode: release   # 生产环境强制使用release, 其他runMode参数会导致性能下滑, 严禁在生产环境使用其他参数.
  tenantMode: single # 租户模式, multi或single, 单租户环境使用single即可.

  # 服务端口信息
  basicServer:          # 基础服务, 需要让Backend和Application直接访问
    bindIP: "0.0.0.0"   # 绑定IP
    port: 28202         # 绑定端口
    authIdentity: none  # 鉴权模式
  downloadServer:       # 文件下载服务, 需要让节点直接访问
    bindIP: "0.0.0.0"
    port: 28203
    authIdentity: none

  # 服务实例配置
  mountHostDir: "/data/bk-nodemgr-file-mount/" # 母机上的文件缓存目录, 将被挂载到Pod里

  # 第三方依赖配置
  repo:                                     # 制品库配置
    endpoint: "http://bkrepo.example.com/"  # 制品库地址
    projectID: ""                           # 项目ID
    repoName: ""                            # 仓库名称
    accessKey: ""                           # access key
    secretKey: ""                           # secret key
  gse:                                      # gse连接配置
    endpoints:                              # APIGW调用地址
      - "https://example.com/api/gse/prod"
    appCode: bk-nodemgr                     # app-code
    appSecret: xxxxxx                       # app-secret
    pluginSlotID: 0                         # 插件slot-id
    pluginSlotToken: ""                     # 插件slot-token
```

### Application配置

```yaml
config:
  # 基础信息
  runMode: release   # 生产环境强制使用release, 其他runMode参数会导致性能下滑, 严禁在生产环境使用其他参数.
  tenantMode: single # 租户模式, multi或single, 单租户环境使用single即可.

  # 前端配置
  front:                                
    passwordVaultSwitch: true           # 是否开启第三方密码库
    passwordVaultName: "password_vault" # 第三方密码库名字

  # 蓝鲸登录配置
  bkSaaS:
    bkLogin: 
      loginURL: "https://login.example.com" # 蓝鲸登录URL
      authType: bk_ticket                   # 登陆验证方式

  # 依赖配置
  backend:                                  # backend服务
    endpoints:                              # APIGW调用地址
      - "https://example.com/api/nodemgr/prod"     
    appCode: bk-nodemgr                     # app-code
    appSecret: xxxxxx                       # app-secret
```

### Relay配置

relay将在安装proxy的时候自动安装，无需手动配置