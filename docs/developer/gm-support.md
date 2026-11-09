# 国密（GM）支持说明与设计

本文档介绍 bk-nodemgr 的国密（SM2/SM4/TLCP）支持范围、设计决策与实现原理。前端凭据加密的
字段级协议（密文格式、接口契约）见 [credential-encryption-contract.md](./credential-encryption-contract.md)，
本文不重复。

## 1. 支持范围

| 场景 | 经典套件（CLASSIC） | 国密套件（SHANGMI） | 说明 |
| --- | --- | --- | --- |
| 前端凭据加密传输（非对称） | RSA-4096（OAEP / 混合加密） | SM2 | 密码/密钥文件内容的端到端加密 |
| 凭据落库加密（对称） | AES-CBC | SM4-GCM | 主机凭据存 Mongo 前的加密 |
| 出站传输加密（TLS） | 标准 TLS | TLCP（GB/T 38636 双证书） | 访问 apigw 类 HTTPS 端点，配置驱动 |
| apigw JWT 验签 | RSA 验签 | 不支持 | apigw 为外部网关，算法选择权不在 nodemgr |

明确排除（有意不覆盖）：

- **apigw JWT 验签**：apigw 用 RSA 签发，nodemgr 只验签，无法也不应替换为国密算法
- **内部服务间 JWT**（`authIdentity: rest-server` 的 HS256）：自签自验的对称签名，不在国密改造范围
- **服务端 TLCP 监听**：TLCP 仅实现客户端；nodemgr 服务的入向流量由 ingress/apigw 终止 TLS
- **基础设施客户端 TLS**（etcd/redis/mongodb）：非国密改造对象

## 2. 核心设计：全局加密套件开关

国密能力由部署级配置 `cryptoType` 统一切换（backend 配置 / helm `backend.config.cryptoType` /
docker-compose `BK_NODEMGR_BACKEND_CRYPTO_TYPE`）：

- `CLASSIC`（默认）：RSA-4096 + AES-CBC，Go 标准库实现
- `SHANGMI`：SM2 + SM4-GCM，纯 Go 实现（`github.com/emmansun/gmsm`）

设计要点：

1. **全局只启用一套**。另一套件的密文在加密、解密两个方向都被拒绝，不存在混合状态。
2. **不做跨套件数据迁移**。切换套件后，切换前落库的临时凭据（带过期时间的登录密码/密钥文件）
   全部失效，需在安装时重新输入凭据。这是刻意决策：临时凭据生命周期短，迁移复杂度不值。
3. **算法选择由服务端决定**。前端/调用方通过统一接口获取当前公钥，按响应的 `key_type` 适配，
   不感知套件切换。

## 3. 场景实现原理

### 3.1 前端凭据加密（非对称）

```
浏览器                          application/backend              Mongo
  │  POST /api/v3/cipher/get_public_key  │                        │
  │ ───────────────────────────────────► │  首次调用时按套件生成    │
  │                                      │  默认密钥对并落库 ─────► │ cipher 集合
  │  {key_type, public_key}              │  (name=DEFAULT)         │ (name=DEFAULT)
  │ ◄─────────────────────────────────── │                        │
  │  按 key_type 选 RSA/SM2 加密密码       │                        │
  │  提交 login_password/login_key_file  │                        │
  │ ───────────────────────────────────► │  私钥解密（版本字节分发）  │
```

- **密钥对管理**：`internal/backend/storage/cipher`，启动/首次使用时幂等生成（`ensureDefaultCipher`），
  只生成启用套件的那一对（CLASSIC→RSA-4096，SHANGMI→SM2）
- **公钥下发**：`GetCurrentPublicKey`（`POST /api/v3/cipher/get_public_key`）返回
  `key_type`（`RSA4096`/`SM2`）+ PEM 公钥，application 与 backend 双侧注册。
  旧接口 `GetRSAPublicKey` 保留并固定返回 RSA 公钥，仅供已发布客户端过渡
- **密文格式**：首字节为版本号，解密按版本分发，SM2 与 RSA 共享同一密文空间：
  - `v1`：RSA-OAEP
  - `v2`：RSA 混合加密（RSA 包裹 AES 密钥 + AES-GCM 加密正文，解决 RSA 明文长度限制）
  - `v3`：SM2
- **后端解密**：安装流程（agent/proxy install）中由 `initCredentialCrypter` 按当前套件
  构造 RSA/SM2 私钥 crypter，经 `DecryptBase64Ciphertext` 统一入口解密
- **前端实现**：`front/src/common/crypto.ts` 统一加密入口，SM2 使用 `sm-crypto` 库；
  所有凭据输入页面（agent 安装/导入、proxy 安装/重装/编辑）均走该入口

