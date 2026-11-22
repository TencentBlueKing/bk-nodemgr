### Description

- API Version: v3.0.1-alpha.116+.
- Required Permissions: Both system_access (System Access) and platform_manage (Platform Management). The platform resource type is `platform_resource`, and the resource ID is `global_settings` (Global Settings), independently authorized from `scheduler_workflow` (Scheduled Workflow Management).
- Function: Query global settings stored in the database through the application-state API, including custom settings, with pagination, name filters, and count-only queries.
- Calls require API Gateway application authorization and the existing application-state authentication chain. Identity and tenant context come from trusted authentication, not the request body. This API shares its query handler, condition conversion, and platform management checks with the user-state endpoint.
- Settings are globally shared across tenants, not isolated to the current tenant. Values are returned as raw strings without redaction and may contain sensitive custom data. Callers should restrict access, logging, and forwarding of the results.

### URL

POST /api/v3/globalsettings/list

### Request Parameters

| Parameter                | Type   | Required    | Description                                                                                                |
| ------------------------ | ------ | ----------- | ---------------------------------------------------------------------------------------------------------- |
| page                     | object | Conditional | Pagination; valid pagination is required when `only_count=false` and may be omitted when `only_count=true` |
| only_count               | bool   | No          | Return only the matching count without setting details; defaults to `false`                                |
| exact_include_conditions | object | No          | Exact-match inclusion conditions for setting names                                                         |
| fuzzy_include_conditions | object | No          | Fuzzy inclusion conditions for setting names                                                               |
| exact_exclude_conditions | object | No          | Exact-match exclusion conditions for setting names                                                         |
| fuzzy_exclude_conditions | object | No          | Fuzzy exclusion conditions for setting names                                                               |

#### page

| Parameter | Type  | Required    | Description                                                                                                                                |
| --------- | ----- | ----------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| offset    | int32 | No          | Zero-based offset; defaults to `0` and cannot be negative                                                                                  |
| limit     | int32 | Conditional | Required for non-count queries, from `1` to `500`; count-only queries do not use pagination, but a supplied value still cannot be negative |

Even when `only_count=true`, supplied `page.offset` and `page.limit` values must pass nonnegative validation.

#### Name Filters

Each of the four condition objects supports only the optional `setting_name` field of type `string array`. Omitting an object or supplying an empty array disables that condition.

| Condition Object         | Meaning of setting_name                                                 |
| ------------------------ | ----------------------------------------------------------------------- |
| exact_include_conditions | Keep settings whose names equal any array value; case-sensitive         |
| exact_exclude_conditions | Exclude settings whose names equal any array value; case-sensitive      |
| fuzzy_include_conditions | Keep settings whose names contain any array string; case-insensitive    |
| fuzzy_exclude_conditions | Exclude settings whose names contain any array string; case-insensitive |

Multiple nonempty conditions are applied together. Fuzzy matching uses literal substrings: regular expression metacharacters are escaped and cannot be used to specify regular expressions. For example, `a.b` matches the literal text `a.b` in a name, not an arbitrary character between `a` and `b`.

### Request Example

Query a specific setting name. This example assumes that the custom setting has already been stored in the database.

```json
{
  "page": { "offset": 0, "limit": 20 },
  "only_count": false,
  "exact_include_conditions": {
    "setting_name": ["custom_integration_options"]
  }
}
```

To count settings whose names contain `integration`, pagination may be omitted:

```json
{
  "only_count": true,
  "fuzzy_include_conditions": {
    "setting_name": ["integration"]
  }
}
```

### Response Example

Successful response to the first request example:

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 1,
    "items": [
      {
        "setting_name": "custom_integration_options",
        "value": "{\"batch_size\":20}"
      }
    ]
  }
}
```

### Response Parameters

| Parameter  | Type   | Description                                 |
| ---------- | ------ | ------------------------------------------- |
| code       | int32  | Status code; `0` means success              |
| message    | string | Request message                             |
| request_id | string | Request ID                                  |
| error      | object | Error information, usually empty on success |
| permission | object | Permission information                      |
| data       | object | Response data                               |

#### data

| Parameter | Type  | Description                                                                                                              |
| --------- | ----- | ------------------------------------------------------------------------------------------------------------------------ |
| total     | int64 | Total matching settings, not the current page size                                                                       |
| items     | array | Settings on the current page; an empty array when `only_count=true`; may be `null` for a non-count query with no results |

#### data.items[n]

| Parameter    | Type   | Description                                                                                                                        |
| ------------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------- |
| setting_name | string | Stored setting name, including custom names                                                                                        |
| value        | string | Raw string stored in the database, without redaction; remains a string even if its contents represent JSON, a number, or a boolean |

### Processing Rules

- Queries cover global settings stored in the database, including names outside the predefined set. Data is not partitioned by tenant, and the request body has no tenant selector.
- This API does not read service startup configuration or fill in defaults for settings that have not been stored. Returned values do not indicate the values currently effective in running processes.
- This API is read-only and provides no setting write or delete operations.
- To retrieve one setting, supply its full name in `exact_include_conditions.setting_name`. An unmatched name returns `total=0` and an empty list without default-value fallback.
