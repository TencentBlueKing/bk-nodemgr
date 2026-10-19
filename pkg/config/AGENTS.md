# CONFIG KNOWLEDGE BASE

## OVERVIEW

`pkg/config` is the single source of truth for service startup configuration types across all services.
It defines typed config structs (infrastructure, server, front-end injection, auth) loaded from YAML
files or environment variables. Every service binary consumes one concrete `*Service` config struct
(e.g. `ApplicationService`, `BackendService`, `FileService`, `RelayService`) that aggregates shared
sub-structs from `config.go`.

Responsibilities:
- Define typed, validated configuration structs for all services.
- Provide `Load(filePath)` / `LoadFromEnv()` / `LoadFromFile()` lifecycle for each service config.
- Centralize default values for every configurable field.
- Expose `Validate()` on every struct to enforce required-field and value-range rules.

Not responsible for:
- Business logic that uses the config values.
- Any runtime state mutation after service startup.

## WHERE TO LOOK

| File | Purpose |
|------|---------|
| `config.go` | Shared infrastructure sub-structs: `Etcd`, `Redis`, `MongoDB`, `Log`, `HTTPServer`, `AuthIdentity`, `JWTServer/ClientConfig`, `Front`, `TraceService`, `TLS`, `APIGatewayClient` |
| `application.go` | `ApplicationService` — application server config; `Load*` / `Validate` lifecycle; front-end port defaults (`WindowsWMIPortDefault`, `WindowsSSHPortDefault`, `UnixSSHPortDefault`) |
| `backend.go` | `BackendService` — backend server config; GSE, CMDB, IAM, plugin, workflow worker settings |
| `file.go` | `FileService` — file server config; download port, mount host dir |
| `relay.go` | `RelayService` — relay server config |
| `kms.go` | `KMS` — shared KMS secret loading; decrypts envelopes via `pkg/kms` and maps values into each service config |
| `adminclient.go` | `AdminClientDatabase` — minimal database-sections-only config for adminclient check/proxy commands |
| `config_test.go` | Unit tests for shared sub-struct validation |

### Key types in `config.go`

| Type | Description |
|------|-------------|
| `Front` | Front-end injectable settings (vault switch, port defaults, URLs) |
| `HTTPServer` | Generic HTTP server bind IP/port with advertise IP |
| `JWTServerConfig` / `JWTClientConfig` | JWT crypto config (symmetric / asymmetric) |
| `APIGatewayClient` | Endpoint + appCode/appSecret for outbound API gateway calls |
| `TLS` | TLS cert/key/CA paths and `insecureSkipVerify` |

### Front-end port injection flow

```
env BK_NODEMGR_APPLICATION_WINDOWS_WMI_PORT_DEFAULT
env BK_NODEMGR_APPLICATION_WINDOWS_SSH_PORT_DEFAULT
env BK_NODEMGR_APPLICATION_UNIX_SSH_PORT_DEFAULT
    → application.go LoadFromEnv()
    → ApplicationService.Front.WindowsWMIPortDefault / WindowsSSHPortDefault / UnixSSHPortDefault
    → internal/application/frontsetting  (injected via Option)
    → internal/application/router/web   (rendered into index.html template)
    → window.PROJECT_CONFIG.WINDOWS_WMI_PORT_DEFAULT / WINDOWS_SSH_PORT_DEFAULT / UNIX_SSH_PORT_DEFAULT
```

## CONVENTIONS

- **One service, one `*Service` struct.** Each service binary has its own top-level config type that
  embeds shared sub-structs by value (not pointer).
- **Defaults in `NewApplicationService()` / `New*Service()`.** All default values are set once in
  the constructor. Never scatter defaults across `LoadFromEnv`.
- **`yaml:` tags drive file loading; env loading is done explicitly.** YAML field names use
  `camelCase`. Env var names follow `BK_NODEMGR_<SERVICE>_<FIELD>` pattern in `SCREAMING_SNAKE_CASE`.
- **`Validate()` must be called after `Load()`.** Validation errors should return descriptive
  `fmt.Errorf` messages referencing the field name.
- **Shared sub-structs live in `config.go`.** If more than one service config reuses a struct,
  it belongs in `config.go`, not in a service-specific file.
- **English comments required on all exported symbols** (project-wide rule).

## ANTI-PATTERNS

- Do not add business logic or derived state to config structs — they are pure value holders.
- Do not read environment variables outside of `LoadFromEnv()` in each `*Service` file.
  - **Exception (KMS):** `kms.go` reads `BK_NODEMGR_KMS_PRIVATE_KEY_*` env vars inside
    `KMS.load()`. KMS private keys are bootstrap secrets shared by all services, so they
    are centralized in one choke point instead of being duplicated across each `LoadFromEnv()`.
- Do not define a service-specific sub-struct in `config.go` — keep it in the service-specific file.
- Do not skip `Validate()` in service startup — missing required fields must fail fast.
- Do not hardcode port numbers or URLs inside business code — always read from the config struct.

## EXCEPTIONS

- **`KMS` lives in `kms.go`, not `config.go`.** Although `KMS` is embedded by
  `ApplicationService` / `BackendService` / `FileService`, it is kept in a dedicated file
  together with the `Parse*ServiceConfig` mapping functions (~250 lines) that translate the
  decrypted KMS envelope into each service's config fields. Moving only the struct into
  `config.go` would split tightly coupled code; treat `kms.go` as the KMS integration unit.
- **`AdminClientDatabase` (in `adminclient.go`) is not a `*Service` struct.** It is a
  client-side config shared by the adminclient database commands and deliberately deviates
  from two conventions: (1) it is not tied to one service binary — it aggregates only the
  database sections common to all service conf files; (2) `LoadFromFile` does not call
  `Validate()` — each command validates only the database section it uses, because the raw
  config file may contain unreplaced rendering placeholders (e.g. `__ADVERTISE_IPV4__`) in
  unrelated fields.
- **`KMS.loadGlobal()` reads env vars like `KMS.load()`.** Both are covered by the KMS
  exception above; `loadGlobal` reads only the global private key
  (`BK_NODEMGR_KMS_PRIVATE_KEY_GLOBAL`) for consumers whose secrets all live in the global
  envelope.
