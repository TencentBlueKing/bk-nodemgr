### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: None.
- Function: Get distinct candidate values of node workflow fields under filter conditions, including workflow type, business ID, operator, and
  workflow status.

### URL

POST /api/v3/node/workflow/distinct

### Request Parameters

| Parameter                | Type   | Required | Description                                               |
|--------------------------|--------|----------|-----------------------------------------------------------|
| exact_include_conditions | object | No       | Exact match include conditions                            |
| fuzzy_include_conditions | object | No       | Fuzzy match include conditions, currently an empty object |

#### exact_include_conditions

| Parameter          | Type         | Required | Description                                                                                                                                                                                                          |
|--------------------|--------------|----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| workflow_id        | string array | No       | Workflow ID list                                                                                                                                                                                                     |
| type               | string array | No       | Workflow type list, available values: `install_agent`, `install_proxy`, `upgrade_agent`, `upgrade_proxy`, `reconfig_agent`, `reconfig_proxy`, `restart_agent`, `restart_proxy`, `uninstall_agent`, `uninstall_proxy` |
| bk_biz_id          | int64 array  | No       | Business ID list                                                                                                                                                                                                     |
| status             | string array | No       | Workflow status list, available values: `running`, `success`, `failed`, `partial_failed`                                                                                                                             |
| operator           | string array | No       | Operator list                                                                                                                                                                                                        |
| bk_host_innerip    | string array | No       | Host inner IPv4 list                                                                                                                                                                                                 |
| bk_host_innerip_v6 | string array | No       | Host inner IPv6 list                                                                                                                                                                                                 |
| node_role          | string array | No       | Node role list, available values: `blank`, `agent`, `proxy`                                                                                                                                                          |

#### fuzzy_include_conditions

The object is currently empty, pass `{}` or omit it.

### Request Example

Query distinct workflow candidates under business `2` with type `upgrade_agent` and status `running`.

```json
{
  "exact_include_conditions": {
    "bk_biz_id": [
      2
    ],
    "type": [
      "upgrade_agent"
    ],
    "status": [
      "running"
    ],
    "node_role": [
      "agent"
    ]
  },
  "fuzzy_include_conditions": {}
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "type": [
      "install_agent",
      "upgrade_agent"
    ],
    "bk_biz_id": [
      2,
      3
    ],
    "operator": [
      "admin",
      "ops"
    ],
    "status": [
      "running",
      "success",
      "failed"
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

| Parameter | Type         | Description                                                                                                                                                                                                                     |
|-----------|--------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| type      | string array | Distinct workflow type values, available values: `install_agent`, `install_proxy`, `upgrade_agent`, `upgrade_proxy`, `reconfig_agent`, `reconfig_proxy`, `restart_agent`, `restart_proxy`, `uninstall_agent`, `uninstall_proxy` |
| bk_biz_id | int64 array  | Distinct business ID values                                                                                                                                                                                                     |
| operator  | string array | Distinct operator values                                                                                                                                                                                                        |
| status    | string array | Distinct workflow status values, available values: `running`, `success`, `failed`, `partial_failed`                                                                                                                             |

### Notes

- The current implementation always returns four distinct field groups, `type`, `bk_biz_id`, `operator`, and `status`, and does not support selecting
  return columns in request.
- `fuzzy_include_conditions` is currently an empty object in protocol, and the converter does not read any fuzzy condition content.
- The current handler `DistinctNodeWorkflow` does not apply permission narrowing for business visibility, unlike `ListNodeWorkflow`.
