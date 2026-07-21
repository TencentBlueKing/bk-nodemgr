### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: config_policy_manage (Manage Config Policy).
- Function: Reorder priorities of enabled policies under the same business and the same policy type in the specified order.

### URL

POST /api/v3/policy/config/reorder_priorities

### Request Parameters

| Parameter               | Type        | Required | Description                                                                                                |
|-------------------------|-------------|----------|------------------------------------------------------------------------------------------------------------|
| bk_biz_id               | int64       | Yes      | Business ID                                                                                                |
| configpolicy_type       | string      | Yes      | Config policy type, available values: `config_policy_agent`, `config_policy_proxy`, `config_policy_plugin` |
| ordered_configpolicy_id | int64 array | No       | Expected ordered policy ID list, duplicate IDs are not allowed                                             |

### Request Example

```json
{
  "bk_biz_id": 2,
  "configpolicy_type": "config_policy_agent",
  "ordered_configpolicy_id": [
    10003,
    10001,
    10002
  ]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123463",
  "data": {}
}
```

### Response Parameters

| Parameter  | Type   | Description                                 |
|------------|--------|---------------------------------------------|
| code       | int32  | Status code, `0` means success              |
| message    | string | Request message                             |
| request_id | string | Request ID                                  |
| error      | object | Error information, usually empty on success |
| permission | object | Permission information                      |
| data       | object | Response data, empty object on success      |

### Notes

- Protocol semantics: policies listed in `ordered_configpolicy_id` are assigned priorities `1..N`.
- Enabled policies in the same scope that are not listed keep their relative order and continue from `N+1`.
