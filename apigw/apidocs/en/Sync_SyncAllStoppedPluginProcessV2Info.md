### Description

- API Version: v3.0.1-alpha.111+.
- Required Permission: None.
- Function: Trigger synchronization of stopped V2 compatible plugin process information for all hosts.

### URL

POST /api/v3/sync/gse/plugin/process_v2/info/stopped/all

### Input Parameters

No business parameters; submit an empty JSON object `{}`. The tenant ID and operator username are taken from the authenticated request context.

### Usage Notes

- This interface reuses the generator of the scheduled workflow `scheduled_sync_stopped_plugin_process_v2_info`: it scans the hosts of the current tenant whose agent ID is not empty, creates synchronization tasks in batches of at most 1,000 hosts, and applies the concurrency configured by `oper_sync_stopped_plugin_process_info_max_concurrency_num` (default 100).
- Each synchronization task only handles the `process_v2` records whose status is `stopped` within its host batch. For online agents the real process information is queried from GSE in the `nodeman` namespace and the results confirmed by GSE are written back; processes whose query fails or that GSE does not report keep their local `stopped` record. For offline agents those processes are marked `unknown`.
- This interface only covers node manager V2 compatible plugin processes; use `Sync_SyncAllStoppedPluginProcessInfo` for V3 plugin processes.
- A successful response only means the workflow was created and activated, not that the synchronization finished; check the workflow execution record for the result.

### Request Example

```json
{}
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
