# Server 配置 FAQ

## 1. 如何开启 IPv4 + IPv6 双监听？

以 basicServer 为例，同时配置 `bindIP` 和 `bindIPV6`。

```yaml
basicServer:
  bindIP: 0.0.0.0
  bindIPV6: "::"
  port: 28100
```

`bindIP` 只控制 IPv4 listener，`bindIPV6` 只控制 IPv6 listener。默认不会启用 dual-stack；如果只配置其中一个，就只监听对应协议栈。

| 配置方式                        | 监听结果                |
| ------------------------------- | ----------------------- |
| 只配置 `bindIP`                 | 只监听 IPv4。           |
| 只配置 `bindIPV6`               | 只监听 IPv6。           |
| 同时配置 `bindIP` 和 `bindIPV6` | 同时监听 IPv4 和 IPv6。 |

检查监听状态：

```bash
ss -nultp | grep <process-name>
```

如果 `bindIPV6: "::1"` 启动失败，先确认系统启用了 IPv6 loopback：

```bash
ip -6 addr show lo
sysctl net.ipv6.conf.lo.disable_ipv6 net.ipv6.conf.all.disable_ipv6
```

## 2. 如何修改全局默认的 proxy filecache 目录？

这个问题对应的是 backend 配置里的 `gseDeployConfs[*].custom.proxyFileCacheDir`。

它的结构位置是：

```yaml
gseDeployConfs:
  - generation: 2
    osType: linux
    custom:
      proxyFileCacheDir: "/data/bknm-proxy-cache/"
```

### 什么时候这个字段会生效？

当前代码里，这个字段只在 **proxy 节点** 场景下参与 fallback 注入：

- 节点角色必须是 `proxy`
- 没有在安装的请求里提供自定义的对应配置
- 当前配置策略里还没有显式覆盖 `fileCacheDirs`

如果这些条件同时满足，backend 会把 `proxyFileCacheDir` 注入到节点部署配置里。

### 一个可直接参考的 demo

```yaml
gseDeployConfs:
  - generation: 2
    osType: linux
    baseWorkDir: "/tmp/bknm/"
    baseDeployDir: "/usr/local/"
    manualScriptPath: "/bk-nodemgr/script/manual/linux/install.sh"
    custom:
      proxyFileCacheDir: "/data/bknm-proxy-cache/"
```

- 你要改的是 backend 的配置项中 `gseDeployConfs[*].custom.proxyFileCacheDir`
- 由于只有 linux 才会存在 proxy , 所以只需要配置 linux 的 `gseDeployConfs` 条目即可
- 如果你的配置策略已经显式下发了 `fileCacheDirs`，这里的 fallback 不会覆盖它

## 3. 如何修改插件所需的 hostid 路径？

这个问题对应的是 backend 配置里的 `gseDeployConfs[*].pluginCustom.hostIDPath`。

它的结构位置是：

```yaml
gseDeployConfs:
  - generation: 2
    osType: linux
    pluginCustom:
      hostIDPath: "/var/lib/bknm_custom/host/hostid"
```

### 一个可直接参考的 demo

```yaml
gseDeployConfs:
  - generation: 2
    osType: linux
    baseWorkDir: "/tmp/bknm/"
    baseDeployDir: "/usr/local/"
    manualScriptPath: "/bk-nodemgr/script/manual/linux/install.sh"
    pluginCustom:
      hostIDPath: "/var/lib/bknm_custom/host/hostid"
```

在此处没有提供默认值的情况下，backend 会在渲染插件配置文件时根据下述规则生成一个默认值：

- Linux / Unix 风格默认值：`/var/lib/<env>/host/hostid`
- Windows 风格默认值：`C:\<env>\data\host\hostid`

这里的 `<env>` 来自 `pkg/system.GetEnv()`。

## 4. 如何修改兼容性安装（插件 V2）的运行目录和数据目录？

这个问题对应的是 backend 配置里的 `gseDeployConfs[*].pluginCustom.v2RunDir` 和 `gseDeployConfs[*].pluginCustom.v2DataDir`。

它的结构位置是：

```yaml
gseDeployConfs:
  - generation: 2
    osType: linux
    pluginCustom:
      v2RunDir: "/data/bknm-plugin-v2/run"
      v2DataDir: "/data/bknm-plugin-v2/data"
```

### 什么时候生效？

