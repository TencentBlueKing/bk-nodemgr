# IAM V4 Callback Error Mapping

The V4 resource callback uses independent JSON DTOs and supports only `list_instance` and `fetch_instance_info`. The router owns HTTP status codes and response envelopes; providers expose typed queries and preserve error chains.

## Response Mapping

| HTTP status | `error.code`       | Scenario                                                                   | `error.message`                     |
| ----------- | ------------------ | -------------------------------------------------------------------------- | ----------------------------------- |
| 200         | Not present        | Query succeeded                                                            | Not present                         |
| 400         | `INVALID_ARGUMENT` | Invalid tenant, JSON, pagination, IDs or filters                           | `invalid callback arguments`        |
| 401         | `UNAUTHENTICATED`  | Missing or invalid Basic Auth credentials                                  | `invalid callback credentials`      |
| 404         | `NOT_FOUND`        | Unsupported callback method or unregistered resource type                  | `resource type or method not found` |
| 500         | `INTERNAL`         | Token lookup, context, provider, storage or response serialization failure | `resource query failed`             |

Success is `{"data": ...}`: `list_instance` returns `{"count": ..., "results": [...]}`, and `fetch_instance_info` returns an array. Errors use `{"error":{"code":"...","message":"..."}}`, not legacy numeric business codes. The incoming `X-Request-Id` is echoed on success and failure. A 401 response also includes `WWW-Authenticate: Basic realm="IAM"`.

## Authentication Order

The router resolves the tenant before checking credentials or retrieving the system token. Single-tenant mode uses the fixed tenant; multi-tenant mode validates the tenant header first. Invalid multi-tenant input returns 400, even if credentials are missing. Basic Auth uses username `bk_iam` and the tenant-scoped IAM system token. Invalid credentials return 401; token retrieval failures return 500, not 401.

## Argument Validation

- The body must contain one JSON object with nonblank `type` and `method`. When present, `filter` and `page` must be objects, and top-level `requires` must be a non-null string array.
- `list_instance` requires `page.page >= 1` and `page.page_size` in `1..1000`. `parseCallbackPage` converts these to `types.Page` offset/limit and rejects offset overflow. There are no legacy proto pagination defaults or `page.limit` input semantics.
- `filter.keyword` is an optional string. `filter.parent`, when supplied, requires nonempty `type` and `id`; resource providers validate the parent relationship and ID format. A valid but nonexistent parent returns an empty list, not 404.
- `fetch_instance_info` requires `filter.ids` as a non-null string array with at most `MaxFetchInstanceIDs` (1000) entries and no empty strings. An empty array returns `{"data":[]}` without querying storage for a registered resource type. Missing instances are omitted, not reported as 404.
- `requires` belongs at the request top level, not inside `filter`. Omitted or empty `requires` selects all supported attributes; unknown attributes are ignored, and `id` is always returned. Callback `_bk_iam_path_` is a single ancestor-path string; runtime enrichment retains path arrays and batches IDs in groups of at most 1000.

## Query Boundaries

`ListInstance` enumerates current-tenant candidates; omitting `parent` does not grant authorization. Parent and keyword filters intersect. Pages use a filtered total and deterministic ordering, with empty arrays for empty or out-of-bounds results. The authorizer's `listAllResourceIDs` reads all pages and rejects changing totals, incomplete pages, and empty or duplicate IDs rather than accepting truncated results.

Provider validation uses `ErrInvalidArgument`; unregistered resource types use `ErrNotFound`. The router maps wrapped sentinels with `errors.Is` and maps other failures to 500. Internal details are logged with request context and request ID, while public error messages remain fixed. This callback does not define legacy 406, 422 or 429 mappings.

## Implementation References

- [V4 router](../../../router/api-v3/iam/v4/v4.go): `basicAuthMiddleware`, `parseCallbackRequest`, `parseCallbackPage`, `parseListFilter`, `query`, `handleResourceCallback`, `callbackError`.
- [Typed query interfaces](iface.go): `IQueryHandler`, `IInstanceLister`, `IAttributeEnricher`.
- [Provider contracts](provider.go): `ErrInvalidArgument`, `ErrNotFound`, `MaxListInstancePageSize`, `MaxFetchInstanceIDs`, `Request`, `InstanceInfo.MarshalJSON`, `BuildIAMPath`.
- [Query handler](handler.go): `ListInstance`, `FetchInstanceInfo`, `FetchResourceAttributes`.
- [V4 authorizer](../auth_iamv4.go): `NewProviderHandler`, `listAllResourceIDs`.
- [IAM V4 client handler](../../../../../pkg/thirdparty/iamv4/handler.go): `IsBasicAuthAllowed`.
