# 标准插件v3文件结构与规范

## 目录结构

``` shell
{{plugin_name}}-{{plugin_version}}.tgz
.
└──{{plugin_name}}
    ├── plugins_{{os_type}}_{{arch}}
    │   ├── definition.yaml
    │   ├── etc
    │   ├── bin
    │   │   └── {{plugin_name}}
    │   └── templates
    │       └── {{plugin_template_name}}.template
    │
    └── project.yaml
```

## 说明

- {{plugin_name}}-{{plugin_version}}.tgz：插件包压缩文件，命名规范为`插件名称-插件版本号.tgz`。
- `plugins_{{os_type}}_{{arch}}`：插件包目录，命名规范为该目录下插件支持的操作系统和系统架构，目录内包含插件的定义文件、可执行文件和配置模板文件。
  - `definition.yaml`：插件定义文件，描述插件的可定义变量信息、模板文件信息与插件控制信息。
  - `bin`：存放插件的可执行文件目录。
    - `{{plugin_name}}`：插件的主可执行文件，windows下需加上`.exe`后缀。
  - `etc`：存放插件配置文件目录。
  - `templates`：存放插件配置模板文件目录。
    - `{{plugin_template_name}}.template`：插件的配置模板文件，使用Go模板语法编写，后缀强制为`.template`。
- `project.yaml`：插件项目文件，描述插件的基本信息。

## nodemgr渲染模板的变量

在插件配置模板文件中，可以使用以下变量进行渲染：

```yaml
PluginInfo:
    LogPath: 插件日志路径
    DataPath: 插件数据路径
    PidPath: 插件PID文件路径
    SetupPath: 插件安装路径
    HostIDPath: 主机ID文件路径
    PluginIPC: PluginIPC路径，windows为端口号，linux为文件路径
    DataIPC: DataIPC路径，windows为端口号，linux为文件路径
    AgentDir: Agent安装路径
    GroupID: 插件所属的Group ID
    SubConfigPath: 子配置文件存放路径
    IsMultiTenant: 是否为多租户插件，true或false
NodeInfo:
    HostID: 主机ID
    HostName: 主机名称
    BizID: 业务ID
    NetworkAreaID: 管控区域ID
    NetworkAreaName: 管控区域名称
    NetworkUnitID: 管控单元ID
    NetworkUnitName: 管控单元名称
    OSType: 操作系统类型
    Arch: 操作系统架构
    InnerIPList: 内网IP列表
    OuterIPList: 外网IP列表
    InnerIPv6List: 内网IPv6列表
    OuterIPv6List: 外网IPv6列表
    AgentID: Agent ID
    CPUNum: CPU核数
    MemCap: 内存容量
    NodeRole: 节点角色，如agent、proxy等
    LoginIP: 登录IP
    LoginUser: 登录用户
PreDefinitionConstants:
    Global: 全局插件配置，定义于nodemgr的backend配置文件gseDeployConfs.pluginCustom.commonConstants.Global中
    Unique: 属于该插件的的特别配置，定义于nodemgr的backend配置文件gseDeployConfs.pluginCustom.commonConstants.{{plugin_name}}中
CustomContext:
    # 自定义变量，安装时由用户在请求参数中提供，请规范使用驼峰命名法定义变量名称
    # variables的配置内容也会出现在这里
```

## 示例

以`bk-nodemgr-relay`插件为例，插件包结构如下：

``` shell
bk-nodemgr-relay-1.0.0.tgz
.
└─bk-nodemgr-relay
    ├── plugins_linux_x86_64
    │     ├── definition.yaml
    │     ├── etc
    │     ├── bin
    │     │     └── bk-nodemgr-relay
    │     └── templates
    │         └── bk-nodemgr-relay.template
    └── project.yaml
```

`project.yaml`文件内容如下：

``` yaml
title: bk-nodemgr-relay
version: 1.0.0
description: 蓝鲸节点管理中继模块
descriptionEn: BlueKing Node Management Relay Module
scenario: 蓝鲸节点管理中继模块，用于节点管理在多级 proxy 场景下的对 agent 的操作命令转发
scenarioEn: BlueKing Node Management Relay Module, used for forwarding operation commands to agents in multi-level proxy scenarios in node management
launchNode: proxy
templateRenderer: go-template
```

