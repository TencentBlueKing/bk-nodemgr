### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: config_policy_history_view (View Config Policy History).
- Function: Query config policy event list, with pagination, conditional filtering, and operation time range filtering support.

### URL

POST /api/v3/policy/config/event/list

### Request Parameters

| Parameter                | Type   | Required | Description                                                  |
|--------------------------|--------|----------|--------------------------------------------------------------|
| page                     | object | No       | Pagination configuration                                     |
| only_count               | bool   | No       | Whether to return only the total count, without item details |
| exact_include_conditions | object | No       | Exact match include conditions                               |
| fuzzy_include_conditions | object | No       | Fuzzy match include conditions                               |
| operate_time_range       | object | No       | Operation time range                                         |

#### page

| Parameter | Type  | Required | Description                                        |
|-----------|-------|----------|----------------------------------------------------|
| offset    | int32 | No       | Pagination start position, starting from `0`       |
| limit     | int32 | No       | Records per page, maximum `1000` at protocol layer |

#### exact_include_conditions

| Parameter         | Type         | Required | Description                                                                                                     |
|-------------------|--------------|----------|-----------------------------------------------------------------------------------------------------------------|
| configpolicy_id   | int64 array  | No       | Config policy ID list                                                                                           |
| configpolicy_type | string array | No       | Config policy type list, available values: `config_policy_agent`, `config_policy_proxy`, `config_policy_plugin` |
| version           | int64 array  | No       | Config policy version list                                                                                      |
| type              | string array | No       | Event type list, available values: `create`, `update`, `delete`, `enable`, `disable`, `reorder_priorities`      |
| bk_biz_id         | int64 array  | No       | Business ID list                                                                                                |

#### fuzzy_include_conditions

| Parameter         | Type         | Required | Description                          |
|-------------------|--------------|----------|--------------------------------------|
| configpolicy_name | string array | No       | Config policy name list, fuzzy match |
| operator          | string array | No       | Operator list, fuzzy match           |

#### operate_time_range

| Parameter           | Type  | Required | Description                |
|---------------------|-------|----------|----------------------------|
| start_timestamp_sec | int64 | No       | Start timestamp in seconds |
| end_timestamp_sec   | int64 | No       | End timestamp in seconds   |

### Request Example

Query create and update events for business `2`.

```json
{
  "page": {
    "offset": 0,
    "limit": 20
  },
  "only_count": false,
  "exact_include_conditions": {
    "bk_biz_id": [
      2
    ],
    "type": [
      "create",
      "update"
    ]
  },
  "fuzzy_include_conditions": {
    "operator": [
      "admin"
    ]
  },
  "operate_time_range": {
    "start_timestamp_sec": 1744588800,
    "end_timestamp_sec": 1744675200
  }
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123465",
  "data": {
    "total": 2,
    "items": [
      {
        "tenant_id": "default",
        "configpolicy_id": 10001,
        "configpolicy_name": "prod-agent-config",
        "type": "create",
        "configpolicy_type": "config_policy_agent",
        "version": 1,
        "operate_time": 1744670000000,
        "operator": "admin",
        "bk_biz_id": 2
      },
      {
        "tenant_id": "default",
        "configpolicy_id": 10001,
        "configpolicy_name": "prod-agent-config-v2",
        "type": "update",
        "configpolicy_type": "config_policy_agent",
        "version": 2,
        "operate_time": 1744671200000,
        "operator": "admin",
        "bk_biz_id": 2
      }
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

| Parameter | Type  | Description                                      |
|-----------|-------|--------------------------------------------------|
| total     | int64 | Total number of records matching current filters |
| items     | array | Config policy event list                         |

#### data.items[n]

| Parameter         | Type   | Description                                                                                                |
|-------------------|--------|------------------------------------------------------------------------------------------------------------|
| tenant_id         | string | Tenant ID                                                                                                  |
| configpolicy_id   | int64  | Config policy ID                                                                                           |
| configpolicy_name | string | Config policy name                                                                                         |
| type              | string | Event type, available values: `create`, `update`, `delete`, `enable`, `disable`, `reorder_priorities`      |
| configpolicy_type | string | Config policy type, available values: `config_policy_agent`, `config_policy_proxy`, `config_policy_plugin` |
| version           | int64  | Config policy version                                                                                      |
| operate_time      | int64  | Operation time, Unix timestamp in milliseconds                                                             |
| operator          | string | Operator                                                                                                   |
| bk_biz_id         | int64  | Business ID                                                                                                |
