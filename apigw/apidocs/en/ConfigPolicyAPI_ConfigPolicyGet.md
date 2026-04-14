### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: config_policy_view (View Config Policy).
- Function: Query details of a single config policy by config policy ID.

### URL

POST /api/v3/policy/config/get

### Request Parameters

| Parameter       | Type  | Required | Description      |
|-----------------|-------|----------|------------------|
| configpolicy_id | int64 | Yes      | Config policy ID |

### Request Example

```json
{
  "configpolicy_id": 10001
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123457",
  "data": {
    "tenant_id": "default",
    "configpolicy_id": 10001,
    "configpolicy_name": "prod-agent-config",
    "configpolicy_type": "config_policy_agent",
    "bk_biz_id": 2,
    "remark": "Production Agent configuration",
    "scopes": [
      {
        "bk_networkarea_id": -1,
        "bk_networkunit_id": -1,
        "os_type": "",
        "cpu_arch": ""
      }
    ],
    "configs_string": {
      "bk_cloud_id": "0"
    },
    "configs_int": {
      "heartbeat_interval": 60
    },
    "configs_bool": {
      "enable_metrics": true
    },
    "enabled": true,
    "updated_time": 1744675200000,
    "operator": "admin",
    "version": 3,
    "priority": 1,
    "target_host_ids": [
      1001,
      1002
    ]
  }
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
| data       | object | Response data                               |

#### data

| Parameter         | Type        | Description                                                                                                |
|-------------------|-------------|------------------------------------------------------------------------------------------------------------|
| tenant_id         | string      | Tenant ID                                                                                                  |
| configpolicy_id   | int64       | Config policy ID                                                                                           |
| configpolicy_name | string      | Config policy name                                                                                         |
| configpolicy_type | string      | Config policy type, available values: `config_policy_agent`, `config_policy_proxy`, `config_policy_plugin` |
| bk_biz_id         | int64       | Business ID                                                                                                |
| remark            | string      | Remark                                                                                                     |
| scopes            | array       | Policy effective scope list                                                                                |
| configs_string    | object      | String config key-value pairs                                                                              |
| configs_int       | object      | Integer config key-value pairs                                                                             |
| configs_bool      | object      | Boolean config key-value pairs                                                                             |
| enabled           | bool        | Whether enabled                                                                                            |
| updated_time      | int64       | Update time, Unix timestamp in milliseconds                                                                |
| operator          | string      | Operator                                                                                                   |
| version           | int64       | Config policy version                                                                                      |
| priority          | int64       | Priority, smaller value means higher priority                                                              |
| target_host_ids   | int64 array | Target host ID list                                                                                        |

#### data.scopes[n]

| Parameter         | Type   | Description                              |
|-------------------|--------|------------------------------------------|
| bk_networkarea_id | int64  | Network area ID, `-1` means any          |
| bk_networkunit_id | int64  | Network unit ID, `-1` means any          |
| os_type           | string | OS type, empty string means any          |
| cpu_arch          | string | CPU architecture, empty string means any |
