### Description

- API Version: v3.0.1-alpha.117+.
- Required Permissions: system_access (System Access), plugin_view (View Plugin).
- Function: Query distinct groups of visible plugins in the current tenant, including plugins that have not been installed on a host. This operation is not paginated.
- This is the backend application-state entry. It uses application-state authentication and checks system access and the authorized business scope for viewing plugins.

### URL

POST /api/v3/plugin/distinct

### Input Parameters

The request body is a JSON object.

| Parameter Name           | Parameter Type | Required | Description                                                                    |
| ------------------------ | -------------- | -------- | ------------------------------------------------------------------------------ |
| distinct_field           | object         | No       | Fields to query; omission selects all supported fields, currently only `group` |
| exact_include_conditions | object         | No       | Exact-match conditions                                                         |
| fuzzy_include_conditions | object         | No       | Fuzzy-match conditions                                                         |

#### distinct_field

| Parameter Name | Parameter Type | Required | Description                                                                             |
| -------------- | -------------- | -------- | --------------------------------------------------------------------------------------- |
| group          | bool           | No       | Query groups when `true`; `distinct_field: {}` or `group: false` does not select groups |

#### exact_include_conditions

| Parameter Name  | Parameter Type | Required | Description                                                                                                    |
| --------------- | -------------- | -------- | -------------------------------------------------------------------------------------------------------------- |
| name            | string array   | No       | Exact match on plugin names                                                                                    |
| group           | string array   | No       | Exact match on plugin groups; omit when loading group options to avoid restricting available options           |
| visible_biz_ids | int64 array    | No       | Business visibility scope, narrowed by the caller's authorization; omission uses the authorized business scope |

#### fuzzy_include_conditions

| Parameter Name | Parameter Type | Required | Description                         |
| -------------- | -------------- | -------- | ----------------------------------- |
| name           | string array   | No       | Fuzzy match on plugin names         |
| pkg_name       | string array   | No       | Fuzzy match on plugin package names |

**Scope Notes**:

- Groups come from plugin records, not host plugin processes or a page of plugin-list results. Exact and fuzzy conditions apply together.
- Supply the `visible_biz_ids` used by the plugin list and refresh the groups when the selected business changes.
- For a nonempty business scope, results include plugins with an empty, null, or missing visibility field, and plugins whose visible-business-ID array exactly equals the effective business-scope array. This is not an any-overlap membership filter.
- A caller with unrestricted business permissions and no business condition has no additional visible-business filter. A restricted caller with an empty effective scope receives an empty group array; a failed permission check returns a permission error.
- Results are not guaranteed to be sorted. Group names are not a fixed enum, and the API does not insert a `default` group when it does not exist.

### Request Example

Query visible plugin groups for business `2` without restricting results to the currently selected group.

```json
{
  "distinct_field": { "group": true },
  "exact_include_conditions": { "visible_biz_ids": [2] }
}
```

### Response Example

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-plugin-distinct",
  "error": null,
  "data": { "group": ["default", "custom"] }
}
```

### Response Parameters

| Parameter Name | Parameter Type | Description                        |
| -------------- | -------------- | ---------------------------------- |
| code           | int32          | Status code; 0 indicates success   |
| message        | string         | Request message                    |
| request_id     | string         | Request ID                         |
| error          | object         | Error information; null on success |
| data           | object         | Distinct values                    |

#### data

| Parameter Name | Parameter Type | Description                                                                                   |
| -------------- | -------------- | --------------------------------------------------------------------------------------------- |
| group          | string array   | Distinct groups of matching plugins; empty when no records match or the field is not selected |