- name：插件名称
- version：插件版本
- description：插件描述信息
- descriptionEn：插件英文描述信息
- scenario：插件使用场景描述
- scenarioEn：插件使用场景英文描述
- launchNode：插件启动节点类型, 支持 proxy（Proxy节点）、agent（Agent节点）、all（所有节点）
- templateRenderer：配置模板渲染器类型, 支持 go-template（Go模板，推荐）、 jinja2（Jinja2模板，不推荐）

> **注意**
> 配置模板渲染器请选择 go-template，因为 jinja2 模板渲染器已不推荐使用，未来版本可能会移除对 jinja2 模板的支持
> go-template后续会提供更多的功能和优化。

`definition.yaml`文件内容如下：

``` yaml
configTemplates:
  - name: bk-nodemgr-relay.conf
    isMainConfig: true
    filePath: etc
    sourcePath: templates/bk-nodemgr-relay.template
    variables:
      - title: Log
        type: object
        default:
        required: false
        description: 日志配置
        descriptionEn: Log Configuration
        properties:
          - title: LogDir
            type: string
            default: ""
            required: false
            description: 日志目录
            descriptionEn: Log Directory
            properties:
          - title: LogLevel
            type: string
            default: INFO
            required: false
            description: 日志级别
            descriptionEn: Log Level
            properties:
          - title: LogMaxSizeMB
            type: number
            default: 100
            required: false
            description: 日志文件最大大小（MB）
            descriptionEn: Log Max Size (MB)
            properties:
          - title: LogMaxNum
            type: number
            default: 7
            required: false
            description: 最大日志文件数量
            descriptionEn: Log Max File Number
            properties:
          - title: LogToStderr
            type: bool
            default: false
            required: false
            description: 是否将日志输出到标准错误
            descriptionEn: Whether to output logs to standard error
            properties:
          - title: LogAlsoToStderr
            type: bool
            default: false
            required: false
            description: 是否同时将日志输出到标准错误
            descriptionEn: Whether to also output logs to standard error
            properties:
      - title: Tracing
        type: object
        default:
        required: false
        description: 分布式追踪配置
        descriptionEn: Distributed Tracing Configuration
        properties:
          - title: ExporterType
            type: string
            default: stdout
            required: false
            description: 追踪导出器类型
            descriptionEn: Tracing Exporter Type
            properties:
          - title: OtlpEndpoint
            type: string
            default: ""
            required: false
            description: OTLP端点
            descriptionEn: OTLP Endpoint
            properties:
          - title: OtlpInsecure
            type: bool
            default: false
            required: false
            description: 是否使用不安全的连接
            descriptionEn: Whether to use insecure connection
            properties:
          - title: OtlpHeaders
            type: string
            default: ""
            required: false
            description: OTLP头信息
            descriptionEn: OTLP Headers
            properties:
      - title: InfoServer
        type: object
        default:
        required: false
        description: 信息服务器信息
        descriptionEn: Info Server Information
        properties:
          - title: BindIP
            type: string
            default: "127.0.0.1"
            required: false
            description: 绑定IP地址
            descriptionEn: Bind IP Address
            properties:
          - title: Port
            type: number
            default: 28300
            required: false
            description: 绑定端口
            descriptionEn: Bind Port
            properties:
          - title: AdvertiseIPV4
            type: string
            default: "127.0.0.1"
            required: false
            description: 广播IPv4地址
            descriptionEn: Advertise IPv4 Address
            properties:
          - title: AdvertiseIPV6
            type: string
            default: "::1"
            required: false
            description: 广播IPv6地址
            descriptionEn: Advertise IPv6 Address
            properties:
          - title: AuthIdentity
            type: string
            default: none
            required: false
            description: 认证身份
            descriptionEn: Authentication Identity
            properties:
      - title: AdminServer
        type: object
        default:
        required: false
        description: 管理服务器信息
        descriptionEn: Admin Server Information
        properties:
          - title: BindIP
            type: string
            default: "127.0.0.1"
            required: false
            description: 绑定IP地址
            descriptionEn: Bind IP Address
            properties:
          - title: Port
            type: number
            default: 28301
            required: false
            description: 绑定端口
            descriptionEn: Bind Port
            properties:
          - title: AdvertiseIPV4
            type: string
            default: "127.0.0.1"
            required: false
            description: 广播IPv4地址
            descriptionEn: Advertise IPv4 Address
            properties:
          - title: AdvertiseIPV6
            type: string
            default: "::1"
            required: false
            description: 广播IPv6地址
            descriptionEn: Advertise IPv6 Address
            properties:
          - title: AuthIdentity
            type: string
            default: none
            required: false
            description: 认证身份
            descriptionEn: Authentication Identity
            properties:
      - title: CallbackServer
        type: object
        default:
        required: false
        description: 回调服务器信息
        descriptionEn: Callback Server Information
        properties:
          - title: BindIP
            type: string
            default: "127.0.0.1"
            required: false
            description: 绑定IP地址
            descriptionEn: Bind IP Address
            properties:
          - title: Port
            type: number
            default: 28302
            required: false
            description: 绑定端口
            descriptionEn: Bind Port
            properties:
          - title: AdvertiseIPV4
            type: string
            default: "127.0.0.1"
            required: false
            description: 广播IPv4地址
            descriptionEn: Advertise IPv4 Address
            properties:
          - title: AdvertiseIPV6
            type: string
            default: "::1"
            required: false
            description: 广播IPv6地址
            descriptionEn: Advertise IPv6 Address
            properties:
          - title: AuthIdentity
            type: string
            default: none
            required: false
            description: 认证身份
            descriptionEn: Authentication Identity
            properties:
      - title: DownloadServer
        type: object
        default:
        required: false
        description: 下载服务器信息
        descriptionEn: Download Server Information
        properties:
          - title: BindIP
            type: string
            default: "127.0.0.1"
            required: false
            description: 绑定IP地址
            descriptionEn: Bind IP Address
            properties:
          - title: Port
            type: number
            default: 28303
            required: false
            description: 绑定端口
            descriptionEn: Bind Port
            properties:
          - title: AdvertiseIPV4
            type: string
            default: "127.0.0.1"
            required: false
            description: 广播IPv4地址
            descriptionEn: Advertise IPv4 Address
            properties:
          - title: AdvertiseIPV6
            type: string
            default: "::1"
            required: false
            description: 广播IPv6地址
            descriptionEn: Advertise IPv6 Address
            properties:
          - title: AuthIdentity
            type: string
            default: none
            required: false
            description: 认证身份
            descriptionEn: Authentication Identity
            properties:
      - title: WorkspaceFileGroupFullPath
        type: string
        default: /tmp/bknm/
        required: false
        description: 文件分发工作空间路径
        descriptionEn: File Distribution Workspace Path
        properties:
control:
  start: "./start.sh bk-nodemgr-relay"
  stop: "./stop.sh bk-nodemgr-relay"
  restart: "./restart.sh bk-nodemgr-relay"
  reload: "./reload.sh bk-nodemgr-relay"
  version: "./bk-nodemgr-relay -v"
```

