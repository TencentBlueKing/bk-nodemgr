# 前端凭据加密传输契约（供前端实现方对接）

本文档定义前端（浏览器）加密主机登录凭据（密码/密钥文件内容）并提交给 bk-nodemgr
后端的协议。**全局只启用一种加密套件**（后端 `cryptoType` 配置：`CLASSIC` 或
`SHANGMI`），算法选择由服务端决定，调用方通过统一接口获取当前公钥并按 `key_type`
适配，不感知算法切换。

## 1. 获取当前公钥

```http
POST /api/v3/cipher/get_public_key
Content-Type: application/json

{}
```

响应：

```json
{
  "code": 0,
  "data": {
    "key_type": "SM2",
    "public_key": "-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----"
  }
}
```

- `key_type` 标识当前启用的非对称算法：`RSA4096`（CLASSIC 套件）或 `SM2`（SHANGMI 套件）
- 公钥为 PKIX PEM 编码；后端首次调用时自动生成密钥对并落库（`cipher` 集合，`name=DEFAULT`，`key_type` 与返回值一致）
- 旧接口 `POST /api/v3/cipher/rsa/get_public_key` 继续可用（仅返回 RSA 公钥），仅供已发布客户端过渡使用，后续将下线；新实现一律使用统一接口

## 2. 密文格式

提交到 `login_password` / `login_key_file` 字段的值：

- `key_type=SM2` 时：

```
base64( 0x03 ‖ SM2密文 )
```

  - 首字节为版本号 `0x03`（`SM2Version = 3`）
  - SM2 密文编码**两种均可**，后端自动识别：
    - **ASN.1 DER**（推荐，jsrsasign / BouncyCastle 风格）：`SEQUENCE { INTEGER C1x, INTEGER C1y, OCTET STRING C3(SM3), OCTET STRING C2 }`
    - **裸 C1C3C2**：`04 ‖ X(32B) ‖ Y(32B) ‖ C3(32B) ‖ C2`
  - C1 为未压缩点（`04` 前缀）；C3 为 SM3 摘要；C2 为密文

- `key_type=RSA4096` 时：沿用既有 RSA 链路格式（版本字节 `0x01`/`0x02`，OAEP / 混合加密），本文档不重复定义

## 3. 算法行为说明

- **全局单算法**：后端按 `cryptoType` 只启用一种套件，加密与解密均受当前算法约束。
  不属于当前启用套件的密文（版本字节不匹配）会被**显式拒绝**，后端不会按密文版本
  自动切换到另一种算法
- SM2 加密**无明文长度限制**，不需要 RSA-OAEP 的超长明文混合加密（V2）回退逻辑
- 后端解密失败（错误公钥/篡改/版本不匹配）直接报错；SM2 依赖 SM3 摘要提供完整性校验
- 前端统一加密模块负责按 `key_type` 识别并适配当前算法与密文协议，业务页面只使用
  统一的初始化、加密入口，不维护两套算法选择和调用流程

## 4. 切换加密套件的影响

**切换 `cryptoType`（如 CLASSIC → SHANGMI）后，切换前落库的临时凭据全部失效**，
不提供跨套件的数据迁移：

- 数据库中存储的登录密码/密钥文件为**临时数据，均带过期时间**（Agent 安装凭证默认
  24 小时过期，Proxy 安装凭证按调用参数 `creditExpiredIntervalSec` 过期）
- 切换后，使用旧套件加密的存量凭证解密将被拒绝，相关任务报错属预期行为
- **恢复方式：重新发起安装并在安装时重新输入凭据即可**，新凭据按当前套件加密存储
- 切换前请确认存量临时凭证已过期或不再使用；生产环境建议在业务低峰期操作

## 5. 参考实现（Go 侧等价逻辑）

```go
// 加密（等价于前端要做的事，SM2 场景）：
ct, _ := sm2.EncryptASN1(rand.Reader, pubKey, plaintext) // ASN.1 C1C3C2
payload := base64.StdEncoding.EncodeToString(append([]byte{0x03}, ct...))

// 后端解密自动识别 ASN.1 与裸 C1C3C2 两种格式。
```

前端实现：`front/src/common/crypto.ts`（sm-crypto，统一 `encrypt()` 入口，按
`key_type` 自动适配算法）。前端密文采用裸 C1C3C2 格式：
`base64(0x03 ‖ 04 ‖ C1x ‖ C1y ‖ C3 ‖ C2)`——sm-crypto `doEncrypt(msg, pubHex, 1)`
输出 C1C3C2 拼接（C1 不含 04 前缀，公钥需含 04 前缀），前置 `0x03 0x04` 两个字节
即得。该实现已与后端 `NewSM2CrypterFromPrivateKey` 完成互通验证（含中文、超长
密码、密钥文件场景）。

## 6. 交叉验证结论

后端 SM2 实现已与本机 BabaSSL/Tongsuo（OpenSSL 国密分支）完成双向交叉验证：
openssl 加密 → Go 解密、Go 加密（ASN.1）→ openssl 解密，均逐字节一致。
