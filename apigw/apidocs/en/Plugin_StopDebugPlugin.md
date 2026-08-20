### Description

- API version: v3.0.1-alpha.72+.
- Required permission: ⚠️ To be confirmed (handler not yet implemented).
- Description: Stops a running plugin debug workflow.

### URL

POST /api/v3/plugin/stop_debug

### Request Parameters

| Parameter   | Type   | Required | Description                                                |
| ----------- | ------ | -------- | ---------------------------------------------------------- |
| workflow_id | string | Yes      | Debug workflow ID to stop, returned by the start debug API |

### Request Example

```json
{
  "workflow_id": "workflow-abc123def456"
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {}
}
```

### Response Parameters

| Parameter  | Type   | Description                  |
| ---------- | ------ | ---------------------------- |
| code       | int32  | Status code, 0 means success |
| message    | string | Request message              |
| request_id | string | Request ID                   |
| error      | object | Error info, null on success  |
| data       | object | Response data, empty object  |

#### error

| Parameter | Type   | Description             |
| --------- | ------ | ----------------------- |
| system    | string | Error system identifier |
| message   | string | Error message           |
| details   | array  | Error detail list       |

#### error.details[n]

| Parameter | Type   | Description   |
| --------- | ------ | ------------- |
| code      | string | Error code    |
| message   | string | Error message |
