# BACKEND ADMIN THIRDPARTY KNOWLEDGE BASE

## OVERVIEW

`pkg/thirdparty/backendadmin` provides standardized integration with backend admin APIs.
It wraps `/admin/globalsettings/*` endpoints and converts transport payloads to/from `pkg/types` models.

## WHERE TO LOOK

- Client bootstrap and header/JWT flow: `pkg/thirdparty/backendadmin/backendadmin.go`
- Public facade and dependency boundary: `pkg/thirdparty/backendadmin/handler.go`
- Network-unit segment rules API mapping: `pkg/thirdparty/backendadmin/networkunit_segment_rules.go`
- Contract tests and seam behavior: `pkg/thirdparty/backendadmin/networkunit_segment_rules_test.go`

## CONVENTIONS

- Keep endpoint paths, response code checks, and JSON mapping local to this package.
- Expose business-facing methods through `IHandler`; do not leak proto structs to callers.
- Propagate `contextx.IContext` through all requests.
- Reuse `pkg/globalsettings.NetworkUnitSegmentRules` constant instead of ad-hoc setting-name literals.
- Keep test seams via `cli` function hooks for deterministic unit tests.

## ANTI-PATTERNS

- Do not place business orchestration/policy decisions in this package.
- Do not call backend admin endpoints directly from upper business modules when this handler exists.
- Do not duplicate header/JWT assembly logic in callers.
