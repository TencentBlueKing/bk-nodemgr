### Description

- API Version: v3.0.1-alpha.60+.
- Required Permission: networkunit_create (Create Network Unit).
- Function: Create a network unit with access points, links, and deployment configuration.

### URL

POST /api/v3/topo/networkunit/create

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_networkunit_name | string | Yes | Network unit name |
| bk_networkarea_id | int64 | Yes | Network area ID |
| accesspoints | object array | No | Access point list |
| links | object | No | Link configuration |
| is_direct | bool | No | Whether it is directly connected |
| direct_endpoints | object | No | Direct endpoint configuration |
| generation | int64 | Yes | Generation |
| custom_deploy_config | object | No | Custom deployment configuration |

#### accesspoints[n]

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| tenant_id | string | No | Tenant ID |
| accesspoint_id | int64 | No | Access point ID |
| accesspoint_name | string | No | Access point name |
| bk_networkarea_id | int64 | No | Network area ID |
| endpoints | object | No | Endpoint configuration |

#### accesspoints[n].endpoints

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| cluster | string array | No | Cluster channel configuration |
| file | string array | No | File channel configuration |
| data | string array | No | Data channel configuration |

#### links

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| cluster | object | No | Cluster channel configuration |
| file | object | No | File channel configuration |
| data | object | No | Data channel configuration |

#### links.cluster

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_networkarea_id | int64 | No | Network area ID |
| bk_networkunit_id | int64 | No | Network unit ID |
| accesspoint_id | int64 | No | Access point ID |

#### links.file

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_networkarea_id | int64 | No | Network area ID |
| bk_networkunit_id | int64 | No | Network unit ID |
| accesspoint_id | int64 | No | Access point ID |

#### links.data

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_networkarea_id | int64 | No | Network area ID |
| bk_networkunit_id | int64 | No | Network unit ID |
| accesspoint_id | int64 | No | Access point ID |

#### direct_endpoints

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| cluster | string array | No | Cluster channel configuration |
| file | string array | No | File channel configuration |
| data | string array | No | Data channel configuration |

#### custom_deploy_config.{key}

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| installer_runtime | object | No | Installer runtime configuration |
| node_runtime | object | No | Node runtime configuration |
| plugin_runtime | object | No | Plugin runtime configuration |

#### custom_deploy_config.{key}.installer_runtime

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| base_work_dir | string | No | Base work directory |

#### custom_deploy_config.{key}.node_runtime

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| base_deploy_dir | string | No | Base deployment directory |
| data_ipc | string | No | Data process IPC path |
| plugin_ipc | string | No | Plugin process IPC path |
| log_dir | string | No | Log directory |
| zone_id | string | No | Zone ID |
| city_id | string | No | City ID |

#### custom_deploy_config.{key}.plugin_runtime

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| base_deploy_dir | string | No | Base deployment directory |
| log_dir | string | No | Log directory |

### Request Example

```json
{
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
  },
  "generation": 1,
  "custom_deploy_config": {
    "default": {
      "installer_runtime": {
        "base_work_dir": "/var/lib/gse"
      },
      "node_runtime": {
        "base_deploy_dir": "/var/lib/gse",
        "data_ipc": "/var/run/ipc.sock",
        "plugin_ipc": "/var/run/ipc.sock",
        "log_dir": "/var/lib/gse",
        "zone_id": "id-001",
        "city_id": "id-001"
      },
      "plugin_runtime": {
        "base_deploy_dir": "/var/lib/gse",
        "log_dir": "/var/lib/gse"
      }
    }
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
