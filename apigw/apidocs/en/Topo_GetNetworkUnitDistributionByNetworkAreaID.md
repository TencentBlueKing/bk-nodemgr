### Description

- API Version: v3.0.1-alpha.60+.
- Required Permission: None.
- Function: Count network unit distribution by network area ID.

### URL

POST /api/v3/topo/networkunit/get_networkunit_distribution_by_networkarea_id

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| exact_include_conditions | object | No | Exact include filter conditions |

#### exact_include_conditions

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_networkunit_id | int64 array | No | Network unit ID |
| bk_networkarea_id | int64 array | No | Network area ID |
| is_direct | bool array | No | Whether it is directly connected |
| generation | int64 array | No | Generation |

### Request Example

```json
{
  "exact_include_conditions": {
    "bk_networkunit_id": [
      1
    ],
    "bk_networkarea_id": [
      1
    ],
    "is_direct": [
      false
    ],
    "generation": [
      1
    ]
  }
}
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
    "default": 1
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, 0 means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, empty on success |
| permission | object | Permission information |
| data | object | Response data |
