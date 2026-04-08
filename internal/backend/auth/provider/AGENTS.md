# IAM V3 Provider Agent Guide

## Scope

This directory implements IAM callback providers for Node Manager resources under
`internal/backend/router/api-v3/iam/v3/provider`.

Core responsibilities:

- Validate and dispatch IAM callback requests by `type` and `method`.
- Convert callback payloads into typed request structures.
- Query backend storage and map data into IAM callback response schema.
- Return stable empty results for callback methods that are not used in current scenarios.
- **Provide canonical resource construction helpers for router permission checks** (e.g., `BuildPackageResources`).

- Validate and dispatch IAM callback requests by `type` and `method`.
- Convert callback payloads into typed request structures.
- Query backend storage and map data into IAM callback response schema.
- Return stable empty results for callback methods that are not used in current scenarios.

## File Map

- `dispatcher.go`: request entrypoint, method validation, provider registration and dispatch.
- `provider.go`: shared callback request/response models and provider interface contract.
- `networkarea.go`: IAM provider for `networkarea` resources.
- `networkunit.go`: IAM provider for `networkunit` resources.
- `package.go`: IAM provider for `package` resources (uses Distinct+pagination pattern).
- `packagetype.go`: IAM provider for `package_type` resources.
- `expression_helper.go`: IAM policy expression evaluation and pagination utilities.
- `constants.go`: resource type constants and display name mappings.
- `provider_test.go`: comprehensive unit tests with mock storage (13 test functions).
- `ERROR_CODES.md`: IAM error code mapping and usage conventions.

## Contract And Data Flow

1. `Dispatch()` binds JSON body into proto callback request.
2. Validate callback method (`list_attr`, `list_instance`, `search_instance`, etc.).
3. Resolve provider by `req.type`.
4. Convert `filter` to typed filter struct.
5. Convert and validate page by `protoBackend.ConvIAMCallbackPageToTypes`.
6. Invoke provider method and return IAM-compatible payload.

Important constraints:

- `filter` fields are dynamically decoded from `map[string]interface{}`.
- Pagination limits are enforced in page conversion logic (see proto conversion).
- `InstanceInfo` uses custom JSON marshal/unmarshal to flatten dynamic attributes.
- `AttributeValueID` supports mixed types (string/int/bool) via custom UnmarshalJSON.

## Implementation Conventions

- Keep provider methods deterministic: return empty list instead of `nil` slices.
- Keep compatibility with mixed ID types from IAM (`string` / `number`) and convert safely.
- Use `pkg/runtime/conv` for conversion when available.
- Log storage/query failures with `logger.G.Biz(ctx)` before wrapping errors.
- Keep search behavior case-insensitive where keyword search is supported.
- **Pagination pattern**: Use storage Distinct → sort → in-memory pagination (NOT List+dedup).
- **Error propagation**: Always propagate storage errors (use `%w`), never swallow with Warn.
- **Resource construction**: Each provider file exports `Build{ResourceType}Resources()` helpers that construct `types.AuthResource` slices matching the resource ID format returned by IAM callbacks. Router handlers MUST use these canonical helpers for permission checks to ensure consistency between IAM provider resource IDs and authorization resource IDs.

## Pagination & Deduplication Pattern

**Correct pattern** (package.go):
```go
// 1. Get distinct names from storage (database-level dedup)
names, err := p.storage.DistinctNameReleasePlugin(ctx)
if err != nil {
    return nil, fmt.Errorf("failed to get distinct plugin names: %w", err)
}

// 2. Sort for stable pagination
sort.Strings(names)

// 3. Apply in-memory pagination
paginatedNames, total := distinctNamesWithPagination(names, page)
```

**Why**: List+in-memory dedup breaks pagination (Count inaccurate, results unstable).

## Error Handling Rules

- Follow `ERROR_CODES.md` mapping for IAM callback semantics.
- Input/validation failures should map to parameter errors.
- Unsupported `type` should map to not found semantics.
- Unexpected provider/storage failures should be wrapped as unknown/internal errors.
- Use `fmt.Errorf("...: %w", err)` to preserve error chain.

## Extending A New Resource Provider

When adding a new IAM resource type in this directory:

1. Define `ResourceTypeXxx` constant and `XxxProvider` struct.
2. Implement the full `IProvider` interface in `provider.go`.
3. Reuse shared request/response structs and typed filters.
4. Register provider in IAM v3 route initialization (outside this directory).
5. Keep unsupported callback methods returning explicit empty results.
6. Add unit tests in `provider_test.go` with mock storage.

## Testing

- `provider_test.go`: 13 test functions covering all providers.
- Mock pattern: Hand-written mocks implementing storage interfaces.
- Coverage: boundary cases (empty, beyond bounds), error propagation, keyword filtering.
- Run: `go test ./internal/backend/router/api-v3/iam/v3/provider/... -v`

## Anti-Patterns

- Do not bypass dispatcher validation by directly parsing raw callback body in providers.
- Do not return proto structs directly as business models; keep provider response types.
- Do not return partial `nil` payload shapes that break IAM callback schema expectations.
- Do not introduce heavy business logic here; keep this layer as adapter + query orchestration.
- Do not use List+in-memory dedup for pagination (breaks Count accuracy and stability).
- Do not swallow storage errors with Warn logs (always propagate with `%w`).
