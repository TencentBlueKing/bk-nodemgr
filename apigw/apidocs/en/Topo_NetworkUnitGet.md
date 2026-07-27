### Description

- API Version: v3.0.1-alpha.60+.
- Required Permission: networkunit_view (View Network Unit).
- Function: Get network unit details by network unit ID.

### URL

POST /api/v3/topo/networkunit/get

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_networkunit_id | int64 | No | Network unit ID |

### Request Example

```json
{
  "bk_networkunit_id": 1
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
    "tenant_id": "id-001",
    "bk_networkunit_id": 1,
    "bk_networkunit_name": "default",
    "bk_networkarea_id": 1,
    "accesspoints": [
      {
        "tenant_id": "id-001",
        "accesspoint_id": 1,
        "accesspoint_name": "default",
        "bk_networkarea_id": 1,
        "endpoints": {
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
| tenant_id | string | Tenant ID |
| bk_networkunit_id | int64 | Network unit ID |
| bk_networkunit_name | string | Network unit name |
| bk_networkarea_id | int64 | Network area ID |
| accesspoints | object array | Access point list |
| links | object | Link configuration |
| is_direct | bool | Whether it is directly connected |
| direct_endpoints | object | Direct endpoint configuration |
| generation | int64 | Generation |
| custom_deploy_config | object | Custom deployment configuration |

#### data.accesspoints[n]

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| tenant_id | string | Tenant ID |
| accesspoint_id | int64 | Access point ID |
| accesspoint_name | string | Access point name |
| bk_networkarea_id | int64 | Network area ID |
| endpoints | object | Endpoint configuration |

#### data.links

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| cluster | object | Cluster channel configuration |
| file | object | File channel configuration |
| data | object | Data channel configuration |

#### data.direct_endpoints

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| cluster | string array | Cluster channel configuration |
| file | string array | File channel configuration |
| data | string array | Data channel configuration |

#### data.custom_deploy_config.{key}

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| installer_runtime | object | Installer runtime configuration |
| node_runtime | object | Node runtime configuration |
| plugin_runtime | object | Plugin runtime configuration |
