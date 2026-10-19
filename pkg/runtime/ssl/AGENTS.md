# SSL KNOWLEDGE BASE

## OVERVIEW

`pkg/runtime/ssl` converts TLS material (CA/Cert/Key **content**, not file paths) into standard
library `*tls.Config` for both client and server use.

* `TLSConfig` is a pure value holder: `InsecureSkipVerify`, `VerifyClient`, `CA`, `Cert`, `Key` (`[]byte`), `Password`.
* `NewClientTLSConf()` builds a client-side `*tls.Config` (RootCAs + client certificate).
* `NewServerTLSConf()` builds a server-side `*tls.Config` (ClientCAs + certificate, `MinVersion` TLS 1.2,
  optional `RequireAndVerifyClientCert` when `VerifyClient` is set).

This package does **not** read files, resolve configuration, or hold any business logic.

## WHERE TO LOOK

| Task | Location |
|------|----------|
| TLSConfig struct / IsEnabled | `ssl.go` |
| Client tls.Config construction | `ssl.go` → `NewClientTLSConf` |
| Server tls.Config construction | `ssl.go` → `NewServerTLSConf` |
| CA pool loading | `ssl.go` → `loadCa` |
| Cert/key pair loading, password-protected key decrypt | `ssl.go` → `loadCertificates` |
| File/inline source resolution + validation (upstream of this package) | `pkg/config` → `TLSConfig.ToSSLConfig()` |

## CONVENTIONS

* `IsEnabled()` requires CA + Cert + Key all non-empty; callers gate TLS setup on it before calling
  `New*TLSConf` (see `pkg/rest/server` start paths, redis/mongo/etcd client wiring).
* Validation (cert/key pairing, password-requires-key) lives in `pkg/config` (`TLSConfig.Validate`),
  which runs inside `ToSSLConfig()` before material reaches this package. Do not re-validate here.
* No logging, no global state; errors are returned to callers.
* `Password` decrypts legacy PEM-encrypted private keys via `x509.DecryptPEMBlock` (deprecated upstream);
  the call is kept for existing deployments and must not be extended to new scenarios.

## ANTI-PATTERNS

* Do not read certificate files here — file/inline/base64 source resolution belongs to `pkg/config`.
* Do not add validation duplicating `pkg/config` rules — single validation point avoids rule drift.
* Do not change `IsEnabled()` semantics without checking all gated call sites
  (`pkg/rest/server`, redis/mongo clients, `pkg/discover/etcddiscover`, export server URL scheme).
* Do not add service-specific TLS behavior (e.g. per-service cipher/policy tweaks) — that belongs
  in the consuming service layer.
