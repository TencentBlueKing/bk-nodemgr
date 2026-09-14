### Description

- API Version: v3.0.1+.
- Required Permission: None.
- Function: Get the default public key of the globally enabled credential encryption suite (RSA4096 for CLASSIC, SM2 for SHANGMI).

### URL

POST /api/v3/cipher/get_public_key

### Request Parameters

This API has no business input parameters, but an empty JSON request body is required.

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
    "key_type": "SM2",
    "public_key": "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAE...\n-----END PUBLIC KEY-----"
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, `0` means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, usually empty on success |
| permission | object | Permission apply information; this API defines no API-level IAM action, usually empty on success |
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
| key_type | string | Asymmetric algorithm of public_key: `RSA4096` (CLASSIC suite) or `SM2` (SHANGMI suite); callers adapt encryption accordingly |
| public_key | string | The default public key in PEM (PKIX) format |

### Notes

- The contract comes from `proto/backend/api/v3/cipher.proto` and `docs/api/swagger/backend/api/v3/cipher.swagger.json`.
- The request message has no fields, but the handler binds a JSON body, so `{}` must be sent.
- The handler performs no API-level IAM action check; the API Gateway still requires application authentication and resource permission.
- Only one crypto suite is enabled globally (the `cryptoType` config); the server decides the algorithm and callers adapt via `key_type`. See `docs/developer/credential-encryption-contract.md` for the ciphertext format contract.
- The legacy `/api/v3/cipher/rsa/get_public_key` endpoint remains available (always returns the RSA key) for released clients during the transition and will be retired later.
