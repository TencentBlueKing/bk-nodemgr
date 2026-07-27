### Description

- API Version: v3.0.1-alpha.60+.
- Required Permission: networkunit_view (View Network Unit).
- Function: List network units with pagination and condition filters.

### URL

POST /api/v3/topo/networkunit/list

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| page | object | No | Pagination configuration |
| only_count | bool | No | Return only the total count without details |
| exact_include_conditions | object | No | Exact include filter conditions |

#### page

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| offset | int32 | No | Offset, starting from 0 |
| limit | int32 | No | Page size limit |

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
  "page": {
    "offset": 1,
    "limit": 1
  },
  "only_count": false,
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
    "total": 1,
    "items": [
      {
        "tenant_id": "id-001",
        "bk_networkunit_id": 1,
        "bk_networkunit_name": "default",
        "bk_networkarea_id": 1,
        "accesspoints": [
          1
        ],
        "links": {
          "cluster": {
            "bk_networkarea_id": 1,
            "bk_networkunit_id": 1,
            "accesspoint_id": 1
          },
          "file": {
            "bk_networkarea_id": 1,
            "bk_networkunit_id": 1,
            "accesspoint_id": 1
          },
          "data": {
            "bk_networkarea_id": 1,
            "bk_networkunit_id": 1,
            "accesspoint_id": 1
          }
        },
        "is_direct": false,
        "direct_endpoints": {
          "cluster": [
            "string"
          ],
          "file": [
            "string"
          ],
          "data": [
            "string"
          ]
        }
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
| total | int64 | Total number of records |
| items | object array | Data list |

#### data.items[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| tenant_id | string | Tenant ID |
| bk_networkunit_id | int64 | Network unit ID |
| bk_networkunit_name | string | Network unit name |
| bk_networkarea_id | int64 | Network area ID |
| accesspoints | int64 array | Access point list |
| links | object | Link configuration |
| is_direct | bool | Whether it is directly connected |
| direct_endpoints | object | Direct endpoint configuration |
| generation | int64 | Generation |
| custom_deploy_config | object | Custom deployment configuration |
