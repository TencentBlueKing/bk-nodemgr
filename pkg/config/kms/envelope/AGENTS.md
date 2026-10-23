# KMS ENVELOPE KNOWLEDGE BASE

## OVERVIEW

`pkg/kms/envelope` holds the typed Go structs that mirror the decrypted KMS envelope JSON
payloads. There is exactly one shared **global** envelope (`KMSGlobalEnvelope`) and one
**module** envelope per service (`KMSApplicationEnvelope`, `KMSBackendEnvelope`,
`KMSFileEnvelope`).

Responsibilities:
- Declare the shape of decryptable KMS payloads, nothing else.

Not responsible for:
- Decryption, file/env I/O, validation, or mapping into `pkg/config` structs.

## WHERE TO LOOK

| File | Purpose |
|------|---------|
| `global.go` | `KMSGlobalEnvelope` — shared secrets: bk app id/secret, etcd, redis, mongodb, per-server JWT configs |
| `application.go` | `KMSApplicationEnvelope` — application-module secrets (currently empty placeholder; module envelope is still decrypted at load time) |
| `backend.go` | `KMSBackendEnvelope` — backend-module secrets: encryption key, GSE plugin slot, credit vault |
| `file.go` | `KMSFileEnvelope` — file-module secrets: bkrepo access/secret keys |

## CONVENTIONS

- **`json` tags are snake_case and match the upstream KMS envelope schema verbatim** — they are
  an external contract, not subject to the repo's camelCase YAML convention. Never rename a tag
  without confirming the producer-side schema.
- **Field names align with `pkg/config` terminology** for the same concept (e.g. `AppCode`,
  `AppSecret`, `SymmetricKey`) so mapping code stays mechanical.
- **New secrets are additive only.** Add new fields/types; do not rename or remove existing
  fields — envelopes produced by older KMS payloads must still unmarshal.
- **English comments required on all exported symbols** (project-wide rule).

## ANTI-PATTERNS

- No logic, methods, or validation in this package — pure type declarations.
- Do not declare fields that do not exist in real KMS payloads "for future use"; empty module
  envelopes (see `application.go`) are the sanctioned placeholder pattern instead.
- Do not embed secret values, fixtures, or example payloads in this package or its tests.
