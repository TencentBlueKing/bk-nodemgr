### Description

- API Version: v3.0.1-alpha.60+.
- Required Permission: networkunit_edit (Edit Network Unit).
- Function: Update network unit base information, access points, links, and deployment configuration.

### URL

POST /api/v3/topo/networkunit/update

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| networkunit | object | Yes | Network unit information |
| fields | object | Yes | Fields to update |

#### networkunit

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| tenant_id | string | No | Tenant ID |
| bk_networkunit_id | int64 | Yes | Network unit ID |
| bk_networkunit_name | string | No | Network unit name |
| bk_networkarea_id | int64 | Yes | Network area ID |
| accesspoints | object array | No | Access point list |
| links | object | No | Link configuration |
| is_direct | bool | No | Whether it is directly connected |
| direct_endpoints | object | No | Direct endpoint configuration |
| generation | int64 | No | Generation |
| custom_deploy_config | object | No | Custom deployment configuration |

#### networkunit.accesspoints[n]

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| tenant_id | string | No | Tenant ID |
| accesspoint_id | int64 | No | Access point ID |
| accesspoint_name | string | No | Access point name |
| bk_networkarea_id | int64 | No | Network area ID |
| endpoints | object | No | Endpoint configuration |

#### networkunit.links

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| cluster | object | No | Cluster channel configuration |
| file | object | No | File channel configuration |
| data | object | No | Data channel configuration |

#### networkunit.direct_endpoints

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| cluster | string array | No | Cluster channel configuration |
| file | string array | No | File channel configuration |
| data | string array | No | Data channel configuration |

#### networkunit.custom_deploy_config.{key}

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| installer_runtime | object | No | Installer runtime configuration |
| node_runtime | object | No | Node runtime configuration |
| plugin_runtime | object | No | Plugin runtime configuration |

#### fields

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_networkunit_name | bool | No | Network unit name |
| accesspoints | bool | No | Access point list |
| links | bool | No | Link configuration |
| direct_endpoints | bool | No | Direct endpoint configuration |
| custom_deploy_config | bool | No | Custom deployment configuration |

### Request Example

```json
{
  "networkunit": {
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
  },
  "fields": {
    "bk_networkunit_name": false,
    "accesspoints": false,
    "links": false,
    "direct_endpoints": false,
    "custom_deploy_config": false
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
    "bk_networkunit_id": 1
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
| bk_networkunit_id | int64 | Network unit ID |