只有插件开启兼容性安装时才会生效。插件安装请求命中兼容性策略（或节点安装携带兼容性安装选项）后，部署参数会带上 `EnableCompatibilityMode=true`，backend 会额外触发 `ensure_plugin_v2` 流程，只有该流程在注入插件 V2 基础运行时时读取这两个字段。

普通的「安装插件」流程（插件 V3）不读取它们，目录仍按 `<baseDeployDir>/<env>/plugin/<plugin_group>/<plugin_name>/run` 和 `.../data` 生成。

### 配置的是最终目录，不是父目录

留空时，兼容性安装使用 `<baseDeployDir>/<env>/plugins/run` 和 `.../data`（Windows 为 `<baseDeployDir>\<env>\plugins\run` 和 `...\data`）。

填入非空值后：

- 该值作为最终目录直接使用，不会再追加 `<env>`、`plugins` 等层级
- 不再受 `baseDeployDir`、`PluginRuntime.BaseDeployDir` 等部署目录配置影响

因此同一 `generation` + `osType` 下所有兼容性安装会共享同一个 `v2RunDir` 和 `v2DataDir`。`v2RunDir` 中的 pid 文件按插件包名区分（`<plugin_pkg_name>.pid`），不会直接冲突；`v2DataDir` 则完全共享，需要确认插件不会写入同名文件，或迁移目标本身就是统一目录。

### 优先级

这两个字段没有网络单元级别的覆盖项，优先级只有两级：

| 配置项      | 优先级                                                                              |
| ----------- | ----------------------------------------------------------------------------------- |
| `v2RunDir`  | 非空的 `pluginCustom.v2RunDir` ＞ `<baseDeployDir>/<env>/plugins/run`               |
| `v2DataDir` | 非空的 `pluginCustom.v2DataDir` ＞ `<baseDeployDir>/<env>/plugins/data`             |
| `logDir`    | 网络单元 `PluginRuntime.LogDir` ＞ 非空的 `pluginCustom.logDir` ＞ 内置默认日志目录 |

`logDir` 的适用范围与上面两个字段不同：兼容性安装的日志目录取自节点部署配置，配置 `pluginCustom.logDir` 不会改变兼容性安装的日志路径。

### 一个可直接参考的 demo

```yaml
gseDeployConfs:
  - generation: 2
    osType: linux
    baseWorkDir: "/tmp/bknm/"
    baseDeployDir: "/usr/local/"
    manualScriptPath: "/bk-nodemgr/script/manual/linux/install.sh"
    pluginCustom:
      v2RunDir: "/data/bknm-plugin-v2/run"
      v2DataDir: "/data/bknm-plugin-v2/data"
```

配置本身不会让插件进入兼容性安装，是否安装插件 V2 仍由兼容性策略决定。改动只影响之后执行的兼容性安装流程；已部署插件记录的目录不会自动迁移。

## 5. 如何做到项目无损升级？

这个问题主要对应 Helm 的 `updateStrategy` 和 Pod 的 `terminationGracePeriodSeconds` 设置。

`updateStrategy` 的结构位置是：

```yaml
updateStrategy:
  type: RollingUpdate
  rollingUpdate:
    maxUnavailable: 1
    maxSurge: 1
```

### 什么时候这个配置会生效？

当前 Helm 模板会把 `.Values.updateStrategy` 渲染到 `backend`、`application` 和 `file` 三类 Server 的 Deployment `strategy` 中。

当 `updateStrategy.type` 为 `RollingUpdate` 时，Kubernetes 会在升级时滚动替换 Pod：

- 先创建新的 Pod，并按 `maxSurge` 控制额外可创建的 Pod 数量
- 再停止旧的 Pod，并按 `maxUnavailable` 控制升级期间允许不可用的 Pod 数量

### 还需要配合什么？

`RollingUpdate` 只负责控制 Pod 的替换顺序。要做到无损升级，还需要给旧 Pod 留出足够的退出时间：

```yaml
terminationGracePeriodSeconds: 120
```

这个值可以使用全局配置，也可以在 `backend`、`application`、`file` 各模块下单独覆盖。它需要和 [Server Graceful Shutdown](../concepts/server/graceful_shutdown.md) 配合使用，确保旧 Pod 收到退出信号后不再接收新流量，并在宽限期内处理完已有请求或任务。
