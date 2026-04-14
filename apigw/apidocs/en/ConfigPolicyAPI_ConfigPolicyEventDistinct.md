### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: config_policy_history_view (View Config Policy History).
- Function: Query distinct field sets of config policy events by filtering conditions.

### URL

POST /api/v3/policy/config/event/distinct

### Request Parameters

| Parameter                | Type   | Required | Description                    |
|--------------------------|--------|----------|--------------------------------|
| exact_include_conditions | object | No       | Exact match include conditions |
| fuzzy_include_conditions | object | No       | Fuzzy match include conditions |
| operate_time_range       | object | No       | Operation time range           |

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

```json
{
  "exact_include_conditions": {
    "bk_biz_id": [
      2
    ],
    "type": [
      "create",
      "update",
      "enable"
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
  "request_id": "req-123466",
  "data": {
    "configpolicy_id": [
      10001,
      10002
    ],
    "configpolicy_name": [
      "prod-agent-config",
      "prod-proxy-config"
    ],
    "configpolicy_type": [
      "config_policy_agent",
      "config_policy_proxy"
    ],
    "type": [
      "create",
      "update",
      "enable"
    ],
    "version": [
      1,
      2,
      3
    ],
    "operator": [
      "admin",
      "ops"
    ],
    "bk_biz_id": [
      2
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
| data       | object | Distinct result                             |

#### data

| Parameter         | Type         | Description                                                                                                                     |
|-------------------|--------------|---------------------------------------------------------------------------------------------------------------------------------|
| configpolicy_id   | int64 array  | Distinct results of config policy IDs                                                                                           |
| configpolicy_name | string array | Distinct results of config policy names                                                                                         |
| configpolicy_type | string array | Distinct results of config policy types, available values: `config_policy_agent`, `config_policy_proxy`, `config_policy_plugin` |
| type              | string array | Distinct results of event types, available values: `create`, `update`, `delete`, `enable`, `disable`, `reorder_priorities`      |
| version           | int64 array  | Distinct results of config policy versions                                                                                      |
| operator          | string array | Distinct results of operators                                                                                                   |
| bk_biz_id         | int64 array  | Distinct results of business IDs                                                                                                |
