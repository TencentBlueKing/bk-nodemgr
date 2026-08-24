# Config Policy FAQ

## 1. 如何通过 Config Policy 修改 Agent / Proxy 日志存储路径？

这个问题对应的是 `support-files/configpolicy/` 下 Agent 配置策略里的 `logger_path` 选项，以及 Proxy 配置策略里的 `all_logger_path` 选项。

### Agent 配置策略 —— `logger_path`

- 文件位置：`support-files/configpolicy/agent_option.json`
- 选项 id：`logger_path`
- 模板 key：`agent.logger.path`
- 含义：控制 Agent 进程日志文件的输出目录。仅影响 Agent 自身（不覆盖 Proxy，也不影响 plugin 进程日志路径）。

它的结构位置是：

```json
{
  "enabled": false,
  "id": "logger_path",
  "name_en": "Log path",
  "name_zh": "日志存储路径",
  "key": "agent.logger.path",
  "type": 0,
  "value_string": ""
}
```

- 默认状态：模板中 `enabled: false`，默认不下发；启用并填入非空路径后才会随配置策略一起下发。
- 路径格式：建议使用绝对路径；Windows 节点用反斜杠或正斜杠都可以，建议与 GSE Agent 安装路径保持同级，例如 `C:\gse\logs\agent`。
- 安全提示：路径所在的目录必须对 Agent 进程可写；若指向已有目录，GSE Agent 不会清理历史日志，需要自行做日志轮转。
- 生效预期：配置策略下发后，下一次 Agent 上报心跳或触发 Reconfig 时由 installer 重新渲染配置；不需要重启主机。

### Proxy 配置策略 —— `all_logger_path`

- 文件位置：`support-files/configpolicy/proxy_option.json`
- 选项 id：`all_logger_path`
- 模板 key：`VALUE_GROUP_PROXY_LOGGER_LEVEL`（group 类型，分配给三个子 key）
- 含义：一次性覆盖 Proxy 节点上 Agent、file、data 三类进程的日志存储目录（与 `all_logger_level` 覆盖日志等级对应）。

实际生效的子 key：

```json
{
  "enabled": false,
  "id": "all_logger_path",
  "name_en": "Log Path",
  "name_zh": "日志存储路径",
  "key": "VALUE_GROUP_PROXY_LOGGER_LEVEL",
  "type": 0,
  "value_string": "",
  "value_group_assigned": {
    "agent.logger.path": "",
    "file.logger.path": "",
    "data.logger.path": ""
  }
}
```

- `agent.logger.path` — Proxy 节点上的 Agent 进程日志路径
- `file.logger.path` — Proxy 节点上的 file server 进程日志路径
- `data.logger.path` — Proxy 节点上的 data 进程日志路径

- 默认状态：模板中 `enabled: false`，默认不下发。
- 路径格式：绝对路径；建议使用 Proxy 安装目录的同级目录，例如 `/var/log/gse-proxy`。
- 安全提示：三个子 key 共享同一个目录设置；如果需要分别控制三个进程路径，目前需要拆成多条 Proxy 策略或后续拆分 group。
- 生效预期：与 Agent 配置策略一致，下一次 Reconfig 时由 installer 渲染；不需要重启主机。

### 与已有 process / plugin log path 字段的关系

- `agent.logger.path` / `file.logger.path` / `data.logger.path` 控制对应进程的日志输出目录（已通过 Config Policy 暴露）。
- plugin 进程的日志路径目前由 GSE Plugin 自身模板决定，不在本次 Config Policy 范围内。运维如需修改 plugin 日志路径，需要通过 plugin 包模板或环境变量调整。
- 配置策略只下发 `value_string`；空字符串 = 不下发，保持 Agent / Proxy 自身默认值。
