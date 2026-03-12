# ETCD DISCOVER KNOWLEDGE BASE

## OVERVIEW

`pkg/discover/etcddiscover` implements the etcd-backed `discover.Provider` for service registration and discovery.
It owns etcd client setup, TLS wiring, lease keepalive, watch/list synchronization, and local instance cache refresh.
It also bridges etcd client internal logs into the project's unified `pkg/logger` pipeline.

Responsibilities:
- Create and manage the etcd client from `config.Etcd`.
- Register, update, deregister, list, and watch service instances stored in etcd.
- Maintain lease metadata (`etcd-lease-id`) required for update and revoke flows.
- Convert etcd client logs into structured bk-nodemgr logs.

Not responsible for:
- Defining service-specific registration payload semantics beyond `discover.Instance`.
- Business routing or caller-side endpoint selection policy.
- Exposing raw etcd client types to upper-layer business code.

## WHERE TO LOOK

| File | Purpose |
|------|---------|
| `etcd.go` | `ProviderEtcd` implementation: client init, TLS setup, watch/list loop, register/update/deregister, endpoint queries, logger bridge |
| `README.md` | Short package intent and boundary notes |
| `etcd_test.go` | Registration/discovery behavior tests; validates CRUD, query, select, and lease-backed update flows |
| `etcd_logger_test.go` | Logger integration tests for `newEtcdClientConfig`, `mapEtcdLogLevel`, and `etcdLoggerCore` |

### Key runtime pieces in `etcd.go`

| Symbol / Area | Description |
|---------------|-------------|
| `ProviderEtcd` | Main provider state: etcd client, local service cache, watch targets, runtime context |
| `NewProviderEtcd()` | Builds the provider from `config.Etcd` plus optional watch/list options |
| `Start()` / `Stop()` | Provider lifecycle; client bootstrap and background loops |
| `Register()` / `Update()` / `Deregister()` | Lease-based instance mutation lifecycle |
| `watch()` / `list()` / `keepListing()` | Read-path synchronization from etcd into local cache |
| `newEtcdClientConfig()` / `etcdLoggerCore` | Inject unified logger into the etcd client |

## CONVENTIONS

- Keep **all etcd-specific concerns inside this package**: client creation, TLS translation, lease handling, watch semantics, and etcd log adaptation should not leak into callers.
- Treat `discover.Instance` as the boundary type. Callers provide and consume `discover.Instance` / `discover.Endpoint`; they should not depend on `clientv3` details.
- Preserve `metaKeyLeaseID` on successful register/update flows. `Update()` and `Deregister()` rely on this metadata to reuse or revoke the correct lease.
- Use structured logging via `pkg/logger` only. etcd internal logs must flow through `newEtcdClientConfig()` and `etcdLoggerCore`, not a separate logger stack.
- Keep local cache refresh behavior consistent: query APIs (`GetAllService`, `GetAllEndpoint`, `GetEndpoint`, `SelectEndpoints`) read from provider-managed state populated by `watch()` / `list()`.
- Tests in `etcd_test.go` are environment-dependent integration-style tests. They expect `.env` with `ETCD_ENDPOINT` and a reachable etcd instance.

## ANTI-PATTERNS

- Do not move etcd client options, TLS wiring, or log mapping into service code or unrelated shared packages.
- Do not bypass lease management by writing directly to etcd without updating `etcd-lease-id` metadata on the in-memory instance.
- Do not hardcode endpoints, credentials, or TLS behavior outside `config.Etcd` and `initTLS()`.
- Do not return or store raw etcd responses in higher layers; convert them into `discover.Instance` / `discover.Endpoint` first.
- Do not add service-specific interpretation of instance metadata in this package; such policy belongs to callers.
