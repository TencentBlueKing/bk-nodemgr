### Description

- API Version: v3.0.1+.
- Required Permission: none.
- Function: Get the RSA public key used to encrypt sensitive data.

### URL

POST /api/v3/cipher/rsa/get_public_key

### Request Parameters

This API uses an empty request body.

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| None | - | - | Send an empty JSON object `{}` |

### Request Example

```json
{}
```

### Response Example

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

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, `0` means success |
| message | string | Response message |
| request_id | string | Request ID |
| error | object | Error information, usually empty on success |
| permission | object | Permission application information. This API does not require IAM authorization, and no permission application contract is defined here |
| data | object | Response data |

#### error

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| system | string | Source system of the error |
| message | string | Error message |
| details | object array | Error detail list |

#### error.details[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | string | Error code |
| message | string | Error detail |

#### data

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| public_key | string | RSA public key string |

### Notes

- The API contract is defined by `proto/backend/api/v3/cipher.proto` and `docs/api/swagger/backend/api/v3/cipher.swagger.json`.
- The route is `POST /api/v3/cipher/rsa/get_public_key`. The request body is defined as an empty message, so sending `{}` is sufficient.
- Per the documented requirement, this endpoint explicitly does not require IAM authorization.
