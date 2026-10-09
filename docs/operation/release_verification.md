# Release Source Archive Verification

bk-nodemgr GitHub Release 的源码包会提供独立的 GPG detached signature。验证者可以用本页的 public key 和 signature file 确认下载到的源码包没有在签名后被篡改。

## Signing Key

| Item | Value |
|------|-------|
| Key name | `BlueKing Node Manager Release` |
| Fingerprint | `98BB EB5B CA65 7B2C 73DC A7BF 7E84 FBD1 D6B1 001B` |
| Public key | [`release-signing-key.asc`](release-signing-key.asc) |

验证时必须核对完整 fingerprint。`Good signature` 只表示 signature file 和源码包匹配；只有 fingerprint 与可信渠道公布的值一致时，才能确认该 public key 属于 bk-nodemgr release signing key。

## Verify A Source Archive

把 `<tag>` 替换成要验证的 release tag，例如 `v3.0.1-alpha.96`。

```bash
curl -LO https://github.com/TencentBlueKing/bk-nodemgr/archive/refs/tags/<tag>.tar.gz
curl -LO https://github.com/TencentBlueKing/bk-nodemgr/releases/download/<tag>/bk-nodemgr-<tag>.tar.gz.asc
curl -LO https://raw.githubusercontent.com/TencentBlueKing/bk-nodemgr/master/docs/operation/release-signing-key.asc

gpg --import release-signing-key.asc
gpg --fingerprint "BlueKing Node Manager Release"
gpg --verify bk-nodemgr-<tag>.tar.gz.asc <tag>.tar.gz
```

确认 `gpg --fingerprint` 输出包含：

```text
98BB EB5B CA65 7B2C 73DC A7BF 7E84 FBD1 D6B1 001B
```

验证成功时，`gpg --verify` 会输出 `Good signature`。如果输出 `BAD signature`、源码包文件名不匹配，或者 fingerprint 与上面的值不一致，不要使用该源码包。

## Verification Scope

该 signature 验证的是 GitHub Release 源码包和对应 `.asc` 文件：

- 验证源码包在签名后没有被修改。
- 不验证 Docker image、Helm Chart、Agent package、Plugin package 或二进制构建产物。
- 不替代源码审计、依赖漏洞检查或运行时安全验证。
