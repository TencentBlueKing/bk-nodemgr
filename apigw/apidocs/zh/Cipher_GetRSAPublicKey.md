### 描述

- 该接口提供版本：v3.0.1-alpha.1+。
- 该接口所需权限：无。
- 该接口功能描述：获取默认 RSA4096 公钥，用于对敏感信息进行加密。

### URL

POST /api/v3/cipher/rsa/get_public_key

### 输入参数

该接口没有业务输入参数，但必须发送空 JSON 请求体。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| 无 | - | - | 发送空 JSON 对象 `{}` |

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
| permission | object | 权限申请信息；该接口未定义接口级 IAM action，成功时通常为空 |
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
| public_key | string | PEM 格式的默认 RSA4096 公钥 |

### 说明

- 接口契约来源于 `proto/backend/api/v3/cipher.proto` 和 `docs/api/swagger/backend/api/v3/cipher.swagger.json`。
- 请求消息没有字段，但 handler 会绑定 JSON 请求体，因此调用时必须发送 `{}`。
- 当前 handler 不执行接口级 IAM action 校验；API Gateway 仍要求应用认证和资源访问权限。
