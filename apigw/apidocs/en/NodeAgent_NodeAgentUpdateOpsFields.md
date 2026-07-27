### Description

- API Version: v3.0.1-alpha.60+.
- Required Permission: agent_operate (Operate Agent).
- Function: Batch-update out-of-band management fields for Agent hosts and synchronize them to CMDB.

### URL

POST /api/v3/node/agent/update_ops_fields

### Input Parameters

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| hosts | object array | Yes | Host information list |

#### hosts[n]

| Parameter Name | Parameter Type | Required | Description |
|---------------|----------------|----------|-------------|
| bk_host_id | int64 | Yes | Host ID |
| ops_console_host_id | int64 | No | Out-of-band console host ID |
| ops_out_band_type | string | No | Out-of-band management type |
| ops_out_band_protocol | string | No | Out-of-band management protocol |
| ops_bmc_ip | string | No | BMC IP address |
| ops_bmc_port | int64 | No | BMC port |

### Request Example

```json
{
  "hosts": [
    {
      "bk_host_id": 1,
      "ops_console_host_id": 1,
      "ops_out_band_type": "string",
      "ops_out_band_protocol": "string",
      "ops_bmc_ip": "10.0.0.1",
      "ops_bmc_port": 22
    }
  ]
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
  "data": {}
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description |
|---------------|----------------|-------------|
| code | int32 | Status code, 0 means success |
| message | string | Request message |
| request_id | string | Request ID |
| error | object | Error information, empty on success |
| permission | object | Permission information |
| data | object | Response data |
