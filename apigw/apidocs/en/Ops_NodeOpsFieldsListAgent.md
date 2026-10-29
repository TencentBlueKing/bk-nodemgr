### Description

- API Version: v3.0.1-alpha.103+.
- Required Permission: agent_view (View Agent).
- Function: Batch-query the current out-of-band management settings of the specified hosts.

### URL

POST /api/v3/node/ops/list_fields/agent

### Input Parameters

| Parameter Name | Parameter Type | Required | Description                                                                                                                                       |
| -------------- | -------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| bk_host_id     | int64 array    | Yes      | Host ID list to query. The list must contain 1-500 IDs, and every ID must be greater than 0; duplicate host IDs are deduplicated before querying. |

**Parameter Notes**:

- Unmatched `bk_host_id` values (non-existent or deleted hosts) are ignored; the request does not return an error because of them.
- The response contains out-of-band management fields only; host information such as CMDB business, topology, and inner IP must be obtained from other APIs such as the host list.
- The response order is not guaranteed to match the order of `bk_host_id` in the request.
- Permission is checked against the CMDB business that the matched hosts belong to: if any business lacks the required permission, the whole request returns a permission denied error.
- This API does not filter results by node role; its response structure is the same as that of `/api/v3/node/ops/list_fields/proxy`, and the only difference is the view permission checked (`agent_view` / `proxy_view`).

### Request Example

Query the out-of-band management settings of two hosts.

```json
{
  "bk_host_id": [10001, 10002]
}
```

### Response Example

```json
{
  "code": 0,
  "message": "OK",
  "request_id": "req-1234567890",
  "data": {
    "items": [
      {
        "bk_host_id": 10001,
        "ops_console_host_id": 20001,
        "ops_out_band_type": "string",
        "ops_out_band_protocol": "string",
        "ops_out_band_ip": "10.0.0.1",
        "ops_out_band_port": 22
      }
    ]
  }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description                                           |
| -------------- | -------------- | ----------------------------------------------------- |
| code           | int32          | Status code, 0 means success                          |
| message        | string         | Request message                                       |
| request_id     | string         | Request ID                                            |
| error          | object         | Error information, not returned on success            |
| permission     | object         | Permission information, returned on permission denied |
| data           | object         | Response data                                         |

#### data

| Parameter Name | Parameter Type | Description                                                                                       |
| -------------- | -------------- | ------------------------------------------------------------------------------------------------- |
| items          | object array   | List of out-of-band management settings for the matched hosts, an empty array when no hosts match |

#### data.items[n]

| Parameter Name        | Parameter Type | Description                                                                                                                                                                                              |
| --------------------- | -------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| bk_host_id            | int64          | Host ID                                                                                                                                                                                                  |
| ops_console_host_id   | int64          | Out-of-band management console host ID: the host that performs out-of-band operations on the target host (usually a jump host or an operations host with the BlueKing GSE Agent installed); 0 when unset |
| ops_out_band_type     | string         | Out-of-band device type, free text: use `onboard` for onboard out-of-band management, or the device type identifier (for example `nanokvm`) for an external out-of-band device; empty string when unset  |
| ops_out_band_protocol | string         | Communication protocol used for out-of-band operations, free text: for example `ipmi` (common standard protocol) or `nanokvm-gpio` (platform extension for specific devices); empty string when unset    |
| ops_out_band_ip       | string         | Out-of-band management IP address of the device, usually independent of the host's own IP and reachable from the console host; empty string when unset                                                   |
| ops_out_band_port     | int64          | Out-of-band management port, usually the protocol default port (for example IPMI commonly uses 623); 0 when unset                                                                                        |
