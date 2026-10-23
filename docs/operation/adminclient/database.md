# 存储检查与客户端代理命令

`bk-nodemgr-adminclient` 是独立的管理 CLI，在镜像中位于 `/bk-nodemgr/bin/bk-nodemgr-adminclient`。本页记录 `database`（别名 `db`）下的存储相关子命令：`database check`、`database etcd`、`database redis`、`database mongodb`。

所有子命令共用同一条配置加载路径：读取服务配置文件中的 `etcd` / `redis` / `mongodb` 段，当配置中 `kms.enabled: true` 时，存储凭据（密码、MongoDB 用户名、etcd TLS 私钥密码）由 KMS global envelope 覆盖，配置文件中可以不写明文凭据。

## 配置文件来源

| 输入 | 语义 |
| ---- | ---- |
| `-f, --file` | 显式指定服务配置文件路径（backend/application/file 的 `*_conf.yaml` 均可） |
| `--conf-dir` | 自动发现目录，缺省 `/bk-nodemgr/etc`；未指定 `-f` 时，在该目录下发现第一个包含目标段的 `*_conf.yaml` |

注意：自动发现只匹配 `*_conf.yaml`；docker-compose 部署的配置文件为 `bk-nodemgr-*.yml`，不在匹配范围内，需用 `-f` 显式指定。

KMS 环境下还需满足：

| 前提 | 说明 |
| ---- | ---- |
| global envelope 文件 | 路径由配置 `kms.globalSecretFile` 指定，当前运行环境必须可读 |
| `BK_NODEMGR_KMS_PRIVATE_KEY_GLOBAL` | global envelope 的解密私钥，通过环境变量注入 |

在 K8S 部署中，Pod 内已挂载 envelope 并注入私钥环境变量，可直接 `kubectl exec` 进容器执行；Pod 外手工运行时需自行准备上述两项。只校验/代理命令用到的目标存储段，配置文件中其他段的渲染占位符（如 `__ADVERTISE_IPV4__`）不影响执行。

## check：存储连通性检查

单次探测：成功退出 0，失败退出非 0，不在进程内重试。重试由调用方负责（helm initContainer 以 1 秒间隔循环调用本命令直至成功）。未知组件名（如拼错的 `database check mongo`）会报错并以非 0 退出。

```bash
bk-nodemgr-adminclient database check etcd [-f config]
bk-nodemgr-adminclient database check redis [-f config]
bk-nodemgr-adminclient database check mongodb [-f config]
```

语义：

- `etcd`：任一 endpoint 响应 `Status` 即视为成功（容忍集群部分成员不可用）。
- `redis`：按配置 `type`（standalone/cluster/sentinel）构造对应客户端执行 `PING`。
- `mongodb`：使用配置凭据对 primary 执行 `ping`。

Helm 部署中 backend/application/file 的存储检查 initContainer 使用本命令。

## etcd：etcdctl 代理

```bash
bk-nodemgr-adminclient database etcd [-f config] -- <etcdctl-subcommand> [args...]
```

`--` 之后的参数原样透传给 `etcdctl`；连接参数（`--endpoints`、`--cacert/--cert/--key`）按配置自动注入，证书文件仅在存在时注入。配置中的凭据通过 `ETCDCTL_USER` 环境变量传递，不出现在进程参数中。adminclient 不回显完整命令行，避免泄露透传参数中的敏感信息。

```bash
bk-nodemgr-adminclient database etcd -- member list
bk-nodemgr-adminclient database etcd -- endpoint health
```

## redis：redis-cli 代理

```bash
bk-nodemgr-adminclient database redis [-f config] [login|status]
```

| 子命令 | 语义 |
| ------ | ---- |
| `login`（默认） | 打开交互式 redis-cli 会话 |
| `status` | `PING` 并输出简要服务器信息（cluster 模式输出 `CLUSTER INFO`） |

密码通过 `REDISCLI_AUTH` 环境变量传递，不出现在进程参数中。sentinel 模式先经首个 sentinel 解析 master 地址再连接；cluster 模式自动加 `-c`。

## mongodb：mongo shell 代理

```bash
bk-nodemgr-adminclient database mongodb [-f config] [login|status]
```

`mongo` 保留为别名，两种写法等价。

| 子命令 | 语义 |
| ------ | ---- |
| `login`（默认） | 打开交互式 mongo shell 会话 |
| `status` | 输出 `serverStatus` 摘要与副本集状态 |

连接 URI 由配置组装：主机缺省端口 27017，凭据按百分号编码（空格编为 `%20`），`authSource`/`authMechanism` 透传为查询参数。

限制：mongosh 不支持以环境变量传递凭据，URI 中的凭据会出现在进程参数（`ps`、`/proc/<pid>/cmdline`）中。命令行本身不会回显到日志；对进程参数可见性有要求的场景，应在受控的运维环境（如集群内容器）中执行。

## 运维脚本

`script_tools/ops/` 下的 `etcd.sh`、`redis.sh`、`mongo.sh` 是上述子命令的薄封装，保持原有用法（`-c` 指定配置、缺省自动发现），供习惯脚本入口的运维场景使用：

```bash
./etcd.sh endpoint health      # 等价于 adminclient database etcd -- endpoint health
./redis.sh status              # 等价于 adminclient database redis status
./mongo.sh login               # 等价于 adminclient database mongodb login
```
