# token

## 设计意图

本包用于生成紧凑、URL 安全且带过期时间的无状态 token。token 使用 HMAC-SHA256 进行完整性校验，调用方无需保存 token 状态即可识别内容是否被篡改以及是否已经过期。

本包接收任意 `[]byte` 作为载荷，但不解释载荷的业务含义。业务层负责将 export ID 等标识转换为字节，并在解析 token 后还原标识、查询数据库记录。

## 功能边界

1. 此包负责：
    - 将版本、过期时间和载荷编码为 URL 安全的 token。
    - 使用 HMAC-SHA256 对 token 进行签名和验签。
    - 校验 token 的格式、版本、签名和过期时间。
2. 此包不负责：
    - export ID、UUID 等业务标识的编解码。
    - 数据库查询、下载鉴权和一次性下载控制。
    - 密钥的读取、保存、轮换和吊销。
    - 对 token 载荷进行加密。

## 实现说明

### Token 布局

token 的二进制布局如下：

`version(1 byte) || expiresAt(4 bytes) || payload(N bytes) || HMAC tag(12 bytes)`

说明：

- `version`：格式版本，当前为 `1`。
- `expiresAt`：大端序的无符号 32 位 Unix 秒级时间戳。
- `payload`：调用方传入的非空字节数据。
- `HMAC tag`：HMAC-SHA256 结果的前 12 字节，即 96 bit。
- HMAC 覆盖版本、过期时间和载荷，并使用包内固定的 v1 签名域进行隔离。
- 完整二进制数据使用无填充的 `base64.RawURLEncoding` 编码，可直接放入 URL 查询参数或路径参数。

### Token 长度

编码前长度为 `17 + len(payload)` 字节，编码后长度由 `base64.RawURLEncoding.EncodedLen` 决定。

当 export ID 使用 UUID 的 16 字节原始形式时：

- 编码前长度：`1 + 4 + 16 + 12 = 33` 字节。
- 编码后长度：44 个 URL 安全字符。

版本、字段顺序、时间戳宽度、签名域、HMAC 长度和 Base64 编码方式都属于兼容性协议。修改这些内容时必须使用新的格式版本，并协调 token 生成端和解析端同步升级。

### 过期时间

- `New` 默认生成有效期为 24 小时的 token。
- `WithTTL` 可覆盖默认有效期，但只接受大于零的 `time.Duration`。
- 过期时间以秒为精度写入 token。
- `Parse` 先完成验签，再校验过期时间，避免信任未经认证的数据。

### 签名密钥

backend 和 file 会在服务启动时分别初始化 token 生成器。部分本地或测试环境没有配置 JWT 对称密钥，因此包内维护固定的默认 key，使两个进程仍能使用相同密钥完成 token 的生成和校验。

密钥选择顺序如下：

1. `New` 接收到非空 key 时，使用调用方传入的 key。
2. `New` 接收到空 key 时，使用包内默认值 `bk-nodemgr/runtime/token/default-key`。

backend 和 file 复用同一份 JWT 对称密钥。生产部署配置 JWT key 后会自然覆盖默认 key，默认 key 只在调用方未传入 key 时生效。

## 使用示例

```go
package example

import (
	"errors"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/token"
)

func issueAndParse(key, exportID []byte) error {
	generator, err := token.New(key, token.WithTTL(24*time.Hour))
	if err != nil {
		return err
	}

	value, expiresAt, err := generator.Generate(exportID)
	if err != nil {
		return err
	}
	_, _ = value, expiresAt

	payload, err := generator.Parse(value)
	if errors.Is(err, token.ErrExpired) {
		return err
	}
	if errors.Is(err, token.ErrInvalid) {
		return err
	}
	if err != nil {
		return err
	}
	_ = payload

	return nil
}
```

export 业务层应在调用 `Generate` 前将 UUID 字符串转换为 16 字节原始值，在 `Parse` 成功后再将返回的字节恢复为 UUID，并使用该 UUID 查询 export 记录。

## 错误约定

- `ErrInvalid`：token 编码错误、长度错误、版本不支持或签名不匹配。
- `ErrExpired`：token 已通过签名校验，但已经到达过期时间。
- 调用方应使用 `errors.Is` 判断上述错误。

## 使用限制

1. token 只提供完整性和有效期校验，不提供保密性。任何获得 token 的人都可以通过 Base64 解码读取过期时间和载荷，因此载荷中不能放置密码、密钥或其他敏感信息。
2. 公网下载必须使用 HTTPS。token 属于 bearer token，获得下载链接的人在 token 过期前都可以重复使用该链接。
3. 需要互通的生成端和解析端必须使用相同密钥。生产环境应传入已配置的 JWT 对称密钥；包内默认 key 用于未配置密钥的环境。
4. 下载接口、网关和日志系统不应记录完整 token，避免有效下载凭证通过访问日志或错误日志泄漏。
5. 本包不提供单次使用和主动吊销能力。如需限制下载次数或提前失效，应由业务层结合 export 记录状态实现。
