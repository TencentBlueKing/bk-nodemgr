### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: None.
- Function: Download the offline installation package by operation ID, returned as a tar.gz binary stream.

### URL

POST /api/v3/node/workflow/operation/offline/download

### Request Parameters

| Parameter    | Type   | Required | Description                                                  |
| ------------ | ------ | -------- | ------------------------------------------------------------ |
| operation_id | string | Yes      | Operation ID used to locate the offline installation package |

### Request Example

```json
{
  "operation_id": "op-20260421-0001"
}
```

### Response Example

```http
HTTP/1.1 200 OK
Content-Type: application/gzip
Content-Disposition: attachment; filename=<package_name>.tar.gz

<binary gzip stream>
```

### Response Parameters

| Parameter           | Type   | Description                                                   |
| ------------------- | ------ | ------------------------------------------------------------- |
| Content-Type        | string | Always `application/gzip`                                     |
| Content-Disposition | string | Attachment header, filename format is `<package_name>.tar.gz` |
| body                | binary | tar.gz binary content stream                                  |

### Notes

- If the request body is empty or `operation_id` is empty, the API returns an invalid parameter error.
- This API returns a streaming download, clients should read and save the response as a stream.
