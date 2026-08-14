### Description

- API Version: v3.0.1-alpha.67+.
- Required Permission: `networkunit_create (Create Network Unit)`, `networkunit_view (View Network Unit)`.
- Function: Batch-create same-name empty non-direct network units under empty network areas for managed-environment initialization.

### URL

POST /api/v3/topo/networkunit/create_default_multi

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_networkarea_id | int64 array | Yes | Target network area IDs; unique, 1-100 items, and must not contain 0 |
| bk_networkunit_name | string | Yes | Name used for every created network unit |
| upstream | object | Yes | Shared upstream access point |

#### upstream

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_networkarea_id | int64 | Yes | Upstream network area ID |
| bk_networkunit_id | int64 | Yes | Upstream network unit ID |
| accesspoint_id | int64 | Yes | Upstream access point ID |

**Creation rules**:

- Target areas must be ordinary network areas. The default area `bk_networkarea_id=0` is not supported.
- Each target area must be empty at execution time, with no active network units.
- Created units are ordinary non-direct units without downstream access points or custom deploy configuration.
- `upstream` is used for the cluster, file, and data channels.
- `networkunit_create` is checked per target network area.
- The upstream network unit requires `networkunit_view`.
- A single request supports at most 100 target network areas.
- Individual target areas may fail without blocking other target areas.

### Request Example

```json
{
  "bk_networkarea_id": [101, 102, 103],
  "bk_networkunit_name": "default",
  "upstream": {
    "bk_networkarea_id": 10,
    "bk_networkunit_id": 20,
    "accesspoint_id": 30
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
    "success_count": 2,
    "failed_count": 1,
    "items": [
      {
        "bk_networkarea_id": 101,
        "success": true,
        "bk_networkunit_id": 1001,
        "error_code": "",
        "message": "network unit created"
      },
      {
        "bk_networkarea_id": 102,
        "success": true,
        "bk_networkunit_id": 1002,
        "error_code": "",
        "message": "network unit created"
      },
      {
        "bk_networkarea_id": 103,
        "success": false,
        "error_code": "networkarea_not_empty",
        "message": "networkarea-id(103) is not empty"
      }
    ]
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code; 0 means the request was processed and item-level failures are returned in data.items |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Request-level error information |
| permission | object | Request-level permission information |
| data | object | Batch creation result |

#### data

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| success_count | int64 | Number of successful creations |
| failed_count | int64 | Number of failed items |
| items | object array | Result for each target area, in request order |

#### items[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| bk_networkarea_id | int64 | Target network area ID |
| success | bool | Whether creation succeeded |
| bk_networkunit_id | int64 | Created network unit ID when successful |
| error_code | string | Stable error code when creation failed |
| message | string | Result message |

### Common Error Codes

| Error Code | Description |
|------------|-------------|
| default_networkarea_not_supported | The default network area cannot be initialized in batch |
| networkarea_permission_denied | No network unit create permission for the target area |
| networkarea_not_found | The target network area does not exist |
| networkarea_not_empty | The target network area already contains a network unit |
| networkarea_busy | The target network area is being initialized; retry later |
| networkunit_create_failed | Network unit creation failed |
| upstream_permission_denied | No upstream network unit view permission; no writes are performed |
| upstream_networkunit_not_found | The upstream network unit does not exist; no writes are performed |
| upstream_networkarea_mismatch | The upstream unit does not match the upstream area; no writes are performed |
| upstream_accesspoint_not_found | The upstream access point does not belong to the upstream unit; no writes are performed |
