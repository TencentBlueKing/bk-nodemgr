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

| 配置方式                       | 监听结果              |
|----------------------------|-------------------|
| 只配置 `bindIP`               | 只监听 IPv4。         |
| 只配置 `bindIPV6`             | 只监听 IPv6。         |
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
