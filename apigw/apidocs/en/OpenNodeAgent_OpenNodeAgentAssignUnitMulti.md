### Description

- API Version: v3.0.1-alpha.105+
- Required Permissions: system_access (System Access), agent_operate (Operate Agent), networkunit_use_for_agent (Use Network Unit to Deploy Agent).
- Function: Assign groups of hosts without a network unit to different network units, as the authenticated user. Only updates host assignments; does not install, upgrade, or reconfigure Agents.
- The API Gateway authenticates and forwards the user identity. Do not include a username or tenant parameter in the request body. `agent_operate` is checked against all target hosts' businesses; `networkunit_use_for_agent` is checked against all target network units.

### URL

POST /api/v3/open/node/agent/assign_unit_multi

### Input Parameters

| Parameter Name | Parameter Type | Required | Description                                                |
| -------------- | -------------- | -------- | ---------------------------------------------------------- |
| items          | object array   | Yes      | Nonempty assignment group list; null items are not allowed |

#### items[n]

| Parameter Name    | Parameter Type | Required | Description                                                                                                |
| ----------------- | -------------- | -------- | ---------------------------------------------------------------------------------------------------------- |
| bk_host_id        | int64 array    | Yes      | Nonempty host ID list for the group; IDs must be unique across the whole request, including within a group |
| bk_networkunit_id | int64          | Yes      | Existing target network unit ID for the group, greater than or equal to `0`                                |

Within each group, all found hosts must share a network area matching that group's target unit. Different groups may belong to different areas. Hosts with an assigned unit cannot change units through this API and count as failures even if the unit is unchanged. Missing hosts also count as failures.

All target units and permissions are checked before any assignment. Assignment then proceeds by group. Failure in one group does not prevent success in others; the request is not atomic. Inspect counts and failure reasons, which may be aggregated by group or batch rather than corresponding one-to-one with hosts.

### Request Example

```json
{
  "items": [
    { "bk_host_id": [1001, 1002], "bk_networkunit_id": 1 },
    { "bk_host_id": [1003], "bk_networkunit_id": 2 }
  ]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "success_count": 3,
    "failed_count": 0,
    "failed_reasons": []
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description                                                              |
| -------------- | -------------- | ------------------------------------------------------------------------ |
| code           | int32          | Status code; `0` does not guarantee successful assignment for every host |
| message        | string         | Request message                                                          |
| request_id     | string         | Request ID                                                               |
| error          | object         | Error information                                                        |
| permission     | object         | Permission information                                                   |
| data           | object         | Synchronous aggregate result for all groups; no workflow ID is returned  |

#### data

| Parameter Name | Parameter Type | Description                                         |
| -------------- | -------------- | --------------------------------------------------- |
| success_count  | int64          | Total successfully assigned hosts across all groups |
| failed_count   | int64          | Total failed hosts across all groups                |
| failed_reasons | string array   | Aggregated failure reasons                          |
