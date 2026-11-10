### Description

- API Version: v3.0.1-alpha.107+.
- Required Permission: None.
- Function: Trigger synchronization of stopped plugin process information for specified hosts.

### URL

POST /api/v3/sync/gse/plugin/process/info/stopped

### Input Parameters

| Parameter Name | Parameter Type | Required | Description  |
| -------------- | -------------- | -------- | ------------ |
| host_ids       | int64 array    | No       | Host ID list |

### Usage Notes

- This interface synchronizes plugin processes whose local record status is `stopped`, complementing the alive process synchronization (`Sync_SyncAlivePluginProcessInfo`).
- For hosts whose agent is online, the real process information is queried from GSE and written back to the local record. Processes that GSE no longer reports are marked `unknown`.
- For hosts whose agent is offline, GSE cannot be queried, so those processes are also marked `unknown`; the local `stopped` record is not kept.
- A successful response only means the workflow was created and activated, not that the synchronization finished; check the workflow execution record for the result.

### Request Example

```json
{
  "host_ids": [1]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "error": null,
  "permission": null,
  "data": {
    "trigger_id": "trigger-001"
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description                                                 |
| -------------- | -------------- | ----------------------------------------------------------- |
| code           | int32          | Status code, 0 means the workflow was created and activated |
| message        | string         | Request message                                             |
| request_id     | string         | Request ID                                                  |
| error          | object         | Error information, empty on success                         |
| permission     | object         | Permission information                                      |
| data           | object         | Response data                                               |

#### data

| Parameter Name | Parameter Type | Description                                          |
| -------------- | -------------- | ---------------------------------------------------- |
| trigger_id     | string         | Trigger ID of the workflow, not the execution result |