> **注意**
> 配置模板必须提供且仅提供一份主配置文件模板，模板必须为yaml格式
> control中必须提供start、stop、restart、reload和version命令，其他命令可选

- configTemplates：插件配置模板列表，描述插件支持的配置模板信息。
  - name：配置文件名称
  - filePath：配置模板文件存放路径
  - isMainConfig：是否为主配置文件, 支持 true（是）、 false（否），只能有一个主配置文件
  - variables：提供给页面配置的变量列表
    - title：变量名称
    - type：变量类型，支持 string（字符串）、 number（数字）、 bool（布尔值）、 object（对象）、 array（数组）
    - default：变量默认值
    - required：变量是否必填，支持 true（是）、 false（否）
    - description：变量描述信息
    - descriptionEn：变量英文描述信息
    - properties: 该变量的子配置，仅支持type为 object 和 array 配置，结构与variables一致，array的子配置仅允许存在一个，用于描述数组元素的类型，即数组内元素均为同一类型

- control：插件控制信息，描述插件的启动、停止、重启、热加载和版本查询命令。
  - start：插件启动命令
  - stop：插件停止命令
  - restart：插件重启命令
  - reload：插件热加载命令
  - version：插件版本查询命令
  - health：插件健康检查命令
  - kill： 插件强制停止命令
