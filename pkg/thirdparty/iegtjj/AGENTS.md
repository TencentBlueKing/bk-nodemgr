# IEGTJJ KNOWLEDGE BASE

## OVERVIEW

`iegtjj` integrates IEGTJJ vault capability by implementing the `CreditHostVault` contract.

## WHERE TO LOOK

- Contract implementation: `pkg/thirdparty/iegtjj/handler.go`, `pkg/thirdparty/iegtjj/iegtjj.go`
- Provider types: `pkg/thirdparty/iegtjj/types.go`
- Crypto/decrypt helpers: `pkg/thirdparty/iegtjj/decryptor.go`

## CONVENTIONS

- Keep integration behind handler interfaces.
- Encapsulate provider-specific payloads and decrypt details inside this package.

## ANTI-PATTERNS

- Do not leak IEGTJJ-specific models to upper business layers.
