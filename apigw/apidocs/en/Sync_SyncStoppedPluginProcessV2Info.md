### Description

- API Version: v3.0.1-alpha.111+.
- Required Permission: None.
- Function: Trigger synchronization of stopped V2 compatible plugin process information for the given hosts.

### URL

POST /api/v3/sync/gse/plugin/process_v2/info/stopped

### Input Parameters

| Parameter Name | Parameter Type | Required | Description   |
| -------------- | -------------- | -------- | ------------- |
| host_ids       | int64 array    | No       | Host ID list  |

### Usage Notes

- This interface only handles `process_v2` records, which are the processes of node manager V2 compatible plugins; use `Sync_SyncStoppedPluginProcessInfo` for V3 plugin processes.
- The synchronized scope is the processes whose local record status is `stopped`, complementing the V2 alive process synchronization (scheduled workflow `scheduled_sync_alive_plugin_process_v2_info`).
- For online agents the real process information is queried from GSE in the `nodeman` namespace and the results confirmed by GSE are written back to the local record; processes whose query fails or that GSE does not report keep their local `stopped` record and are not marked `unknown`.
- For offline agents the stopped state cannot be confirmed from GSE, so those processes are marked `unknown` instead of keeping the local `stopped` record.
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
