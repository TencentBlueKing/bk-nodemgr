### Description

- API Version: v3.0.1-alpha.83+.
- Required Permission: none.
- Function: Provide Node Manager resource instance lists and details to BlueKing IAM V4 callbacks, dispatching requests by resource type and callback method.

### URL

POST /api/v3/iam/v4/resource

### Input Parameters

#### Request Headers

| Parameter      | Type   | Required    | Description                                                                                                                     |
| -------------- | ------ | ----------- | ------------------------------------------------------------------------------------------------------------------------------- |
| Authorization  | string | Yes         | Basic Auth credentials. The username must be `bk_iam`, and the password must be the IAM V4 system token for the current tenant. |
| Content-Type   | string | Yes         | `application/json`.                                                                                                             |
| X-Bk-Tenant-Id | string | Conditional | Required and must be a valid tenant ID in multi-tenant mode; single-tenant mode uses the fixed system tenant.                   |
| X-Request-Id   | string | No          | Request trace identifier. The callback echoes the received value in both successful and failed response headers.                |

#### Request Body

| Parameter | Type   | Required    | Description                                                                                                                                                                                                        |
| --------- | ------ | ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| type      | string | Yes         | Nonempty resource type. Registered types are `biz`, `networkarea`, `networkunit`, `package_type`, and `package`.                                                                                                   |
| method    | string | Yes         | Only `list_instance` and `fetch_instance_info` are supported. Must be a string; numeric method identifiers are not supported.                                                                                      |
| filter    | object | Conditional | Optional for `list_instance`; required with `ids` for `fetch_instance_info`. Must not be `null` when supplied.                                                                                                     |
| page      | object | Conditional | Required only for `list_instance`; `fetch_instance_info` does not use pagination. Must not be `null` when supplied.                                                                                                |
| requires  | array  | No          | Top-level string array used only for `fetch_instance_info` attribute selection. Omitted or `[]` selects all supported attributes; unknown attributes are ignored, and `id` is always returned. Must not be `null`. |

#### filter: list_instance

| Parameter | Type   | Required | Description                                                                                                                                                                                                              |
| --------- | ------ | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| parent    | object | No       | Direct parent with nonempty string fields `type` and `id`. `networkunit` accepts a `networkarea` parent, and `package` accepts a `package_type` parent. `biz`, `networkarea`, and `package_type` do not accept a parent. |
| keyword   | string | No       | Resource display-name filter. Intersected with `parent` when both are supplied.                                                                                                                                          |

Omitting `parent` queries candidates in the current tenant; it does not grant resource permissions.

#### filter: fetch_instance_info

| Parameter | Type  | Required | Description                                                                                                                                                                           |
| --------- | ----- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ids       | array | Yes      | An array of nonempty strings with at most 1000 entries. The array itself may be `[]`, which returns `data: []`; it must not be `null`. Missing instances are omitted from the result. |

#### page: list_instance

| Parameter | Type  | Required | Description                                                                                     |
| --------- | ----- | -------- | ----------------------------------------------------------------------------------------------- |
| page      | int64 | Yes      | Page number starting at 1, with no default. The calculated pagination offset must not overflow. |
| page_size | int64 | Yes      | Page size from 1 to 1000, with no default.                                                      |

### Request Examples

The following examples show request bodies; callers must also supply the headers described above. Resource IDs and names are illustrative.

#### list_instance

```json
{
  "type": "networkarea",
  "method": "list_instance",
  "filter": {
    "keyword": "default"
  },
  "page": {
    "page": 1,
    "page_size": 100
  }
}
```

#### fetch_instance_info

```json
{
  "type": "networkarea",
  "method": "fetch_instance_info",
  "filter": {
    "ids": ["2"]
  },
  "requires": ["display_name"]
}
```

### Response Examples

#### list_instance: HTTP 200

```json
{
  "data": {
    "count": 1,
    "results": [
      {
        "id": "2",
        "display_name": "default-network-area"
      }
    ]
  }
}
```

#### fetch_instance_info: HTTP 200

```json
{
  "data": [
    {
      "id": "2",
      "display_name": "default-network-area"
    }
  ]
}
```

#### Invalid Arguments: HTTP 400

```json
{
  "error": {
    "code": "INVALID_ARGUMENT",
    "message": "invalid callback arguments"
  }
}
```

### Response Parameters

| Parameter | Type         | Description                                                                                                                               |
| --------- | ------------ | ----------------------------------------------------------------------------------------------------------------------------------------- |
| data      | object/array | Returned on success. An object with `count` and `results` for `list_instance`, or an array of instance details for `fetch_instance_info`. |
| error     | object       | Returned on failure with string fields `code` and `message`, without success data.                                                        |

#### data: list_instance

| Parameter               | Type   | Description                                                                                        |
| ----------------------- | ------ | -------------------------------------------------------------------------------------------------- |
| count                   | int64  | Total number of matching resources after filtering, not the current page size.                     |
| results                 | array  | Resource instances on the current page. `[]` when no instances match or the page is out of bounds. |
| results[n].id           | string | Unique resource instance identifier.                                                               |
| results[n].display_name | string | Resource instance display name.                                                                    |

#### data[n]: fetch_instance_info

| Parameter     | Type   | Description                                                                                                                                                            |
| ------------- | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| id            | string | Unique resource instance identifier, always returned.                                                                                                                  |
| display_name  | string | Resource display name, subject to top-level `requires` selection.                                                                                                      |
| _bk_iam_path_ | string | Ancestor path for instances with a parent, such as `/networkarea,2/` for a network unit, subject to `requires` selection. The callback returns a string, not an array. |

Supported attributes are flattened into each instance object, without a nested `attributes` object. Absent or unselected attributes are omitted.

#### error

| HTTP Status | error.code       | Description                                                                                             |
| ----------- | ---------------- | ------------------------------------------------------------------------------------------------------- |
| 400         | INVALID_ARGUMENT | Invalid JSON, arguments, pagination, or tenant ID in multi-tenant mode.                                 |
| 401         | UNAUTHENTICATED  | Missing or invalid Basic Auth credentials. The response includes `WWW-Authenticate: Basic realm="IAM"`. |
| 404         | NOT_FOUND        | Unregistered resource type or unsupported callback method.                                              |
| 500         | INTERNAL         | Internal errors such as token lookup, storage query, or serialization failures.                         |

### Notes

- This is a dedicated IAM V4 callback endpoint, not intended for normal business users. It is registered only when the service has an IAM V4 handler configured.
- APIGW user verification, application verification, and resource permission checks are disabled, but the backend still validates dedicated Basic Auth credentials.
- V4 reports failures through HTTP status codes and `error.code`/`error.message`, not the V3 top-level `code`/`message` fields. Pagination does not use V3 `offset`/`limit` either.
