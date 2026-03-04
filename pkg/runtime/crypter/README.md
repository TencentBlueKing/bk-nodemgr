# crypter

## 设计意图

本包旨在提供多种加密算法的实现，方便用户在不同场景下选择合适的加密方式。通过统一的接口设计，用户可以轻松切换不同的加密算法，而无需修改大量代码。

## 功能边界

1. 此包负责：实现多种加密算法（如 RSA、AES 等），并提供统一的加密和解密接口。
2. 此包不负责：密钥管理、加密算法的安全性评估等复杂功能。

## 实现说明

本包当前提供：

### AES

- 算法：AES-CBC + PKCS7 padding
- 密钥派生：`SHA-256(key || salt)`（派生出固定 32 字节 key）
- IV：确定性 IV，取 `SHA-256(plaintext || derivedKey)` 的前 16 字节

**密文布局**

`AESVersion(1 byte) || IV(16 bytes) || ciphertext`

说明：

- `AESVersion` 当前为 `1`。
- 因为 IV 是确定性的，所以同一把 key 加密相同 plaintext 会得到相同 ciphertext。

### RSA

#### v1：RSA-OAEP（默认）

- 算法：RSA-OAEP（SHA-256）
- Label：可通过 `WithRSALabel` 配置，默认 `RSADefaultLabel`

**密文布局（v1）**

`RSAVersion(1 byte) || RSA ciphertext`

`RSAVersion` 当前为 `1`。

**明文长度限制**

RSA-OAEP 无法直接加密任意长度的数据。对 SHA-256 OAEP 来说，单次可加密的最大明文长度为：

$$
max = k - 2\cdot hLen - 2
$$

其中 `k` 是 RSA 模数的字节长度（例如 RSA-2048 为 256，RSA-4096 为 512），`hLen=32`（SHA-256 输出长度）。

#### v2：混合加密（明文超限自动回退）

当明文长度超过上述 RSA-OAEP 上限时，`(*RSA).Encrypt` 会自动回退为混合加密：

- 生成随机 32 字节 AES key（AES-256）
- 使用 AES-256-GCM 加密明文（随机 12 字节 nonce）
- 使用 RSA-OAEP(SHA-256) 包裹 AES key

**密文布局（v2）**

`RSAVersionHybrid(1 byte) || wrappedKey(RSA ciphertext, keySize bytes) || nonce(12 bytes) || GCM ciphertext`

说明：

- `RSAVersionHybrid` 当前为 `2`。
- `keySize` 为 RSA 模数的字节长度，即 `r.pub.Size()` / `r.priv.Size()`。
- RSA 的 `label` 会作为 AES-GCM 的 AAD（associated data）参与认证，用于将密文绑定到配置的 label。

AAD 说明：

- 加密和解密必须使用同一个 `label`。如果 `label` 被修改或两端不一致，AES-GCM 的认证会失败，解密会返回错误。
- 这是刻意的设计：用于将密文与特定上下文/配置绑定，避免在错误的上下文中被错误接受。
