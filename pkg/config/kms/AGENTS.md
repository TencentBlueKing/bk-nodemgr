# KMS KNOWLEDGE BASE

## OVERVIEW

`pkg/kms` decrypts KMS secret envelopes via `bk-kms-sdk` and unmarshals them into typed
`envelope` structs. It is the single boundary between encrypted secret blobs (files + private
keys supplied at deploy time) and the rest of the codebase, which only ever sees decrypted,
typed values.

Responsibilities:
- Decrypt the global envelope and the per-module envelope with the provided private keys.
- Unmarshal plaintext JSON into `pkg/kms/envelope` structs.
- Expose one load entrypoint per service (`application` / `backend` / `file`).

Not responsible for:
- Reading secret files or private-key environment variables (that is `pkg/config` `kms.go`).
- Mapping envelope values into service config fields (that is `pkg/config` `Parse*ServiceConfig`).
- Caching, rotation, or any runtime secret refresh — load happens once at service startup.

## WHERE TO LOOK

| File        | Purpose                                                                                                                                                  |
| ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kms.go`    | `LoadParam` input contract; per-service entrypoints (`LoadApplicationEnvelopes` / `LoadBackendEnvelopes` / `LoadFileEnvelopes`) plus the global-only `LoadGlobalEnvelope` (for consumers whose secrets all live in the global envelope, e.g. adminclient database check/proxy); internal per-envelope decrypt helpers |
| `envelope/` | Typed structs for decrypted global and per-module envelope JSON                                                                                          |

### Call flow

```
pkg/config kms.go KMS.load()                (reads envelope files + private-key env vars)
    → pkg/kms LoadKMS<Service>Secret(param) (decrypt + unmarshal)
    → pkg/kms/envelope KMS<Service>Envelope (typed result)
    → pkg/config kms.go Parse*ServiceConfig (maps values into *Service config fields)
```

Global-only consumers (database credentials) use a shortened path:

```
pkg/config kms.go KMS.loadGlobal()          (reads only the global envelope file + global key env)
    → pkg/kms LoadGlobalEnvelope(param)     (decrypt + unmarshal global envelope only)
    → pkg/config kms.go ParseAdminClientDatabase (maps etcd/redis/mongodb credentials)
```

## CONVENTIONS

- **One `LoadKMS*Secret` entrypoint per service.** Each returns the shared global envelope plus
  its own module envelope type. New services add a new entrypoint and a new module envelope
  type — do not widen existing signatures (OCP).
- **Global-only consumers use `LoadGlobalEnvelope`.** Consumers whose secrets all live in the
  global envelope (database credentials: etcd/redis/mongodb) must not require a module private
  key; they use the global-only entrypoint instead of a per-service one.
- **Secrets enter only via `LoadParam`.** This package never touches `os.Getenv` or
  `os.ReadFile`; callers assemble `LoadParam`.
- **Errors are wrapped with context.** Every decrypt/unmarshal failure returns
  `fmt.Errorf("failed to ...: %w", err)` naming the envelope that failed.
- **English comments required on all exported symbols** (project-wide rule).

## ANTI-PATTERNS

- Do not log envelopes, plaintext, private keys, or any decrypted field — secrets must never
  reach `pkg/logger`.
- Do not add service-config field mapping here — transport boundary only; mapping lives in
  `pkg/config`.
- Do not add retry, caching, or background refresh — startup-time, fail-fast, single load.
- Do not introduce generic `map[string]any` passthroughs — every consumed secret must be a
  typed field in `envelope/`.
