### Description

- Available since: v3.0.1-alpha.91+
- Required permissions: no additional IAM action; existing authentication and tenant isolation apply.
- Synchronously previews an enabled deploy policy and returns each Target’s identity and satisfied, unsatisfied or unmanaged status for each original Spec.
- Initially available on Backend only.

### URL

POST /api/v3/deploy_policy/preview

### Request parameters

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| deploy_policy_id | int64 | Yes | Non-negative policy ID in the current tenant. Zero identifies a specific policy; it does not select all policies. |

### Request example

```json
{"deploy_policy_id": 1001}
```

### Response example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-example",
  "data": {
    "items": [
      {
        "spec": {
          "type": "specify_plugin",
          "param": {
            "plugin_name": "example",
            "version": "2.0",
            "custom_config_context": {}
          }
        },
        "results": [
          {
            "target": {
              "host": {
                "bk_host_id": 1
              },
              "service_instance": null
            },
            "status": "satisfied"
          },
          {
            "target": {
              "host": {
                "bk_host_id": 2
              },
              "service_instance": null
            },
            "status": "unsatisfied"
          },
          {
            "target": {
              "host": {
                "bk_host_id": 3
              },
              "service_instance": null
            },
            "status": "unmanaged"
          },
          {
            "target": {
              "host": {
                "bk_host_id": 4
              },
              "service_instance": {
                "id": 101,
                "bk_module_id": 20
              }
            },
            "status": "satisfied"
          }
        ]
      }
    ]
  }
}
```

### Response fields

| Name | Type | Description |
| --- | --- | --- |
| code | int32 | Status code; zero means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error details |
| permission | object | Permission details |
| data | object | Complete calculation result |
| data.items | array | Items in the requested policy's original Spec order; empty Specs produce an empty array |
| data.items[].spec | object | Original type and param, using the same definition as policy list responses |
| data.items[].results | array | Unified target result list; [] when there are no targets |
| data.items[].results[].target | object | Target identity from the current Scope calculation |
| data.items[].results[].target.host | object | Host identity, present for both host and service instance targets |
| data.items[].results[].target.host.bk_host_id | int64 | CMDB host ID |
| data.items[].results[].target.service_instance | object/null | Service instance identity; null for host targets |
| data.items[].results[].target.service_instance.id | int64 | CMDB service instance ID |
| data.items[].results[].target.service_instance.bk_module_id | int64 | CMDB module ID of the service instance |
| data.items[].results[].status | string | satisfied: retained with no change task for this Spec; unsatisfied: retained with a change task for this Spec; unmanaged: excluded by existing policy conflict resolution |

### Semantics and limits

- Every request reruns related-policy discovery, Scope calculation, conflict resolution and analysis. There is no pagination or reuse of historical execution results.
- Related policies participate in calculation, but only the requested policy is returned. All Specs in a policy share its resolved target set.
- Satisfaction uses execute's existing change-task decisions, including skipped cases. No additional configuration-content, process-health or on-host file checks are introduced.
- Target sets retain existing Scope calculation and deduplication semantics. Each Spec's results are sorted by host ID, then module ID; empty lists serialize as []. Identity fields come directly from this calculation, without additional CC queries.
- Only targets in the current calculated Scope appear. Cleanup tasks for targets outside this Scope do not add response instances.
- The query does not execute change tasks, create triggers, operations or workflows, or update execution timestamps or dsu_id.
- Invalid input, missing or disabled policies, discovery errors, CC failures and analysis errors fail the entire request without partial results.
- Results reflect existing data sources at query time without active host probing or an atomic cross-source snapshot. Runtime depends on Scope size, related policies and dependencies; existing request cancellation and timeout limits apply.
