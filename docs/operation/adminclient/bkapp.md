# bkapp：输出蓝鲸应用凭证

`bkapp` 命令组按凭证名输出指定消费者的蓝鲸应用 appCode/appSecret，供安装/升级类 Job（bkiam sync、apigw sync）获取应用凭证。凭证来源由 adminclient 自行解析，调用方无需感知 KMS 是否开启：

| 凭证名 | KMS 模式（`kms.enabled: true`） | 明文模式（`kms.enabled: false`） |
| ------ | ------------------------------ | -------------------------------- |
| `iamv3` | KMS global envelope 的 `bkapp_id_secret.default` | Backend 配置 `iamV3.appCode/appSecret` |
| `iamv4` | 同上 | Backend 配置 `iamV4.appCode/appSecret` |
| `apigw-sync` | 同上 | 进程环境变量 `BK_APP_CODE`/`BK_APP_SECRET`（由 Job 的 `extraEnvVars` 注入） |

```bash
bk-nodemgr-adminclient bkapp code <name> [-f config] [--conf-dir dir]
bk-nodemgr-adminclient bkapp secret <name> [-f config] [--conf-dir dir]
```

输出为纯值（stdout 仅一行），可直接被脚本捕获：

```bash
APP_CODE=$(bk-nodemgr-adminclient bkapp code iamv3 -f /bk-nodemgr/etc/backend_conf.yaml)
APP_SECRET=$(bk-nodemgr-adminclient bkapp secret iamv3 -f /bk-nodemgr/etc/backend_conf.yaml)
```

## 前提

- `-f/--file`、`--conf-dir` 语义与 [database](database.md) 命令一致；凭证均位于 Backend 配置，自动发现匹配 `<conf-dir>` 下含 `callbackServer:` 段的 yaml/yml 文件（与文件名无关，helm 的 `backend_conf.yaml` 与 legacy `bk-nodemgr-backend.yml` 均覆盖）；多个文件命中时报错，需 `-f` 显式指定。
- KMS 模式：运行环境需可读 `kms.globalSecretFile` 指向的 global envelope 文件，并注入 `BK_NODEMGR_KMS_PRIVATE_KEY_GLOBAL` 环境变量（K8S 中即 `*-global-kms-secret` 的 `privateKey` key）。只需 global envelope，无需 module envelope。
- 明文模式：`iamv3`/`iamv4` 要求配置文件中对应段的 `appCode/appSecret` 已渲染为明文值；`apigw-sync` 要求环境已注入 `BK_APP_CODE`/`BK_APP_SECRET`。
- 解析结果为空时报错退出，不会输出空值；KMS 模式下 envelope 缺失或解密失败同样明确报错，不回退明文。

## 命令合同说明

`code`/`secret` 必须携带凭证名（`bkapp code` 裸调用为参数错误）。凭证名是消费者的凭证选择语义：不同消费者在明文部署中可使用独立的应用凭证（如 IAM V3/V4 各自的 appCode/appSecret），KMS 模式下统一映射到 envelope 的 default 凭证。
