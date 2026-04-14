### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: config_policy_manage (Manage Config Policy).
- Function: Create a new config policy.

### URL

POST /api/v3/policy/config/create

### Request Parameters

| Parameter         | Type        | Required | Description                                                                                                |
|-------------------|-------------|----------|------------------------------------------------------------------------------------------------------------|
| configpolicy_name | string      | Yes      | Config policy name                                                                                         |
| configpolicy_type | string      | Yes      | Config policy type, available values: `config_policy_agent`, `config_policy_proxy`, `config_policy_plugin` |
| bk_biz_id         | int64       | Yes      | Business ID                                                                                                |
| remark            | string      | No       | Remark                                                                                                     |
| scopes            | array       | Yes      | Policy effective scope list                                                                                |
| configs_string    | object      | No       | String config key-value pairs                                                                              |
| configs_int       | object      | No       | Integer config key-value pairs                                                                             |
| configs_bool      | object      | No       | Boolean config key-value pairs                                                                             |
| operator          | string      | No       | Operator                                                                                                   |
| target_host_ids   | int64 array | No       | Target host ID list                                                                                        |

#### scopes[n]

| Parameter         | Type   | Required | Description                              |
|-------------------|--------|----------|------------------------------------------|
| bk_networkarea_id | int64  | No       | Network area ID, `-1` means any          |
| bk_networkunit_id | int64  | No       | Network unit ID, `-1` means any          |
| os_type           | string | No       | OS type, empty string means any          |
| cpu_arch          | string | No       | CPU architecture, empty string means any |

### Request Example

```json
{
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
  "operator": "admin",
  "target_host_ids": [
    1001,
    1002
  ]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123458",
  "data": {
    "configpolicy_id": 10001
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

| Parameter       | Type  | Description                    |
|-----------------|-------|--------------------------------|
| configpolicy_id | int64 | Newly created config policy ID |
