### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: config_policy_view (View Config Policy).
- Function: Preview policy matching results and merged config results by business, policy type, and host list.

### URL

POST /api/v3/policy/config/preview

### Request Parameters

| Parameter   | Type   | Required | Description                                                                                                |
|-------------|--------|----------|------------------------------------------------------------------------------------------------------------|
| bk_biz_id   | int64  | Yes      | Business ID                                                                                                |
| policy_type | string | Yes      | Config policy type, available values: `config_policy_agent`, `config_policy_proxy`, `config_policy_plugin` |
| hosts       | array  | Yes      | Target host list for preview, at least one element                                                         |

#### hosts[n]

| Parameter         | Type   | Required | Description                                                         |
|-------------------|--------|----------|---------------------------------------------------------------------|
| bk_host_id        | int64  | Yes      | Host ID, must be greater than `0`                                   |
| bk_networkunit_id | int64  | No       | Network unit ID, auto-filled as `-1` by protocol layer when omitted |
| bk_networkarea_id | int64  | No       | Network area ID, auto-filled as `-1` by protocol layer when omitted |
| os_type           | string | No       | OS type                                                             |
| cpu_arch          | string | No       | CPU architecture                                                    |

### Request Example

```json
{
  "bk_biz_id": 2,
  "policy_type": "config_policy_agent",
  "hosts": [
    {
      "bk_host_id": 1001,
      "bk_networkunit_id": 2001,
      "bk_networkarea_id": 1,
      "os_type": "linux",
      "cpu_arch": "x86_64"
    },
    {
      "bk_host_id": 1002,
      "os_type": "linux",
      "cpu_arch": "x86_64"
    }
  ]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123464",
  "data": {
    "reliable_items": [
      {
        "bk_host_id": 1001,
        "matched_policies": [
          {
            "configpolicy_id": 10001,
            "configpolicy_name": "prod-agent-config",
            "priority": 1
          },
          {
            "configpolicy_id": 10003,
            "configpolicy_name": "fallback-agent-config",
            "priority": 2
          }
        ],
        "merged_config": "{\"agent\":{\"access\":{\"bk_cloud_id\":\"0\"},\"heartbeat_interval\":30,\"enable_metrics\":true}}"
      }
    ],
    "unreliable_items": [
      {
        "bk_host_id": 1002,
        "matched_policies": [],
        "merged_config": "{}"
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
| data       | object | Preview result                              |

#### data

| Parameter        | Type  | Description                                      |
|------------------|-------|--------------------------------------------------|
| reliable_items   | array | Host preview items with reliable match results   |
| unreliable_items | array | Host preview items with unreliable match results |

#### data.reliable_items[n] / data.unreliable_items[n]

| Parameter        | Type   | Description                                                     |
|------------------|--------|-----------------------------------------------------------------|
| bk_host_id       | int64  | Host ID                                                         |
| matched_policies | array  | Matched policy list, ordered by priority                        |
| merged_config    | string | Merged nested JSON config, keys sorted lexicographically        |

#### data.reliable_items[n].matched_policies[n]

| Parameter         | Type   | Description        |
|-------------------|--------|--------------------|
| configpolicy_id   | int64  | Config policy ID   |
| configpolicy_name | string | Config policy name |
| priority          | int64  | Policy priority    |
