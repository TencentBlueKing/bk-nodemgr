### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：无。
- 该接口功能描述：获取 RSA 公钥，用于对敏感信息进行加密。

### URL

POST /api/v3/cipher/rsa/get_public_key

### 输入参数

该接口请求体为空。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| 无 | - | - | 发送空 JSON 对象 `{}` 即可 |

### 调用示例

```json
{}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "error": null,
  "permission": null,
  "data": {
    "public_key": "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA...\n-----END PUBLIC KEY-----"
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，`0` 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时通常为空 |
| permission | object | 权限申请信息。该接口无需 IAM 鉴权，文档不定义权限申请要求 |
| data | object | 响应数据 |

#### error

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| system | string | 错误来源系统 |
| message | string | 错误消息 |
| details | object array | 错误详情列表 |

#### error.details[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | string | 错误码 |
| message | string | 错误详情 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| public_key | string | RSA 公钥字符串 |

### 说明

- 接口契约来源于 `proto/backend/api/v3/cipher.proto` 和 `docs/api/swagger/backend/api/v3/cipher.swagger.json`。
- 路由为 `POST /api/v3/cipher/rsa/get_public_key`，请求体定义为空消息，因此调用时发送 `{}` 即可。
- 根据需求约束，该接口明确不要求 IAM 鉴权。
