### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: None.
- Function: Submit offline install result data to resume and continue the corresponding offline install operation step.

### URL

POST /api/v3/node/workflow/operation/offline/result

### Input Parameters

| Parameter Name | Parameter Type | Required | Description                                                                    |
| -------------- | -------------- | -------- | ------------------------------------------------------------------------------ |
| operation_id   | string         | Yes      | Operation ID, must point to an offline install operation                       |
| result_data    | string         | Yes      | Offline install result JSON string, usually the content of installer.data.json |

### Request Example

```json
{
  "operation_id": "oper-20260421-0001",
  "result_data": "{\"status\":\"success\",\"message\":\"offline install completed\",\"instance_id\":\"inst-123\"}"
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": null
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description                                 |
| -------------- | -------------- | ------------------------------------------- |
| code           | int32          | Status code, `0` means success              |
| message        | string         | Request message                             |
| request_id     | string         | Request ID                                  |
| error          | object         | Error information, usually empty on success |
| permission     | object         | Permission information                      |
| data           | object         | Response data, currently empty              |

#### error

| Parameter Name | Parameter Type | Description             |
| -------------- | -------------- | ----------------------- |
| system         | string         | Error system identifier |
| message        | string         | Error message           |
| details        | array          | Error detail list       |

#### error.details[n]

| Parameter Name | Parameter Type | Description   |
| -------------- | -------------- | ------------- |
| code           | string         | Error code    |
| message        | string         | Error message |

#### permission

| Parameter Name | Parameter Type | Description            |
| -------------- | -------------- | ---------------------- |
| system         | string         | Permission system ID   |
| system_name    | string         | Permission system name |
| apply_url      | string         | Permission apply URL   |
| actions        | array          | Related action list    |

#### permission.actions[n]

| Parameter Name         | Parameter Type | Description            |
| ---------------------- | -------------- | ---------------------- |
| id                     | string         | Action ID              |
| name                   | string         | Action name            |
| related_resource_types | array          | Related resource types |
