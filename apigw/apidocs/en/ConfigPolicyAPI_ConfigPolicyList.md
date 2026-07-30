### Description

- API Version: v3.0.1-alpha.63+.
- Required Permission: config_policy_view (View Config Policy).
- Function: Query the config policy list, with pagination, exact filtering, and fuzzy filtering support.

### URL

POST /api/v3/policy/config/list

### Request Parameters

| Parameter                | Type   | Required | Description                                                  |
| ------------------------ | ------ | -------- | ------------------------------------------------------------ |
| page                     | object | No       | Pagination configuration                                     |
| only_count               | bool   | No       | Whether to return only the total count, without item details |
| exact_include_conditions | object | No       | Exact match include conditions                               |
| fuzzy_include_conditions | object | No       | Fuzzy match include conditions                               |

#### page

| Parameter | Type  | Required | Description                                        |
| --------- | ----- | -------- | -------------------------------------------------- |
| offset    | int32 | No       | Pagination start position, starting from `0`       |
| limit     | int32 | No       | Records per page, maximum `1000` at protocol layer |

#### exact_include_conditions

| Parameter          | Type         | Required | Description                                                                                                     |
| ------------------ | ------------ | -------- | --------------------------------------------------------------------------------------------------------------- |
| configpolicy_id    | int64 array  | No       | Config policy ID list                                                                                           |
| bk_biz_id          | int64 array  | No       | Business ID list                                                                                                |
| configpolicy_type  | string array | No       | Config policy type list, available values: `config_policy_agent`, `config_policy_proxy`, `config_policy_plugin` |
| enabled            | bool array   | No       | Enabled status list                                                                                             |
| target_plugin_name | string array | No       | Target plugin name list                                                                                         |

#### fuzzy_include_conditions

| Parameter         | Type         | Required | Description                          |
| ----------------- | ------------ | -------- | ------------------------------------ |
| configpolicy_name | string array | No       | Config policy name list, fuzzy match |
| operator          | string array | No       | Operator list, fuzzy match           |

### Request Example

Query enabled plugin config policies under business `2`.

```json
{
  "page": {
    "offset": 0,
    "limit": 20
  },
  "only_count": false,
  "exact_include_conditions": {
    "bk_biz_id": [2],
    "configpolicy_type": ["config_policy_plugin"],
    "enabled": [true],
    "target_plugin_name": ["bkmonitorbeat"]
  },
  "fuzzy_include_conditions": {
    "configpolicy_name": ["prod"]
  }
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 1,
    "items": [
      {
        "tenant_id": "default",
        "configpolicy_id": 10001,
        "configpolicy_name": "prod-plugin-config",
        "configpolicy_type": "config_policy_plugin",
        "bk_biz_id": 2,
        "remark": "Production plugin configuration",
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
          "plugin.base.cpu_percent_limit": 10,
          "plugin.base.mem_percent_limit": 10
        },
        "configs_bool": {
          "enable_metrics": true
        },
        "enabled": true,
        "updated_time": 1744675200000,
        "operator": "admin",
        "version": 3,
        "priority": 1,
        "target_host_ids": [1001, 1002],
        "target_plugin_name": "bkmonitorbeat"
      }
    ]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                 |
| ---------- | ------ | ------------------------------------------- |
| code       | int32  | Status code, `0` means success              |
| message    | string | Request message                             |
| request_id | string | Request ID                                  |
| error      | object | Error information, usually empty on success |
| permission | object | Permission information                      |
| data       | object | Response data                               |

#### data

| Parameter | Type  | Description                                      |
| --------- | ----- | ------------------------------------------------ |
| total     | int64 | Total number of records matching current filters |
| items     | array | Config policy list                               |

#### data.items[n]

| Parameter          | Type        | Description                                                                                                |
| ------------------ | ----------- | ---------------------------------------------------------------------------------------------------------- |
| tenant_id          | string      | Tenant ID                                                                                                  |
| configpolicy_id    | int64       | Config policy ID                                                                                           |
| configpolicy_name  | string      | Config policy name                                                                                         |
| configpolicy_type  | string      | Config policy type, available values: `config_policy_agent`, `config_policy_proxy`, `config_policy_plugin` |
| bk_biz_id          | int64       | Business ID                                                                                                |
| remark             | string      | Remark                                                                                                     |
| scopes             | array       | Policy effective scope list                                                                                |
| configs_string     | object      | String config key-value pairs                                                                              |
| configs_int        | object      | Integer config key-value pairs                                                                             |
| configs_bool       | object      | Boolean config key-value pairs                                                                             |
| enabled            | bool        | Whether enabled                                                                                            |
| updated_time       | int64       | Update time, Unix timestamp in milliseconds                                                                |
| operator           | string      | Operator                                                                                                   |
| version            | int64       | Config policy version                                                                                      |
| priority           | int64       | Priority, smaller value means higher priority                                                              |
| target_host_ids    | int64 array | Target host ID list                                                                                        |
| target_plugin_name | string      | Target plugin name                                                                                         |

#### data.items[n].scopes[n]

| Parameter         | Type   | Description                              |
| ----------------- | ------ | ---------------------------------------- |
| bk_networkarea_id | int64  | Network area ID, `-1` means any          |
| bk_networkunit_id | int64  | Network unit ID, `-1` means any          |
| os_type           | string | OS type, empty string means any          |
| cpu_arch          | string | CPU architecture, empty string means any |
