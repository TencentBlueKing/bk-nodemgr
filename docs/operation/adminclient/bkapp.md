# bkapp：输出蓝鲸应用凭证

`bkapp` 命令组从 KMS global envelope 解密并输出 `bkapp_id_secret.default` 的 appCode/appSecret，供安装/升级类 Job（如 bkiam sync、apigw sync）在不接触 values 明文的前提下获取应用凭证。

```bash
bk-nodemgr-adminclient bkapp code [-f config] [--conf-dir dir]
bk-nodemgr-adminclient bkapp secret [-f config] [--conf-dir dir]
```

输出为纯值（stdout 仅一行），可直接被脚本捕获：

```bash
APP_CODE=$(bk-nodemgr-adminclient bkapp code -f /bk-nodemgr/etc/backend_conf.yaml)
APP_SECRET=$(bk-nodemgr-adminclient bkapp secret -f /bk-nodemgr/etc/backend_conf.yaml)
```

## 前提

- 配置文件中 `kms.enabled: true`；kms 未开启时命令报错退出（非 KMS 部署的明文路径不在此命令范围）。
- `-f/--file`、`--conf-dir` 语义与 [database](database.md) 命令一致；自动发现按 `kms:` 段匹配。
- 运行环境需可读 `kms.globalSecretFile` 指向的 global envelope 文件，并注入 `BK_NODEMGR_KMS_PRIVATE_KEY_GLOBAL` 环境变量（K8S 中即 `*-global-kms-secret` 的 `privateKey` key）。
