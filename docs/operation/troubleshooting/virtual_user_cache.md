# Virtual user changes temporarily return a cached username

## Symptoms

In multi-tenant mode (`tenantMode=multiple`), a virtual user's `bk_username` mapping has changed in User Management, but Node Manager may still resolve the same login name to the previous value. Requests handled by different replicas may temporarily resolve to different values.

This can affect current virtual user resolution and APIGateway client credential resolution. A brief discrepancy after a mapping change is not sufficient evidence of a configuration or authentication failure.

## Cause and scope

- Successful `login_name` to `bk_username` lookups are cached by `(tenantID, loginName)` in process memory. Replicas do not share this cache.
- Each cache entry expires one minute after it is written. Cache hits do not extend its lifetime. An expired or missing entry triggers a lookup in User Management on the next request; there is no periodic refresh.
- A previously cached mapping may remain visible for the rest of that entry's lifetime. Replicas can hold entries written at different times, so they may temporarily return different results.
- Lookup failures, missing users, duplicate matches, and empty `bk_username` values are not cached. Creating a user that previously could not be found does not require waiting for a cached failure to expire: the next lookup queries User Management again.
- Single-tenant mode returns the login name directly and does not use this virtual user lookup cache.

The one-minute TTL limits the lifetime of a local cache entry. It is not an end-to-end guarantee that a User Management change becomes visible within one minute; upstream visibility and request failures must also be considered.

## Verification

1. Confirm that the affected service uses `tenantMode=multiple`. Identify the exact tenant ID and login name used by the failing request.
2. Confirm that User Management returns the expected non-empty `bk_username` for that tenant and login name. If it does not, investigate the upstream mapping or lookup failure first.
3. Record the request time, tenant, login name, and serving replica for repeated requests, using the available request logs. Compare requests for the same tenant and login name across replicas.
4. After confirming the upstream result, wait longer than one minute and repeat the same lookup. Where possible, check each affected replica. Expiration is checked on lookup; waiting alone does not fetch a new value.
5. If the discrepancy persists, inspect the lookup error and verify the tenant, login name, User Management endpoint, and APIGateway credentials and permissions. Do not attribute a persistent failure to this cache solely because caching is enabled.

## Handling

Allow existing successful entries to expire, then retest before concluding that a transient discrepancy is a configuration or authentication fault. For missing users or lookup failures, investigate the returned error directly because those results are not cached.

The default cache TTL is fixed in code as `time.Minute`; there is no deployment configuration option for changing it. The lookup path does not actively invalidate entries when User Management changes. Configurable TTLs, explicit invalidation, or cross-replica consistency require a separate implementation decision and follow-up issue.

## Implementation references

- [User Management handler](../../../pkg/thirdparty/usermanager/handler.go): `NewHandlerMultiTenant`, `GetBKUsernameByLoginName`, and `refreshVirtualUser` define the cache policy; `HandlerSingle` bypasses the lookup.
- [Memory cache](../../../pkg/runtime/cache/memory.go): process-local storage and default expiration.
- [Current virtual user resolution](../../../pkg/access/access.go) and [APIGateway client credentials](../../../pkg/thirdparty/apigw/client/config.go): callers of the resolver.
- [Installation configuration](../installation.md): `tenantMode` and `userManager` connection settings.
