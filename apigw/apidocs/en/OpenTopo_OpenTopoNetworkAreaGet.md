### Description

- Available since: v3.0.1-alpha.107+.
- Required permissions: system_access (System Access), networkarea_view (View Network Area).
- Function: Query network area details by ID through the user-state API. Viewing permission is checked for the target network area.
- User identity and tenant information are supplied by API Gateway. Do not provide a username or tenant ID in the request body.

### URL

POST /api/v3/open/topo/networkarea/get

### Input Parameters

| Parameter         | Type  | Required | Description                                                   |
| ----------------- | ----- | -------- | ------------------------------------------------------------- |
| bk_networkarea_id | int64 | No       | Network area ID; must be nonnegative. Omission queries ID `0` |

### Request Example

```json
{ "bk_networkarea_id": 1 }
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "tenant_id": "default",
    "bk_networkarea_id": 1,
    "bk_networkarea_name": "Example Area",
    "cloud_vendor": "1"
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                                                                                  |
| ---------- | ------ | ------------------------------------------------------------------------------------------------------------ |
| code       | int32  | Status code; `0` indicates success                                                                           |
| message    | string | Request result message                                                                                       |
| request_id | string | Request ID                                                                                                   |
| error      | object | Error information, including `system`, `message`, and `details` (error codes and messages); empty on success |
| permission | object | Permission application information that may be returned when access is denied                                |
| data       | object | Network area details                                                                                         |

#### data

| Parameter           | Type   | Description                  |
| ------------------- | ------ | ---------------------------- |
| tenant_id           | string | Tenant ID                    |
| bk_networkarea_id   | int64  | Network area ID              |
| bk_networkarea_name | string | Network area name            |
| cloud_vendor        | string | CMDB cloud vendor identifier |
