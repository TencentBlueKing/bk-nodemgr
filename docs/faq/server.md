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
