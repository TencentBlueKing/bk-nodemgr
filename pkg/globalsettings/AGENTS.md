# GLOBAL SETTINGS KNOWLEDGE BASE

## OVERVIEW

`pkg/globalsettings` provides the shared global-settings facade and storage abstraction.
It owns predefined setting keys/defaults and exposes package-level error semantics for callers.

## WHERE TO LOOK

- Setting keys and predefinitions: `pkg/globalsettings/definition.go`
- Facade lifecycle and APIs: `pkg/globalsettings/globalsettings.go`
- Storage boundary interface: `pkg/globalsettings/storage.go`

## CONVENTIONS

- Keep DB/DAO details outside this package; inject storage only through `IStorage`.
- Keep this package as a shared facade; service-specific validation/orchestration belongs to `internal/*` layers.
- Reuse constants in `definition.go` for setting names; avoid duplicated literals in callers.
- `Get()` is a tolerant fallback API (returns default on exist/get failures), while mutating APIs return `ErrUninitialized()` when storage is absent.

## ANTI-PATTERNS

- Do not wire Mongo/REST/thirdparty clients directly in this package.
- Do not import `internal/*` packages from this shared package.
- Do not put service-private business policy logic into this facade.
