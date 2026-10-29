# jwt_token：签发服务 JWT token

`jwt_token` 从 KMS global envelope 解密指定服务 HTTP server 的 JWT 对称密钥，并直接签发 token。stdout 只输出 token，不输出密钥，供安装/升级脚本调用需要认证的内部 API 时使用（如初始化包导入）。

```bash
bk-nodemgr-adminclient jwt_token <service> <server> [-f config] [--conf-dir dir] [--login-name admin] [--bk-username admin]
```

支持的 service/server 组合：

| service     | server                                   |
| ----------- | ---------------------------------------- |
| application | info、admin、basic                        |
| backend     | info、admin、basic、callback、proxy       |
| file        | info、admin、basic、download、export      |

示例（为 File basic server 签发）：

```bash
JWT_TOKEN=$(bk-nodemgr-adminclient jwt_token file basic -f /bk-nodemgr/etc/file_conf.yaml)
```

输出为单行 token，可直接被脚本捕获；命令替换会去掉末尾换行。

## 前提

- 配置文件中 `kms.enabled: true`；kms 未开启时命令报错退出（非 KMS 部署的明文路径不在此命令范围）。
- `-f/--file`、`--conf-dir` 语义与 [database](database.md) 命令一致；自动发现按 `kms:` 段匹配。
- 运行环境需可读 `kms.globalSecretFile` 指向的 global envelope 文件，并注入 `BK_NODEMGR_KMS_PRIVATE_KEY_GLOBAL` 环境变量（K8S 中即 `*-global-kms-secret` 的 `privateKey` key）。只需 global envelope，无需 module envelope。
- `--login-name`、`--bk-username` 默认均为 `admin`，会写入 JWT 声明用于标识调用方。
- token 使用 HS256 签名，有效期固定 24 小时；签名密钥为空时报错退出，不会输出空 token。
