### Description

- API Version: v3.0.1-alpha.60+.
- Required Permission: networkarea_view (View Network Area).
- Function: Recommend available network units by host network segment.

### URL

POST /api/v3/topo/networkunit/recommend_by_network_segment

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| items | object array | No | Data list |

#### items[n]

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_networkarea_id | int64 | No | Network area ID |
| ip | string | No | ip |

### Request Example

```json
{
  "items": [
    {
      "bk_networkarea_id": 1,
      "ip": "10.0.0.1"
    }
  ]
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
    "items": [
      {
        "bk_networkarea_id": 1,
        "ip": "10.0.0.1",
        "bk_networkunit_id": 1,
        "message": "string"
      }
    ]
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

#### data

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| items | object array | Data list |

#### data.items[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| bk_networkarea_id | int64 | Network area ID |
| ip | string | ip |
| bk_networkunit_id | int64 | Network unit ID |
| message | string | Request message |
