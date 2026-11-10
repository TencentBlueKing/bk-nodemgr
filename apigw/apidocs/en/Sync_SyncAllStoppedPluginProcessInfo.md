### Description

- API Version: v3.0.1-alpha.107+.
- Required Permission: None.
- Function: Trigger synchronization of stopped plugin process information for all hosts.

### URL

POST /api/v3/sync/gse/plugin/process/info/stopped/all

### Input Parameters

No business parameters; submit an empty JSON object `{}`. The tenant ID and operator username are taken from the authenticated request context.

### Usage Notes

- This interface reuses the generator of the scheduled workflow `scheduled_sync_stopped_plugin_process_info`: it scans the hosts of the current tenant whose agent ID is not empty, creates synchronization tasks in batches of at most 1,000 hosts, and applies the concurrency configured by `oper_sync_stopped_plugin_process_info_max_concurrency_num` (default 100).
- Each synchronization task only handles the plugin processes whose local record status is `stopped` within its host batch. For online agents the real process information is queried from GSE; processes that GSE no longer reports are marked `unknown`. For offline agents those processes are also marked `unknown`.
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
