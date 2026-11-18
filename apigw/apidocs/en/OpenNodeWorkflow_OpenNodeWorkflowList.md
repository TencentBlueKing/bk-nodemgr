### Description

- API Version: v3.0.1-alpha.114+.
- Required Permissions: system_access (System Access), agent_history_view (View Agent History), proxy_history_view (View Proxy History). History permissions are determined by the rules below.
- Function: Query node workflows and execution status as an authenticated user, with pagination, exact filters, and an operation time range.
- The API Gateway authenticates and forwards the user identity. Do not supply username or tenant parameters in the request body.

### URL

POST /api/v3/open/node/workflow/list

### Request Parameters

| Parameter                | Type   | Required | Description                                      |
| ------------------------ | ------ | -------- | ------------------------------------------------ |
| page                     | object | No       | Pagination; required when `only_count=false`     |
| only_count               | bool   | No       | Return only the total count; defaults to `false` |
| exact_include_conditions | object | No       | Exact-match include conditions                   |
| fuzzy_include_conditions | object | No       | No supported fields; omit or pass `{}`           |
| operate_time_range       | object | No       | Operation time range in Unix seconds             |

#### page

| Parameter | Type  | Required | Description                                            |
| --------- | ----- | -------- | ------------------------------------------------------ |
| offset    | int32 | No       | Starting offset, defaults to `0`; must be non-negative |
| limit     | int32 | Yes      | Records per page, from `1` to `500`                    |

#### exact_include_conditions

| Parameter          | Type         | Required | Description                                                                  |
| ------------------ | ------------ | -------- | ---------------------------------------------------------------------------- |
| workflow_id        | string array | No       | Workflow IDs; filter known IDs here, not at the top level                    |
| type               | string array | No       | Workflow types, such as `install_agent`, `upgrade_agent`, or `restart_proxy` |
| bk_biz_id          | int64 array  | No       | Business IDs                                                                 |
| status             | string array | No       | Workflow states: `running`, `success`, `failed`, `partial_failed`            |
| operator           | string array | No       | Historical operators; not the identity credentials for this request          |
| bk_host_innerip    | string array | No       | Host inner IPv4 addresses                                                    |
| bk_host_innerip_v6 | string array | No       | Host inner IPv6 addresses                                                    |
| node_role          | string array | No       | Node roles: `blank`, `agent`, `proxy`                                        |
| deploy_policy_id   | int64 array  | No       | Deployment policy IDs                                                        |

#### operate_time_range

| Parameter           | Type  | Required | Description                                                                                        |
| ------------------- | ----- | -------- | -------------------------------------------------------------------------------------------------- |
| start_timestamp_sec | int64 | No       | Non-negative starting Unix timestamp in seconds; defaults to `0`                                   |
| end_timestamp_sec   | int64 | No       | Non-negative ending Unix timestamp in seconds; defaults to `0` and must not be less than the start |

Omit the entire `operate_time_range` object when no time filter is needed.

### Permission Rules

- `system_access` is always required.
- When `exact_include_conditions.workflow_id` is supplied, history permissions are checked against the stored roles and businesses of every matching workflow before returning a page or count. Failure for any matching workflow fails the request.
- Without a workflow ID filter, the query is narrowed to the user's authorized businesses. `blank` or `agent` requires Agent history permission; `proxy` requires Proxy history permission. Missing or unknown roles, or both node categories, require both permissions.
- With `only_count=true`, `page` may be omitted and no detail records are returned.

### Request Example

Query execution status by a known workflow ID.

```json
{
  "page": { "offset": 0, "limit": 20 },
  "only_count": false,
  "exact_include_conditions": {
    "workflow_id": ["wf-20240601-0001"]
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
        "tenant_id": null,
        "workflow_id": "wf-20240601-0001",
        "trigger_id": "trigger-9f2b",
        "type": "upgrade_agent",
        "bk_biz_id": [2],
        "bk_networkarea_id": [0],
        "bk_networkunit_id": [1],
        "operator": "admin",
        "operate_time": 1719820000000,
        "finish_time": 0,
        "status": "running",
        "node_role": ["agent"],
        "deploy_policy_ids": []
      }
    ]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                        |
| ---------- | ------ | ---------------------------------- |
| code       | int32  | Status code; `0` indicates success |
| message    | string | Request message                    |
| request_id | string | Request ID                         |
| error      | object | Error information                  |
| permission | object | Permission information             |
| data       | object | Response data                      |

#### data

| Parameter | Type  | Description                                                            |
| --------- | ----- | ---------------------------------------------------------------------- |
| total     | int64 | Total matching workflows in the query scope, not the current page size |
| items     | array | Current page of workflows; empty when `only_count=true`                |

#### data.items[n]

| Parameter         | Type          | Description                                                                                                                         |
| ----------------- | ------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| tenant_id         | string / null | Tenant ID; currently returned as `null` in list responses. This is a response field; no tenant parameter is required in the request |
| workflow_id       | string        | Workflow ID, used to query its operations                                                                                           |
| trigger_id        | string        | Trigger ID; not the workflow ID                                                                                                     |
| type              | string        | Workflow type                                                                                                                       |
| bk_biz_id         | int64 array   | Business IDs                                                                                                                        |
| bk_networkarea_id | int64 array   | Network area IDs                                                                                                                    |
| bk_networkunit_id | int64 array   | Network unit IDs                                                                                                                    |
| operator          | string        | Historical operator                                                                                                                 |
| operate_time      | int64         | Operation time in Unix milliseconds                                                                                                 |
| finish_time       | int64         | Completion time in Unix milliseconds; `0` when not recorded                                                                         |
| status            | string        | Workflow state; values match the request `status` field                                                                             |
| node_role         | string array  | Node roles stored on the workflow                                                                                                   |
| deploy_policy_ids | int64 array   | Deployment policy IDs associated with the workflow                                                                                  |
