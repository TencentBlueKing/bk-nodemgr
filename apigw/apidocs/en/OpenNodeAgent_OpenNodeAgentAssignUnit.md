### Description

- API Version: v3.0.1-alpha.105+
- Required Permissions: system_access (System Access), agent_operate (Operate Agent), networkunit_use_for_agent (Use Network Unit to Deploy Agent).
- Function: Assign the same network unit to hosts that do not yet have one, as the authenticated user. Only updates the hosts' network unit assignment; does not install, upgrade, or reconfigure Agents.
- The API Gateway authenticates and forwards the user identity. Do not include a username or tenant parameter in the request body. `agent_operate` is checked against the target hosts' businesses; `networkunit_use_for_agent` is checked against the target network unit.

### URL

POST /api/v3/open/node/agent/assign_unit

### Input Parameters

| Parameter Name    | Parameter Type | Required | Description                                                   |
| ----------------- | -------------- | -------- | ------------------------------------------------------------- |
| bk_host_id        | int64 array    | Yes      | Nonempty host ID list; duplicate IDs are deduplicated         |
| bk_networkunit_id | int64          | Yes      | Existing target network unit ID, greater than or equal to `0` |

All found hosts must belong to the same network area, which must match the target unit's network area; otherwise the request fails. Hosts with an assigned unit cannot change units through this API and count as failures even if the unit is unchanged. Missing hosts count as failures. Partial success is possible; inspect the counts and failure reasons, not only `code`.

### Request Example

```json
{
  "bk_host_id": [1001, 1002],
  "bk_networkunit_id": 1
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "success_count": 2,
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
| data           | object         | Synchronous assignment result; no workflow ID is returned                |

#### data

| Parameter Name | Parameter Type | Description                                                                                        |
| -------------- | -------------- | -------------------------------------------------------------------------------------------------- |
| success_count  | int64          | Number of successfully assigned hosts                                                              |
| failed_count   | int64          | Number of failed hosts                                                                             |
| failed_reasons | string array   | Failure reasons; may be aggregated by batch rather than corresponding one-to-one with failed hosts |