### 3.2 凭据落库加密（对称）

- `internal/backend/service` 启动时按 `cryptoType` 构造全局对称 crypter
  （`newSymmetricCrypter`）：CLASSIC→AES-CBC，SHANGMI→SM4-GCM
- 密钥来自部署配置 `encryptKey`，经固定 salt 派生
- 消费方为 `internal/backend/storage/credit`（主机凭据存取），接口不变，底层算法可替换

### 3.3 出站传输加密（TLCP 客户端）

所有第三方系统（GSE/CMDB/UserManager 等）均经 apigw 访问，TLCP 能力在 HTTP 客户端层
统一实现，对**所有** apigw 类客户端配置平等生效：

- **触发条件**：某客户端的 `tls` 配置同时设置 `encCertFile` + `encKeyFile`（即 `GMEnabled()`），
  则该客户端的 HTTPS 请求改用 TLCP 拨号；否则维持标准 TLS，零开销闲置
- **双证书**：TLCP（GB/T 38636）要求双证书对——`certFile`/`keyFile` 为签名对，
  `encCertFile`/`encKeyFile` 为加密对；`pkg/config` 的 `TLSConfig.Validate` 校验配对完整性
  （加密对存在则签名对必须存在）
- **实现位置**：
  - `pkg/runtime/ssl/tlcp.go`：TLCP client config 构建（`NewClientTLCPConf`）、
    证书加载（签名/加密对、加密 PKCS#8 私钥、SM CA）、`ResolveTLCPDialConf`
    （按拨号地址补 ServerName——gotlcp 不会自动校验 hostname，必须显式填充）
  - `pkg/rest/client/http_client.go`：`http.Transport.DialTLSContext` 注入 TLCP 拨号器
  - `pkg/relayhandler/server.go`：backend→GSE 消息下发的独立 HTTP 客户端，复用同一 TLSConfig
- **协议栈**：`gitee.com/Trisia/gotlcp`

注意：TLCP 仅客户端。

## 4. 配置方式

### 4.1 加密套件

```yaml
# helm values.yaml
backend:
  config:
    cryptoType: "SHANGMI"   # 或 CLASSIC（默认）
    encryptKey: "<对称密钥>"  # 两个套件共用，切换套件建议一并更换
```

docker-compose：`BK_NODEMGR_BACKEND_CRYPTO_TYPE="SHANGMI"`。

### 4.2 TLCP 客户端证书

需要 TLCP 时，运维自行将双证书挂载进 Pod，并在对应客户端配置引用：

```yaml
backend:
  config:
    gse:   # 任意 APIGatewayClient 配置段均可，不限于 gse
      tls:
        certFile: "/path/to/sign.crt"      # 签名对
        keyFile: "/path/to/sign.key"
        encCertFile: "/path/to/enc.crt"    # 加密对，设置后即启用 TLCP
        encKeyFile: "/path/to/enc.key"
        caFile: "/path/to/ca.crt"          # 可选，SM CA
        password: ""                        # 可选，加密 PKCS#8 私钥口令
```

## 5. 依赖与代码地图

新增依赖：

| 依赖 | 用途 |
| --- | --- |
| `github.com/emmansun/gmsm` | SM2/SM4/smx509/pkcs8，纯 Go |
| `gitee.com/Trisia/gotlcp` | TLCP（GB/T 38636）协议栈 |
| `sm-crypto`（前端） | 浏览器侧 SM2 |

关键代码：

| 位置 | 职责 |
| --- | --- |
| `pkg/runtime/crypter` | RSA/SM2/AES/SM4 crypter 与密文版本分发，见包内 AGENTS.md |
| `pkg/runtime/ssl` | 标准 TLS / TLCP client config 构建，见包内 AGENTS.md |
| `pkg/config` | `cryptoType`、`TLSConfig`（含 TLCP 配对校验） |
| `internal/backend/storage/cipher` | 非对称密钥对生成与存储 |
| `internal/backend/storage/credit` | 凭据对称加密落库 |
| `internal/{backend,application}/router/api-v3/cipher` | 公钥下发接口 |
| `pkg/rest/client/http_client.go` | TLCP 拨号器注入 |
| `front/src/common/crypto.ts` | 前端统一加密入口 |

## 6. 兼容性与升级说明

- 默认 `CLASSIC`，存量部署无感知
- 旧公钥接口 `GetRSAPublicKey` 保留，已发布客户端不受影响
- 切换到 `SHANGMI`：未过期的新增凭据全部用 SM2/SM4；**切换前已落库的临时凭据失效**，
  需重新录入（无迁移，见 credential-encryption-contract.md）
- TLCP 为逐客户端 opt-in，不配置 `encCertFile/encKeyFile` 则行为与之前完全一致
